package shapes

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"

	// import to initialize png decoding
	"image/draw"
	_ "image/png"
)

type Edge struct {
	Shapes []*Shape
}

type ShapeMeta struct {
	DpiMultiplier float32
	UnitPixels    [2]int
}

type Shape struct {
	Index       int
	Name        string
	Image       *image.RGBA
	Size        [3]float32
	PixelOffset [2]float32
	PixelDim    [2]float32
	TexOffset   [2]float32
	TexDim      [2]float32
	Fudge       float32
	ImageIndex  int
	ShapeMeta   *ShapeMeta
	Flags       map[string]bool
	Edges       map[string]map[string][]*Shape
	Offset      [3]float32
}

var Shapes []*Shape
var Names map[string]int = map[string]int{}
var Images []image.Image

func InitShapes(gameDir string) error {
	bytes, err := ioutil.ReadFile(filepath.Join(gameDir, "shapes.json"))
	if err != nil {
		// return nil not error: file missing is not an error
		return err
	}
	data := []map[string]interface{}{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}

	for _, block := range data {
		imgFile := block["image"].(string)
		shapes := block["shapes"].([]interface{})
		fmt.Printf("Processing %s - %d shapes...\n", imgFile, len(shapes))

		// per-image meta data
		grid := block["grid"].(map[string]interface{})
		units := grid["units"].([]interface{})
		dpi := block["dpi"].(float64)
		shapeMeta := &ShapeMeta{
			DpiMultiplier: float32(dpi) / 96.0,
			UnitPixels:    [2]int{int(units[0].(float64)), int(units[1].(float64))},
		}

		img, err := loadImage(filepath.Join(gameDir, "images", imgFile))
		if err != nil {
			return err
		}
		imageIndex := len(Images)
		Images = append(Images, img)
		for _, s := range shapes {
			shapeDef := s.(map[string]interface{})
			name := shapeDef["name"].(string)
			fmt.Printf("\tProcessing %s\n", name)
			Names[name] = len(Shapes)
			appendShape(name, shapeDef, imageIndex, img, shapeMeta)
		}
	}
	fmt.Printf("Loaded %d shapes.\n", len(Shapes))
	return nil
}

func appendShape(name string, shapeDef map[string]interface{}, imageIndex int, img image.Image, shapeMeta *ShapeMeta) {
	// flags
	flagsSet := map[string]bool{}
	flags, ok := shapeDef["flags"]
	if ok {
		for _, flag := range flags.([]interface{}) {
			flagsSet[flag.(string)] = true
		}
	}

	// size
	sizeI := shapeDef["size"].([]interface{})
	size := [3]float32{float32(sizeI[0].(float64)), float32(sizeI[1].(float64)), float32(sizeI[2].(float64))}
	if _, ok := flagsSet["stamp"]; ok {
		size[2] = 0.01
	}

	// pixel bounding box
	posI := shapeDef["pos"].([]interface{})
	px := float32(posI[0].(float64)) * shapeMeta.DpiMultiplier
	py := float32(posI[1].(float64)) * shapeMeta.DpiMultiplier
	unitPixelX := float32(shapeMeta.UnitPixels[0])
	unitPixelY := float32(shapeMeta.UnitPixels[1])
	pw := (size[0] + size[1]) * unitPixelX * shapeMeta.DpiMultiplier
	ph := (size[0] + size[1] + size[2]) * unitPixelY * shapeMeta.DpiMultiplier

	// fudge
	fudge64, ok := shapeDef["fudge"].(float64)
	var fudge float32 = 0.01
	if ok {
		fudge = float32(fudge64)
	}

	// offset
	offset := [3]float32{}
	if offsetI, ok := shapeDef["offset"].([]interface{}); ok {
		offset[0] = float32(offsetI[0].(float64))
		offset[1] = float32(offsetI[1].(float64))
		offset[2] = float32(offsetI[2].(float64))
	}

	shape := newShape(
		len(Shapes),
		name,
		size,
		px, py, pw, ph,
		img,
		fudge,
		imageIndex,
		shapeMeta,
		flagsSet,
		offset,
	)

	// edges
	refName, ok := shapeDef["ref"]
	if ok {
		targetI, ok := shapeDef["target"]
		var target string
		if ok == false {
			target = "default"
		} else {
			target = targetI.(string)
		}

		parts := strings.Split(name, ".")
		ref := findShape(refName.(string))
		if _, ok := ref.Edges[target]; ok == false {
			ref.Edges[target] = map[string][]*Shape{}
		}
		if _, ok := ref.Edges[target][parts[2]]; ok {
			ref.Edges[target][parts[2]] = append(ref.Edges[target][parts[2]], shape)
		} else {
			ref.Edges[target][parts[2]] = []*Shape{shape}
		}
	}

	Shapes = append(Shapes, shape)
}

