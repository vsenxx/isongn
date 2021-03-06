package runner

import (
	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
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

	// register some runner-specific callbacks and init bscript
	runner.ctx = InitScript(app)

	runner.deltaArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.eventsCall = gfx.NewFunctionCall("events", runner.deltaArg)
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
