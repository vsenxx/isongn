package editor

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

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
}

func NewEditor() *Editor {
	return &Editor{
		shapeSelectorUpdate: true,
	}
}

func (e *Editor) Init(app *gfx.App) {
	e.app = app
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
		e.shapeSelectorIndex--
		e.shapeSelectorUpdate = true
	}
	if e.app.IsDownAlt1(glfw.KeyRightBracket) && e.shapeSelectorIndex < len(shapes.Shapes)-1 {
		e.shapeSelectorIndex++
		e.shapeSelectorUpdate = true
	}

	shape := shapes.Shapes[e.shapeSelectorIndex]
	dx, dy, insertMode := e.isMoveKey()
	if insertMode {
		dx *= int(shape.Size[0])
		dy *= int(shape.Size[1])
	}
	e.app.Loader.MoveTo(e.app.Loader.X+dx, e.app.Loader.Y+dy)

	if insertMode || e.app.IsFirstDown(glfw.KeySpace) {
		e.Z = e.findTop()
		e.setShape()
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

	if e.shapeSelectorUpdate || e.app.Reload {
		e.Z = e.findTop()
		e.app.View.SetCursor(e.shapeSelectorIndex, e.Z)
		e.app.Invalidate()
	}
}

func (e *Editor) setShape() {
	x := e.app.Loader.X
	y := e.app.Loader.Y
	z := e.Z
	shape := shapes.Shapes[e.shapeSelectorIndex]
	if strings.HasPrefix(shape.Name, "ground.") {
		x = (x / 4) * 4
		y = (y / 4) * 4
		z = 0
	}
	e.app.Loader.SetShape(x, y, z, byte(e.shapeSelectorIndex))

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

func (e *Editor) setEdges(x, y int, shape *shapes.Shape) {
	e.app.Loader.ClearEdge(x, y)

	w := int(shape.Size[0])
	h := int(shape.Size[1])

	shapeN := e.getEdgeShape(x, y-h)
	shapeS := e.getEdgeShape(x, y+h)
	shapeE := e.getEdgeShape(x-w, y)
	shapeW := e.getEdgeShape(x+w, y)

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
			e.app.Loader.SetEdge(x, y, byte(edge.Index))
		}
	}
}

func (e *Editor) getEdgeShape(x, y int) *shapes.Shape {
	shapeIndex, _, _, _, found := e.app.Loader.GetShape(x, y, 0)
	if found == false {
		return nil
	}
	shape := shapes.Shapes[shapeIndex]
	if shape.HasEdges == false {
		return nil
	}
	return shape
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
			shapeW := shape.Image.Bounds().Dx()
			shapeH := shape.Image.Bounds().Dy()
			if i == e.shapeSelectorIndex {
				draw.Draw(panel.Rgba, image.Rect(0, y, 150, y+shapeH), &image.Uniform{color.RGBA{0xff, 0xa0, 0, 0xff}}, image.ZP, draw.Src)
			}
			draw.Draw(panel.Rgba, image.Rect(0, y, shapeW, y+shapeH), shape.Image, image.Point{0, 0}, draw.Over)
			y += shape.Image.Bounds().Dy()
		}
		e.shapeSelectorUpdate = false
		return true
	}
	return false
}