func (shape *Shape) HasEdges(shapeName string) bool {
	_, ok := shape.Edges[shapeName]
	if ok {
		return true
	}
	_, ok = shape.Edges["default"]
	if ok {
		return true
	}
	return false
}

func (shape *Shape) GetEdge(shapeName, edgeName string) *Shape {
	edgeMap, ok := shape.Edges[shapeName]
	if ok == false {
		edgeMap, ok = shape.Edges["default"]
	}
	if ok == false {
		fmt.Printf("No edges for shape %s\n", shape.Name)
		return nil
	}
	if edges, ok := edgeMap[edgeName]; ok {
		return edges[rand.Intn(len(edges))]
	}
	fmt.Printf("Can't find edge shape %s for %s\n", edgeName, shape.Name)
	return nil
}

func findShape(name string) *Shape {
	for _, s := range Shapes {
		if s.Name == name {
			return s
		}
	}
	panic("Can't find shape: " + name)
}

func newShape(index int, name string, size [3]float32, px, py, pw, ph float32, img image.Image, fudge float32, imageIndex int, shapeMeta *ShapeMeta, flagsSet map[string]bool, offset [3]float32) *Shape {
	imageBounds := img.Bounds()
	shape := &Shape{
		Index:       len(Shapes),
		Name:        name,
		Size:        size,
		PixelOffset: [2]float32{px, py},
		PixelDim:    [2]float32{pw, ph},
		TexOffset:   [2]float32{float32(px) / float32(imageBounds.Max.X), float32(py) / float32(imageBounds.Max.Y)},
		TexDim:      [2]float32{float32(pw) / float32(imageBounds.Max.X), float32(ph) / float32(imageBounds.Max.Y)},
		Fudge:       fudge,
		ImageIndex:  imageIndex,
		ShapeMeta:   shapeMeta,
		Flags:       flagsSet,
		Edges:       map[string]map[string][]*Shape{},
		Offset:      offset,
	}

	// create a half-size thumbnail
	rgba := image.NewRGBA(image.Rect(0, 0, int(pw), int(ph)))
	draw.Draw(rgba, image.Rect(0, 0, int(pw), int(ph)), img, image.Point{int(px), int(py)}, draw.Src)
	w := uint(float32(rgba.Bounds().Max.X) / shapeMeta.DpiMultiplier)
	h := uint(float32(rgba.Bounds().Max.Y) / shapeMeta.DpiMultiplier)
	resized := resize.Resize(w, h, rgba, resize.NearestNeighbor)
	shape.Image = image.NewRGBA(resized.Bounds())
	draw.Draw(shape.Image, resized.Bounds(), resized, image.ZP, draw.Src)

	return shape
}

func loadImage(path string) (image.Image, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	return img, err
}

func (shape *Shape) Traverse(fx func(x, y, z int)) {
	for xx := 0; xx < int(shape.Size[0]); xx++ {
		for yy := 0; yy < int(shape.Size[1]); yy++ {
			for zz := 0; zz < int(shape.Size[2]); zz++ {
				fx(xx, yy, zz)
			}
		}
	}
}
