package gfx

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/shapes"
)

func print(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	fmt.Println(arg[0].(string))
	return nil, nil
}

func eraseShape(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	x := int(arg[0].(float64))
	y := int(arg[1].(float64))
	z := int(arg[2].(float64))
	app := ctx.App["app"].(*App)
	app.View.EraseShape(x, y, z)
	return nil, nil
}

func setShape(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	x := int(arg[0].(float64))
	y := int(arg[1].(float64))
	z := int(arg[2].(float64))
	name := arg[3].(string)
	app := ctx.App["app"].(*App)
	app.View.SetShape(x, y, z, shapes.Names[name])
	return nil, nil
}

func getShape(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	x := int(arg[0].(float64))
	y := int(arg[1].(float64))
	z := int(arg[2].(float64))
	app := ctx.App["app"].(*App)
	if shapeIndex, ox, oy, oz, found := app.View.GetShape(x, y, z); found {
		r := make([]interface{}, 4)
		r[0] = shapes.Shapes[shapeIndex].Name
		r[1] = float64(ox)
		r[2] = float64(oy)
		r[3] = float64(oz)
		return &r, nil
	}
	return nil, nil
}

func getPosition(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	r := make([]interface{}, 3)
	app := ctx.App["app"].(*App)
	r[0] = float64(app.Loader.X)
	r[1] = float64(app.Loader.Y)
	r[2] = float64(app.Game.GetZ())
	return &r, nil
}

func isPressed(ctx *bscript.Context, arg ...interface{}) (interface{}, error) {
	a, ok := arg[0].(float64)
	if ok {
		key := glfw.Key(int(a))
		app := ctx.App["app"].(*App)
		return app.IsFirstDown(key), nil
	}
	return nil, fmt.Errorf("%s unable to parse key at", ctx.Pos)
}

