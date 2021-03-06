package runner

import (
	"path/filepath"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

func intersectsPlayer(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	x := int(arg[0].(float64))
	y := int(arg[1].(float64))
	z := int(arg[2].(float64))
	name := arg[3].(string)
	shape := shapes.Shapes[shapes.Names[name]]

	app := ctx.App["app"].(*gfx.App)
	runner := app.Game.(*Runner)

	if intersects(x, x+int(shape.Size[0]), app.Loader.X, app.Loader.X+int(runner.player.Size[0])) &&
		intersects(y, y+int(shape.Size[1]), app.Loader.Y, app.Loader.Y+int(runner.player.Size[1])) &&
		intersects(z, z+int(shape.Size[2]), runner.GetZ(), runner.GetZ()+int(runner.player.Size[2])) {
		return true, nil
	}
	return false, nil
}

func intersects(start, end, start2, end2 int) bool {
	return (end2 <= start || start2 >= end) == false
}

func setMaxZ(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	if f, ok := arg[0].(float64); ok {
		app := ctx.App["app"].(*gfx.App)
		app.View.SetMaxZ(int(f))
	}
	return nil, nil
}

func registerKey(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	if f, ok := arg[0].(float64); ok {
		app := ctx.App["app"].(*gfx.App)
		runner := app.Game.(*Runner)
		runner.RegisterKey(glfw.Key(f))
	}
	return nil, nil
}

func registerShape(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	if s, ok := arg[0].(string); ok {
		app := ctx.App["app"].(*gfx.App)
		runner := app.Game.(*Runner)
		runner.RegisterShapeCallback(s)
	}
	return nil, nil
}

func InitScript(app *gfx.App) *bscript.Context {
	bscript.AddBuiltin("registerShape", registerShape)
	bscript.AddBuiltin("registerKey", registerKey)
	bscript.AddBuiltin("intersectsPlayer", intersectsPlayer)
	bscript.AddBuiltin("setMaxZ", setMaxZ)

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
