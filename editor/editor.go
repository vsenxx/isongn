package editor

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/uzudil/isongn/world"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/gfx"
	"github.com/uzudil/isongn/shapes"
)

type Editor struct {
	app                 *gfx.App
	shapeSelectorIndex  int
	shapeSelectorUpdate bool
	reload              bool
	Z                   int
}

func NewEditor(app *gfx.App) *Editor {
	e := &Editor{
		app:                 app,
		shapeSelectorUpdate: true,
		reload:              true,
	}

	// add a ui
	app.Ui.Add(650, 0, 150, 600, e.shapeSelectorContents)

	return e
}

func (e *Editor) Events() {
	if (e.app.IsFirstDown(glfw.KeyUp) || e.app.IsDownMod(glfw.KeyUp, glfw.ModShift)) && e.shapeSelectorIndex > 0 {
		e.shapeSelectorIndex--
		e.shapeSelectorUpdate = true
	}
	if (e.app.IsFirstDown(glfw.KeyDown) || e.app.IsDownMod(glfw.KeyDown, glfw.ModShift)) && e.shapeSelectorIndex < len(shapes.Shapes)-1 {
		e.shapeSelectorIndex++
		e.shapeSelectorUpdate = true
	}
	if (e.app.IsFirstDown(glfw.KeyA) || e.app.IsDownMod(glfw.KeyA, glfw.ModShift)) && e.app.Loader.X > 0 {
		e.app.Loader.X++
		e.reload = true
	}
	if e.app.IsFirstDown(glfw.KeyD) || e.app.IsDownMod(glfw.KeyD, glfw.ModShift) {
		e.app.Loader.X--
		e.reload = true
	}
	if (e.app.IsFirstDown(glfw.KeyW) || e.app.IsDownMod(glfw.KeyW, glfw.ModShift)) && e.app.Loader.Y > 0 {
		e.app.Loader.Y--
		e.reload = true
	}
	if e.app.IsFirstDown(glfw.KeyS) || e.app.IsDownMod(glfw.KeyS, glfw.ModShift) {
		e.app.Loader.Y++
		e.reload = true
	}
	if e.app.IsFirstDown(glfw.KeySpace) {
		e.app.Loader.SetShape(e.app.Loader.X, e.app.Loader.Y, e.Z, byte(e.shapeSelectorIndex))
		e.reload = true
	}
	if e.app.IsFirstDown(glfw.KeyX) {
		e.app.Loader.SaveAll()
	}

	if e.shapeSelectorUpdate || e.reload {
		e.Z = e.findTop(e.shapeSelectorIndex)
		e.app.View.SetCursor(e.shapeSelectorIndex, e.Z)
		e.reload = true
	}
}

func (e *Editor) findTop(shapeIndex int) int {
	lastZ := e.Z
	shape := shapes.Shapes[shapeIndex]
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

func (e *Editor) Draw() {
	if e.reload {
		e.app.View.Load(e.app.Loader)
		e.reload = false
	}
	e.app.View.Draw()
	e.app.Ui.Draw()
}