var constants map[string]interface{} = map[string]interface{}{
	// keyboard keys
	"KeyUnknown":      float64(glfw.KeyUnknown),
	"KeySpace":        float64(glfw.KeySpace),
	"KeyApostrophe":   float64(glfw.KeyApostrophe),
	"KeyComma":        float64(glfw.KeyComma),
	"KeyMinus":        float64(glfw.KeyMinus),
	"KeyPeriod":       float64(glfw.KeyPeriod),
	"KeySlash":        float64(glfw.KeySlash),
	"Key0":            float64(glfw.Key0),
	"Key1":            float64(glfw.Key1),
	"Key2":            float64(glfw.Key2),
	"Key3":            float64(glfw.Key3),
	"Key4":            float64(glfw.Key4),
	"Key5":            float64(glfw.Key5),
	"Key6":            float64(glfw.Key6),
	"Key7":            float64(glfw.Key7),
	"Key8":            float64(glfw.Key8),
	"Key9":            float64(glfw.Key9),
	"KeySemicolon":    float64(glfw.KeySemicolon),
	"KeyEqual":        float64(glfw.KeyEqual),
	"KeyA":            float64(glfw.KeyA),
	"KeyB":            float64(glfw.KeyB),
	"KeyC":            float64(glfw.KeyC),
	"KeyD":            float64(glfw.KeyD),
	"KeyE":            float64(glfw.KeyE),
	"KeyF":            float64(glfw.KeyF),
	"KeyG":            float64(glfw.KeyG),
	"KeyH":            float64(glfw.KeyH),
	"KeyI":            float64(glfw.KeyI),
	"KeyJ":            float64(glfw.KeyJ),
	"KeyK":            float64(glfw.KeyK),
	"KeyL":            float64(glfw.KeyL),
	"KeyM":            float64(glfw.KeyM),
	"KeyN":            float64(glfw.KeyN),
	"KeyO":            float64(glfw.KeyO),
	"KeyP":            float64(glfw.KeyP),
	"KeyQ":            float64(glfw.KeyQ),
	"KeyR":            float64(glfw.KeyR),
	"KeyS":            float64(glfw.KeyS),
	"KeyT":            float64(glfw.KeyT),
	"KeyU":            float64(glfw.KeyU),
	"KeyV":            float64(glfw.KeyV),
	"KeyW":            float64(glfw.KeyW),
	"KeyX":            float64(glfw.KeyX),
	"KeyY":            float64(glfw.KeyY),
	"KeyZ":            float64(glfw.KeyZ),
	"KeyLeftBracket":  float64(glfw.KeyLeftBracket),
	"KeyBackslash":    float64(glfw.KeyBackslash),
	"KeyRightBracket": float64(glfw.KeyRightBracket),
	"KeyGraveAccent":  float64(glfw.KeyGraveAccent),
	"KeyWorld1":       float64(glfw.KeyWorld1),
	"KeyWorld2":       float64(glfw.KeyWorld2),
	"KeyEscape":       float64(glfw.KeyEscape),
	"KeyEnter":        float64(glfw.KeyEnter),
	"KeyTab":          float64(glfw.KeyTab),
	"KeyBackspace":    float64(glfw.KeyBackspace),
	"KeyInsert":       float64(glfw.KeyInsert),
	"KeyDelete":       float64(glfw.KeyDelete),
	"KeyRight":        float64(glfw.KeyRight),
	"KeyLeft":         float64(glfw.KeyLeft),
	"KeyDown":         float64(glfw.KeyDown),
	"KeyUp":           float64(glfw.KeyUp),
	"KeyPageUp":       float64(glfw.KeyPageUp),
	"KeyPageDown":     float64(glfw.KeyPageDown),
	"KeyHome":         float64(glfw.KeyHome),
	"KeyEnd":          float64(glfw.KeyEnd),
	"KeyCapsLock":     float64(glfw.KeyCapsLock),
	"KeyScrollLock":   float64(glfw.KeyScrollLock),
	"KeyNumLock":      float64(glfw.KeyNumLock),
	"KeyPrintScreen":  float64(glfw.KeyPrintScreen),
	"KeyPause":        float64(glfw.KeyPause),
	"KeyF1":           float64(glfw.KeyF1),
	"KeyF2":           float64(glfw.KeyF2),
	"KeyF3":           float64(glfw.KeyF3),
	"KeyF4":           float64(glfw.KeyF4),
	"KeyF5":           float64(glfw.KeyF5),
	"KeyF6":           float64(glfw.KeyF6),
	"KeyF7":           float64(glfw.KeyF7),
	"KeyF8":           float64(glfw.KeyF8),
	"KeyF9":           float64(glfw.KeyF9),
	"KeyF10":          float64(glfw.KeyF10),
	"KeyF11":          float64(glfw.KeyF11),
	"KeyF12":          float64(glfw.KeyF12),
	"KeyF13":          float64(glfw.KeyF13),
	"KeyF14":          float64(glfw.KeyF14),
	"KeyF15":          float64(glfw.KeyF15),
	"KeyF16":          float64(glfw.KeyF16),
	"KeyF17":          float64(glfw.KeyF17),
	"KeyF18":          float64(glfw.KeyF18),
	"KeyF19":          float64(glfw.KeyF19),
	"KeyF20":          float64(glfw.KeyF20),
	"KeyF21":          float64(glfw.KeyF21),
	"KeyF22":          float64(glfw.KeyF22),
	"KeyF23":          float64(glfw.KeyF23),
	"KeyF24":          float64(glfw.KeyF24),
	"KeyF25":          float64(glfw.KeyF25),
	"KeyKP0":          float64(glfw.KeyKP0),
	"KeyKP1":          float64(glfw.KeyKP1),
	"KeyKP2":          float64(glfw.KeyKP2),
	"KeyKP3":          float64(glfw.KeyKP3),
	"KeyKP4":          float64(glfw.KeyKP4),
	"KeyKP5":          float64(glfw.KeyKP5),
	"KeyKP6":          float64(glfw.KeyKP6),
	"KeyKP7":          float64(glfw.KeyKP7),
	"KeyKP8":          float64(glfw.KeyKP8),
	"KeyKP9":          float64(glfw.KeyKP9),
	"KeyKPDecimal":    float64(glfw.KeyKPDecimal),
	"KeyKPDivide":     float64(glfw.KeyKPDivide),
	"KeyKPMultiply":   float64(glfw.KeyKPMultiply),
	"KeyKPSubtract":   float64(glfw.KeyKPSubtract),
	"KeyKPAdd":        float64(glfw.KeyKPAdd),
	"KeyKPEnter":      float64(glfw.KeyKPEnter),
	"KeyKPEqual":      float64(glfw.KeyKPEqual),
	"KeyLeftShift":    float64(glfw.KeyLeftShift),
	"KeyLeftControl":  float64(glfw.KeyLeftControl),
	"KeyLeftAlt":      float64(glfw.KeyLeftAlt),
	"KeyLeftSuper":    float64(glfw.KeyLeftSuper),
	"KeyRightShift":   float64(glfw.KeyRightShift),
	"KeyRightControl": float64(glfw.KeyRightControl),
	"KeyRightAlt":     float64(glfw.KeyRightAlt),
	"KeyRightSuper":   float64(glfw.KeyRightSuper),
	"KeyMenu":         float64(glfw.KeyMenu),
	"KeyLast":         float64(glfw.KeyLast),
}

func InitScript() {
	bscript.AddBuiltin("isPressed", isPressed)
	bscript.AddBuiltin("getPosition", getPosition)
	bscript.AddBuiltin("eraseShape", eraseShape)
	bscript.AddBuiltin("setShape", setShape)
	bscript.AddBuiltin("getShape", getShape)
	bscript.AddBuiltin("print", print)
	for k, v := range constants {
		bscript.AddConstant(k, v)
	}
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
