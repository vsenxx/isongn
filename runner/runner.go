package runner

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

type Runner struct {
	app            *gfx.App
	player         *shapes.Shape
	z              int
	dir            shapes.Direction
	speed          float64
	lastDx, lastDy int
	shapeCallbacks map[string]bool
	keyCallbacks   map[glfw.Key]bool
	ctx            *bscript.Context
	onShapeCall    *bscript.Variable
	onShapeArg     *bscript.Value
	onKeyCall      *bscript.Variable
	onKeyArg       *bscript.Value
	onMoveXArg     *bscript.Value
	onMoveYArg     *bscript.Value
	onMoveZArg     *bscript.Value
	onMoveCall     *bscript.Variable
}

func NewRunner() *Runner {
	return &Runner{
		z:              1,
		shapeCallbacks: map[string]bool{},
		keyCallbacks:   map[glfw.Key]bool{},
	}
}

func (runner *Runner) Init(app *gfx.App, config map[string]interface{}) {
	runner.app = app

	runner.player = shapes.Shapes[shapes.Names[config["player"].(string)]]
	runner.speed = config["playerSpeed"].(float64)

	// init the player
	start := config["start"].([]interface{})
	runner.app.Loader.MoveTo(int(start[0].(float64)), int(start[1].(float64)))
	runner.app.View.Load()
	runner.app.View.SetShape(runner.app.Loader.X, runner.app.Loader.Y, runner.z, runner.player.Index)

	// register some runner-specific callbacks and init bscript
	runner.ctx = InitScript(app)

	// create a function call + arg for "onShape()"
	runner.onShapeArg = &bscript.Value{}
	runner.onShapeCall = gfx.NewFunctionCall("onShape", runner.onShapeArg)

	runner.onKeyArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.onKeyCall = gfx.NewFunctionCall("onKey", runner.onKeyArg)

	runner.onMoveXArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.onMoveYArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.onMoveZArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.onMoveCall = gfx.NewFunctionCall("onPlayerMove", runner.onMoveXArg, runner.onMoveYArg, runner.onMoveZArg)
}

func (runner *Runner) Name() string {
	return "runner"
}

func (runner *Runner) isDown(key, altKey glfw.Key) bool {
	return runner.app.IsDown(key) || runner.app.IsDown(altKey)
}

func (runner *Runner) Events(delta float64) {
	animationType := shapes.ANIMATION_STAND
	dx := 0
	dy := 0
	if runner.isDown(glfw.KeyA, glfw.KeyLeft) && runner.app.Loader.X > 0 {
		dx++
	}
	if runner.isDown(glfw.KeyD, glfw.KeyRight) {
		dx--
	}
	if runner.isDown(glfw.KeyW, glfw.KeyUp) && runner.app.Loader.Y > 0 {
		dy--
	}
	if runner.isDown(glfw.KeyS, glfw.KeyDown) {
		dy++
	}

	if !(dx == 0 && dy == 0) {
		animationType = shapes.ANIMATION_MOVE
		runner.playerMove(dx, dy, delta)
	}
	runner.app.View.SetShapeAnimation(runner.app.Loader.X, runner.app.Loader.Y, runner.z, animationType, runner.dir)

	// make any registered key callbacks
	for k := range runner.keyCallbacks {
		if runner.app.IsFirstDown(k) {
			runner.onKeyArg.Number.Number = float64(k)
			runner.onKeyCall.Evaluate(runner.ctx)
		}
	}
}

func (runner *Runner) playerMove(dx, dy int, delta float64) {
	// try movement
	moved, newXf, newYf, newX, newY := runner.playerMoveDir(dx, dy, delta)

	// if that fails, try in a recent direction...
	if !moved && runner.lastDx != 0 {
		moved, newXf, newYf, newX, newY = runner.playerMoveDir(runner.lastDx, 0, delta)
	}
	if !moved && runner.lastDy != 0 {
		moved, newXf, newYf, newX, newY = runner.playerMoveDir(0, runner.lastDy, delta)
	}

	// pixel-scroll view
	if moved {
		scrollX := newXf - float32(newX)
		scrollY := newYf - float32(newY)
		runner.app.View.Scroll(scrollX, scrollY)
		runner.app.View.SetOffset(runner.app.Loader.X, runner.app.Loader.Y, runner.z, scrollX, scrollY)
	}
}

func (runner *Runner) playerMoveDir(dx, dy int, delta float64) (bool, float32, float32, int, int) {
	oldX := runner.app.Loader.X
	oldY := runner.app.Loader.Y
	oldZ := runner.z
	runner.dir = shapes.GetDir(dx, dy)
	if dx != 0 {
		runner.lastDx = dx
	}
	if dy != 0 {
		runner.lastDy = dy
	}
	moved := true

	// adjust speed for diagonal movement... maybe this should be computed from iso angles?
	speed := runner.speed
	if dx != 0 && dy != 0 {
		speed *= 1.5
	}
	newXf := float32(runner.app.Loader.X) + runner.app.View.ScrollOffset[0] + float32(float64(dx)*delta/speed)
	newYf := float32(runner.app.Loader.Y) + runner.app.View.ScrollOffset[1] + float32(float64(dy)*delta/speed)
	newX := int(newXf + 0.5)
	newY := int(newYf + 0.5)
	if newX != oldX || newY != oldY {
		runner.app.View.EraseShape(oldX, oldY, oldZ)
		newZ := runner.app.View.FindTopFit(newX, newY, runner.player)
		if newZ <= runner.z+1 && runner.inspectUnder(newX, newY, newZ) {
			runner.app.Loader.MoveTo(newX, newY)
			runner.app.View.Load()
			runner.z = newZ

			// update the script
			runner.onMoveXArg.Number.Number = float64(newX)
			runner.onMoveYArg.Number.Number = float64(newY)
			runner.onMoveZArg.Number.Number = float64(newZ)
			runner.onMoveCall.Evaluate(runner.ctx)
		} else {
			// player is blocked
			moved = false
		}
		runner.app.View.SetShape(runner.app.Loader.X, runner.app.Loader.Y, runner.z, runner.player.Index)
	}
	return moved, newXf, newYf, newX, newY
}

func (runner *Runner) inspectUnder(newX, newY, newZ int) bool {
	res := true
	// are we standing on a shape we're interested in?
	if len(runner.shapeCallbacks) > 0 && runner.app.View.InspectUnder(newX, newY, newZ, runner.player, runner.shapeCallbacks) {
		for name := range runner.shapeCallbacks {
			if runner.shapeCallbacks[name] {
				// call the script
				runner.onShapeArg.String = &name
				result, err := runner.onShapeCall.Evaluate(runner.ctx)
				if err != nil {
					panic(err)
				}

				res = result.(bool)

				// reset the callback
				runner.shapeCallbacks[name] = false
			}
		}
	}
	return res
}

func (runner *Runner) GetZ() int {
	return runner.z
}

func (runner *Runner) RegisterShapeCallback(name string) {
	runner.shapeCallbacks[name] = false
}

func (runner *Runner) RegisterKey(key glfw.Key) {
	runner.keyCallbacks[key] = true
}
