package runner

import (
	"path/filepath"

	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

func intersectsShapes(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	x := int(arg[0].(float64))
	y := int(arg[1].(float64))
	z := int(arg[2].(float64))
	nameA := arg[3].(string)
	nameB := arg[4].(string)
	shapeA := shapes.Shapes[shapes.Names[nameA]]
	shapeB := shapes.Shapes[shapes.Names[nameB]]

	app := ctx.App["app"].(*gfx.App)
	runner := app.Game.(*Runner)

	if intersects(x, x+int(shapeA.Size[0]), app.Loader.X, app.Loader.X+int(shapeB.Size[0])) &&
		intersects(y, y+int(shapeA.Size[1]), app.Loader.Y, app.Loader.Y+int(shapeB.Size[1])) &&
		intersects(z, z+int(shapeA.Size[2]), runner.GetZ(), runner.GetZ()+int(shapeB.Size[2])) {
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

func InitScript(app *gfx.App) *bscript.Context {
	bscript.AddBuiltin("intersectsShapes", intersectsShapes)
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
