package runner

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

type Runner struct {
	app    *gfx.App
	player *shapes.Shape
	z      int
	dir    *shapes.Direction
}

func NewRunner() *Runner {
	return &Runner{
		z: 1,
	}
}

func (runner *Runner) Init(app *gfx.App) {
	runner.app = app
	runner.player = shapes.Shapes[shapes.Names["cow"]]
	// runner.z = 1

	// init the player
	runner.app.Loader.MoveTo(976, 1028)
	runner.app.View.Load()
	runner.app.View.SetShape(runner.app.Loader.X, runner.app.Loader.Y, runner.z, runner.player.Index)
}

func (runner *Runner) Name() string {
	return "runner"
}

func (runner *Runner) Events() {
	ox := runner.app.Loader.X
	oy := runner.app.Loader.Y
	if runner.app.IsDownAlt(glfw.KeyA, glfw.KeyLeft) && runner.app.Loader.X > 0 {
		ox++
	}
	if runner.app.IsDownAlt(glfw.KeyD, glfw.KeyRight) {
		ox--
	}
	if runner.app.IsDownAlt(glfw.KeyW, glfw.KeyUp) && runner.app.Loader.Y > 0 {
		oy--
	}
	if runner.app.IsDownAlt(glfw.KeyS, glfw.KeyDown) {
		oy++
	}

	if ox != runner.app.Loader.X || oy != runner.app.Loader.Y {
		runner.playerMove(ox, oy)
	}
}

func (runner *Runner) playerMove(newX, newY int) {
	oldX := runner.app.Loader.X
	oldY := runner.app.Loader.Y
	oldZ := runner.z
	runner.dir = shapes.GetDir(oldX, oldY, newX, newY)
	runner.app.View.EraseShape(oldX, oldY, oldZ)
	newZ := runner.app.View.FindTop(newX, newY, runner.player)
	if newZ <= runner.z+1 {
		runner.app.Loader.MoveTo(newX, newY)
		runner.app.View.Load()
		runner.z = newZ
		runner.app.View.SetShapeDir(newX, newY, newZ, runner.player.Index, runner.dir)
	} else {
		runner.app.View.SetShapeDir(oldX, oldY, oldZ, runner.player.Index, runner.dir)
	}
}

func (runner *Runner) GetZ() int {
	return runner.z
}
