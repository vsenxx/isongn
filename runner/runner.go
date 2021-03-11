package runner

import (
	"path/filepath"

	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/world"
)

type Runner struct {
	app        *gfx.App
	ctx        *bscript.Context
	eventsCall *bscript.Variable
	deltaArg   *bscript.Value
}

func NewRunner() *Runner {
	return &Runner{}
}

func (runner *Runner) Init(app *gfx.App, config map[string]interface{}) {
	runner.app = app
	runner.app.Loader.SetIoMode(world.RUNNER_MODE)

	// register some runner-specific callbacks and init bscript
	runner.ctx = initScript(app)

	runner.deltaArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.eventsCall = gfx.NewFunctionCall("events", runner.deltaArg)
}

func initScript(app *gfx.App) *bscript.Context {

	// compile the editor script code
	ast, ctx, err := bscript.Build(
		filepath.Join(app.Config.GameDir, "src", "runner"),
		false,
		map[string]interface{}{
			"app": app,
		},
	)
	if err != nil {
		panic(err)
	}

	// run the main method
	_, err = ast.Evaluate(ctx)
	if err != nil {
		panic(err)
	}

	return ctx
}

func (runner *Runner) Name() string {
	return "runner"
}

func (runner *Runner) Events(delta float64) {
	runner.deltaArg.Number.Number = delta
	runner.eventsCall.Evaluate(runner.ctx)
}

func (runner *Runner) GetZ() int {
	return 0
}
