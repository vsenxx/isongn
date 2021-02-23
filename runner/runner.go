package runner

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/gfx"
)

type Runner struct {
	app *gfx.App
}

func NewRunner() *Runner {
	return &Runner{}
}

func (runner *Runner) Init(app *gfx.App) {
	runner.app = app
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
	runner.app.Loader.MoveTo(ox, oy)
}

func (runner *Runner) GetZ() int {
	return 0
}
