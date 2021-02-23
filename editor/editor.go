package editor

import (
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"strings"

	"github.com/uzudil/bscript/bscript"
	"github.com/uzudil/isongn/world"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

type Editor struct {
	app                 *gfx.App
	shapeSelectorIndex  int
	shapeSelectorUpdate bool
	Z                   int
	ctx                 *bscript.Context
	command             *bscript.Command
}

func NewEditor() *Editor {
	return &Editor{
		shapeSelectorUpdate: true,
	}
}

func (e *Editor) Init(app *gfx.App) {
	e.app = app

	// compile the editor script code
	_, ctx, err := bscript.Build(
		filepath.Join(e.app.Config.GameDir, "src", "editor.b"),
		false,
		map[string]interface{}{
			"app":    app,
			"editor": e,
		},
	)
	if err != nil {
		panic(err)
	}
	e.ctx = ctx

	// create the editor bscript calls
	e.command = &bscript.Command{}
	err = bscript.CommandParser.ParseString("editorCommand();", e.command)
	if err != nil {
		panic(err)
	}

	// add a ui
	e.app.Ui.Add(int(e.app.Width)-150, 0, 150, int(e.app.Height), e.shapeSelectorContents)
}

func (e *Editor) Name() string {
	return "editor"
}

var moveKeys map[glfw.Key][2]int = map[glfw.Key][2]int{
	glfw.KeyA:     {1, 0},
	glfw.KeyLeft:  {1, 0},
	glfw.KeyD:     {-1, 0},
	glfw.KeyRight: {-1, 0},
	glfw.KeyW:     {0, -1},
	glfw.KeyUp:    {0, -1},
	glfw.KeyS:     {0, 1},
	glfw.KeyDown:  {0, 1},
}

func (e *Editor) isMoveKey() (int, int, bool) {
	for key, delta := range moveKeys {
		f := e.app.IsFirstDown(key)
		if f && e.app.IsDownMod(key, glfw.ModAlt) {
			return delta[0], delta[1], true
		}
		if f || e.app.IsDownMod(key, glfw.ModShift) {
			return delta[0], delta[1], false
		}
	}
	return 0, 0, false
}

func (e *Editor) Events() {
	if e.app.IsDownAlt1(glfw.KeyLeftBracket) && e.shapeSelectorIndex > 0 {
		for i := e.shapeSelectorIndex - 1; i >= 0; i-- {
			if shapes.Shapes[i] != nil && shapes.Shapes[i].EditorVisible {
				e.shapeSelectorIndex = i
				e.shapeSelectorUpdate = true
				break
			}
		}
	}
	if e.app.IsDownAlt1(glfw.KeyRightBracket) && e.shapeSelectorIndex < len(shapes.Shapes)-1 {
		for i := e.shapeSelectorIndex + 1; i < len(shapes.Shapes); i++ {
			if shapes.Shapes[i] != nil && shapes.Shapes[i].EditorVisible {
				e.shapeSelectorIndex = i
				e.shapeSelectorUpdate = true
				break
			}
		}
	}

	shape := shapes.Shapes[e.shapeSelectorIndex]
	dx, dy, insertMode := e.isMoveKey()
	if insertMode {
		dx *= int(shape.Size[0])
		dy *= int(shape.Size[1])
	}
	e.app.Loader.MoveTo(e.app.Loader.X+dx, e.app.Loader.Y+dy)

	f := e.app.IsFirstDown(glfw.KeySpace)
	ff := e.app.IsDownMod(glfw.KeySpace, glfw.ModShift)
	if insertMode || f || ff {
		e.Z = e.findTop()
		e.setShape(e.app.Loader.X, e.app.Loader.Y, e.Z, shapes.Shapes[e.shapeSelectorIndex], ff)
	}
	if e.app.IsFirstDown(glfw.KeyE) && e.Z > 0 {
		shapes.Shapes[e.shapeSelectorIndex].Traverse(func(xx, yy, zz int) {
			if zz == 0 {
				e.app.Loader.EraseShape(e.app.Loader.X+xx, e.app.Loader.Y+yy, e.Z-1)
			}
		})
	}
	if e.app.IsFirstDown(glfw.KeyX) {
		e.app.Loader.SaveAll()
	}
	if e.app.IsFirstDown(glfw.KeyF) {
		e.fill()
	}

	// call bscript
	e.command.Evaluate(e.ctx)

	if e.shapeSelectorUpdate || e.app.Reload {
		e.Z = e.findTop()
		e.app.View.SetCursor(e.shapeSelectorIndex, e.Z)
		e.app.Invalidate()
	}
}

func (e *Editor) GetZ() int {
	return e.Z
}

func (e *Editor) fill() {
	shape := shapes.Shapes[e.shapeSelectorIndex]
	if strings.HasPrefix(shape.Name, "ground.") {
		shapeIndex, _, _, _, found := e.app.Loader.GetShape(e.app.Loader.X, e.app.Loader.Y, 0)
		var replaceShape *shapes.Shape
		if found {
			replaceShape = shapes.Shapes[shapeIndex]
		}
		e.fillAt(e.app.Loader.X, e.app.Loader.Y, shape, replaceShape)
	}
}

func (e *Editor) fillAt(x, y int, shape, replaceShape *shapes.Shape) {
	shapeIndex, _, _, _, found := e.app.Loader.GetShape(x, y, 0)
	if (replaceShape == nil && found == false) || (replaceShape != nil && int(shapeIndex) == replaceShape.Index) {
		e.setShape(x, y, 0, shape, false)
		w := int(shape.Size[0])
		h := int(shape.Size[1])
		e.fillAt(x-w, y, shape, replaceShape)
		e.fillAt(x+w, y, shape, replaceShape)
		e.fillAt(x, y-h, shape, replaceShape)
		e.fillAt(x, y+h, shape, replaceShape)
	}
}

func (e *Editor) setShape(x, y, z int, shape *shapes.Shape, skipEdge bool) {
	if strings.HasPrefix(shape.Name, "ground.") {
		w := int(shape.Size[0])
		h := int(shape.Size[1])
		x = (x / w) * w
		y = (y / h) * h
		z = 0
	}
	e.app.Loader.SetShape(x, y, z, e.shapeSelectorIndex)

	if skipEdge {
		e.app.Loader.ClearEdge(x, y)
	} else {
		if z == 0 {
			for xx := -1; xx <= 1; xx++ {
				for yy := -1; yy <= 1; yy++ {
					sx := x + xx*int(shape.Size[0])
					sy := y + yy*int(shape.Size[1])
					shapeIndex, _, _, _, found := e.app.Loader.GetShape(sx, sy, z)
					if found {
						e.setEdges(sx, sy, shapes.Shapes[shapeIndex])
					}
				}
			}
		}
	}
}

func (e *Editor) setEdges(x, y int, shape *shapes.Shape) {
	e.app.Loader.ClearEdge(x, y)

	w := int(shape.Size[0])
	h := int(shape.Size[1])

	shapeN := e.getEdgeShape(x, y-h, shape)
	shapeS := e.getEdgeShape(x, y+h, shape)
	shapeE := e.getEdgeShape(x-w, y, shape)
	shapeW := e.getEdgeShape(x+w, y, shape)

	var edgeShape *shapes.Shape
	var edgeName string = ""
	if shapeN != nil && shapeS != nil && shapeE != nil && shapeW != nil {
		edgeShape = shapeN
		edgeName = "nsew"
	} else if shapeN != nil && shapeS != nil && shapeE != nil {
		edgeShape = shapeN
		edgeName = "nse"
	} else if shapeN != nil && shapeS != nil && shapeW != nil {
		edgeShape = shapeN
		edgeName = "nsw"
	} else if shapeE != nil && shapeS != nil && shapeW != nil {
		edgeShape = shapeS
		edgeName = "sew"
	} else if shapeE != nil && shapeN != nil && shapeW != nil {
		edgeShape = shapeN
		edgeName = "new"
	} else if shapeE != nil && shapeW != nil {
		edgeShape = shapeE
		edgeName = "ew"
	} else if shapeN != nil && shapeS != nil {
		edgeShape = shapeN
		edgeName = "ns"
	} else if shapeN != nil && shapeE != nil {
		edgeShape = shapeN
		edgeName = "ne"
	} else if shapeN != nil && shapeW != nil {
		edgeShape = shapeN
		edgeName = "nw"
	} else if shapeS != nil && shapeE != nil {
		edgeShape = shapeS
		edgeName = "se"
	} else if shapeS != nil && shapeW != nil {
		edgeShape = shapeS
		edgeName = "sw"
	} else if shapeN != nil {
		edgeShape = shapeN
		edgeName = "n"
	} else if shapeS != nil {
		edgeShape = shapeS
		edgeName = "s"
	} else if shapeE != nil {
		edgeShape = shapeE
		edgeName = "e"
	} else if shapeW != nil {
		edgeShape = shapeW
		edgeName = "w"
	}

	if edgeName != "" && edgeShape.Index != shape.Index {
		edge := edgeShape.GetEdge(shape.Name, edgeName)
		if edge != nil {
			e.app.Loader.SetEdge(x, y, edge.Index)
		}
	}
}

func (e *Editor) getEdgeShape(x, y int, target *shapes.Shape) *shapes.Shape {
	shapeIndex, _, _, _, found := e.app.Loader.GetShape(x, y, 0)
	if found == false {
		return nil
	}
	shape := shapes.Shapes[shapeIndex]
	if shape.HasEdges(target.Name) {
		return shape
	}
	return nil
}

func (e *Editor) findTop() int {
	lastZ := e.Z
	shape := shapes.Shapes[e.shapeSelectorIndex]
	for x := 0; x < int(shape.Size[0]); x++ {
		for y := 0; y < int(shape.Size[1]); y++ {
			for z := world.SECTION_Z_SIZE - 1; z >= 0; z-- {
				shapeIndex, _, _, _, found := e.app.Loader.GetShape(e.app.Loader.X+x, e.app.Loader.Y+y, z)
				if found {
					shape := shapes.Shapes[shapeIndex]
					if shape.Size[2] == 0 {
						return z
					}
					return lastZ
				}
				lastZ = z
			}
		}
	}
	return 0
}

func (e *Editor) shapeSelectorContents(panel *gfx.Panel) bool {
	if e.shapeSelectorUpdate {
		panel.Clear()
		y := 0
		for i := e.shapeSelectorIndex; i < len(shapes.Shapes) && y < panel.H; i++ {
			shape := shapes.Shapes[i]
			if shape != nil && shape.EditorVisible {
				shapeW := shape.Image.Bounds().Dx()
				shapeH := shape.Image.Bounds().Dy()
				if i == e.shapeSelectorIndex {
					draw.Draw(panel.Rgba, image.Rect(0, y, 150, y+shapeH), &image.Uniform{color.RGBA{0xff, 0xa0, 0, 0xff}}, image.ZP, draw.Src)
				}
				draw.Draw(panel.Rgba, image.Rect(0, y, shapeW, y+shapeH), shape.Image, image.Point{0, 0}, draw.Over)
				y += shape.Image.Bounds().Dy()
			}
		}
		e.shapeSelectorUpdate = false
		return true
	}
	return false
}
