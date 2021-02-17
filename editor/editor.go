package editor

import (
	"fmt"
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

var edgeDefs [4][2]int = [4][2]int{{0, 1}, {1, 0}, {-1, 0}, {0, -1}}

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

	for xx := -1; xx <= 1; xx++ {
		for yy := -1; yy <= 1; yy++ {
			sx := x + xx*int(shape.Size[0])
			sy := y + yy*int(shape.Size[1])
			shapeIndex, _, _, _, found := e.app.Loader.GetShape(sx, sy, z)
			if found {
				e.setEdges(sx, sy, z, shapes.Shapes[shapeIndex])
			}
		}
	}
}

func (e *Editor) setEdges(x, y, z int, shape *shapes.Shape) {
	for edgeIndex, edgeDef := range edgeDefs {
		ex := x + edgeDef[0]*int(shape.Size[0])
		ey := y + edgeDef[1]*int(shape.Size[1])
		shapeIndex, _, _, _, found := e.app.Loader.GetShape(ex, ey, z)
		if found && int(shapeIndex) == shape.Index {
			fmt.Printf("clearing: %s %d\n", shape.Name, edgeIndex)
			e.app.Loader.ClearEdge(ex, ey, z, edgeIndex)
		} else if shape.HasEdges {
			toShapeName := ""
			if found {
				toShapeName = shapes.Shapes[shapeIndex].Name
			}
			e.app.Loader.SetEdge(ex, ey, z, edgeIndex, shape.EdgeShapeIndex(toShapeName, edgeIndex))
		}
	}
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
