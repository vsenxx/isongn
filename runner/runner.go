package runner

import (
	"github.com/go-gl/glfw/v3.3/glfw"
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
}

func NewRunner() *Runner {
	return &Runner{
		z: 1,
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
		newZ := runner.app.View.FindTop(newX, newY, runner.player)
		if newZ <= runner.z+1 {
			runner.app.Loader.MoveTo(newX, newY)
			runner.app.View.Load()
			runner.z = newZ
		} else {
			// player is blocked
			moved = false
		}
		runner.app.View.SetShape(runner.app.Loader.X, runner.app.Loader.Y, runner.z, runner.player.Index)
	}
	return moved, newXf, newYf, newX, newY
}

func (runner *Runner) GetZ() int {
	return runner.z
}
