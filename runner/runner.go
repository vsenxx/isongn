package runner

import (
	"image/color"
	"path/filepath"

	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/world"
)

type Message struct {
	x, y    int
	message string
}

type Runner struct {
	app                *gfx.App
	ctx                *bscript.Context
	eventsCall         *bscript.Variable
	deltaArg           *bscript.Value
	fadeDirArg         *bscript.Value
	sectionLoadCall    *bscript.Variable
	sectionLoadXArg    *bscript.Value
	sectionLoadYArg    *bscript.Value
	sectionLoadDataArg *bscript.Value
	sectionSaveCall    *bscript.Variable
	sectionSaveXArg    *bscript.Value
	sectionSaveYArg    *bscript.Value
	messages           []*Message
}

func NewRunner() *Runner {
	return &Runner{}
}

func (runner *Runner) Init(app *gfx.App, config map[string]interface{}) {
	runner.app = app
	runner.app.Loader.SetIoMode(world.RUNNER_MODE)

	runner.app.Ui.AddBg(0, 0, int(runner.app.Width), int(runner.app.Height), color.Transparent, runner.overlayContents)

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

	runner.ctx = ctx

	runner.deltaArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.fadeDirArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.eventsCall = gfx.NewFunctionCall("events", runner.deltaArg, runner.fadeDirArg)

	runner.sectionLoadXArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.sectionLoadYArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.sectionLoadDataArg = &bscript.Value{}
	runner.sectionLoadCall = gfx.NewFunctionCall("onSectionLoad", runner.sectionLoadXArg, runner.sectionLoadYArg, runner.sectionLoadDataArg)

	runner.sectionSaveXArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.sectionSaveYArg = &bscript.Value{Number: &bscript.SignedNumber{}}
	runner.sectionSaveCall = gfx.NewFunctionCall("beforeSectionSave", runner.sectionSaveXArg, runner.sectionSaveYArg)

	// run the main method
	_, err = ast.Evaluate(ctx)
	if err != nil {
		panic(err)
	}
}

func (runner *Runner) Name() string {
	return "runner"
}

func (runner *Runner) Events(delta float64, fadeDir int) {
	runner.deltaArg.Number.Number = delta
	runner.fadeDirArg.Number.Number = float64(fadeDir)
	runner.eventsCall.Evaluate(runner.ctx)
}

func (runner *Runner) GetZ() int {
	return 0
}

func (runner *Runner) SectionLoad(x, y int, data map[string]interface{}) {
	runner.sectionLoadXArg.Number.Number = float64(x)
	runner.sectionLoadYArg.Number.Number = float64(y)
	runner.sectionLoadDataArg.Map = gfx.ToBscriptMap(data)
	runner.sectionLoadCall.Evaluate(runner.ctx)
}

func (runner *Runner) SectionSave(x, y int) map[string]interface{} {
	runner.sectionSaveXArg.Number.Number = float64(x)
	runner.sectionSaveYArg.Number.Number = float64(y)
	ret, _ := runner.sectionSaveCall.Evaluate(runner.ctx)
	return ret.(map[string]interface{})
}

func (runner *Runner) overlayContents(panel *gfx.Panel) bool {
	if len(runner.messages) > 0 {
		for _, msg := range runner.messages {
			runner.app.Font.Printf(panel.Rgba, color.White, msg.x, msg.y, msg.message)
		}

		// clear the messages (but keep memory)
		runner.messages = runner.messages[:]

		return true
	}
	return false
}

func (runner *Runner) PrintMessage(x, y int, message string) {
	runner.messages = append(runner.messages, &Message{x, y, message})
}
