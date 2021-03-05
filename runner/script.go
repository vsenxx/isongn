package runner

import (
	"path/filepath"

	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/gfx"
)

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

func NewFunctionCall(functionName string, values ...*bscript.Value) *bscript.Variable {
	args := make([]*bscript.Expression, len(values))
	for i, v := range values {
		args[i] = &bscript.Expression{
			BoolTerm: &bscript.BoolTerm{
				Left: &bscript.Cmp{
					Left: &bscript.Term{
						Left: &bscript.Factor{
							Base: v,
						},
					},
				},
			},
		}
	}

	return &bscript.Variable{
		Variable: functionName,
		Suffixes: []*bscript.VariableSuffix{
			{
				CallParams: &bscript.CallParams{
					Args: args,
				},
			},
		},
	}
}
