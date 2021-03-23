package gfx

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/uzudil/isongn/shapes"
	"github.com/uzudil/isongn/world"
)

type Texture struct {
	texture      uint32
	textureIndex int32
}

// Block is a displayed Shape
type Block struct {
	vbo                 uint32
	sizeX, sizeY, sizeZ float32
	shape               *shapes.Shape
	texture             *Texture
	index               int32
}

// BlockPos is a displayed Shape at a location
type BlockPos struct {
	model          mgl32.Mat4
	x, y, z        int
	block          *Block
	dir            shapes.Direction
	animationTimer float64
	animationType  int
	animationSpeed float64
	ScrollOffset   [2]float32
}

type View struct {
	Loader               *world.Loader
	projection, camera   mgl32.Mat4
	program              uint32
	projectionUniform    int32
	cameraUniform        int32
	modelUniform         int32
	textureUniform       int32
	textureOffsetUniform int32
	alphaMinUniform      int32
	viewScrollUniform    int32
	modelScrollUniform   int32
	timeUniform          int32
	heightUniform        int32
	swayEnabledUniform   int32
	vertAttrib           uint32
	texCoordAttrib       uint32
	textures             map[int]*Texture
	blocks               []*Block
	vao                  uint32
	blockPos             [SIZE][SIZE][world.SECTION_Z_SIZE]*BlockPos
	edges                [SIZE][SIZE]*BlockPos
	origins              [SIZE][SIZE][world.SECTION_Z_SIZE]*BlockPos
	zoom                 float64
	shear                [3]float32
	Cursor               *BlockPos
	ScrollOffset         [3]float32
	maxZ                 int
	underShape           *shapes.Shape
}

const viewSize = 10
const SIZE = 64
const VIEW_BORDER = 8

func getProjection(zoom float32, shear [3]float32) mgl32.Mat4 {
	projection := mgl32.Ortho(-viewSize*zoom*0.95, viewSize*zoom*0.95, -viewSize*zoom*0.95, viewSize*zoom*0.95, -viewSize*zoom*2, viewSize*zoom*2)
	m := mgl32.Ident4()
	m.Set(0, 2, shear[0])
	m.Set(1, 2, shear[1])
	m.Set(2, 1, shear[2])
	projection = m.Mul4(projection)
	return projection
}

func InitView(zoom float64, camera, shear [3]float32, loader *world.Loader) *View {
	// does this have to be called in every file?
	var err error
	if err = gl.Init(); err != nil {
		panic(err)
	}

	view := &View{
		zoom:   zoom,
		shear:  shear,
		Loader: loader,
		maxZ:   world.SECTION_Z_SIZE,
	}
	view.projection = getProjection(float32(view.zoom), view.shear)

	// coordinate system: Z is up
	view.camera = mgl32.LookAtV(mgl32.Vec3{camera[0], camera[1], camera[2]}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 1})

	// Configure the vertex and fragment shaders
	view.program, err = NewProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(view.program)
	view.projectionUniform = gl.GetUniformLocation(view.program, gl.Str("projection\x00"))
	view.cameraUniform = gl.GetUniformLocation(view.program, gl.Str("camera\x00"))
	view.modelUniform = gl.GetUniformLocation(view.program, gl.Str("model\x00"))
	view.viewScrollUniform = gl.GetUniformLocation(view.program, gl.Str("viewScroll\x00"))
	view.modelScrollUniform = gl.GetUniformLocation(view.program, gl.Str("modelScroll\x00"))
	view.timeUniform = gl.GetUniformLocation(view.program, gl.Str("time\x00"))
	view.heightUniform = gl.GetUniformLocation(view.program, gl.Str("height\x00"))
	view.swayEnabledUniform = gl.GetUniformLocation(view.program, gl.Str("swayEnabled\x00"))
	view.textureUniform = gl.GetUniformLocation(view.program, gl.Str("tex\x00"))
	view.textureOffsetUniform = gl.GetUniformLocation(view.program, gl.Str("textureOffset\x00"))
	view.alphaMinUniform = gl.GetUniformLocation(view.program, gl.Str("alphaMin\x00"))
	gl.BindFragDataLocation(view.program, 0, gl.Str("outputColor\x00"))
	view.vertAttrib = uint32(gl.GetAttribLocation(view.program, gl.Str("vert\x00")))
	view.texCoordAttrib = uint32(gl.GetAttribLocation(view.program, gl.Str("vertTexCoord\x00")))

	gl.UniformMatrix4fv(view.projectionUniform, 1, false, &view.projection[0])
	gl.UniformMatrix4fv(view.cameraUniform, 1, false, &view.camera[0])
	gl.Uniform1i(view.textureUniform, 0)

	view.textures = map[int]*Texture{}
	gl.GenVertexArrays(1, &view.vao)

	// create a block for each shape
	view.blocks = []*Block{}
	for index, shape := range shapes.Shapes {
		if shape == nil {
			view.blocks = append(view.blocks, nil)
		} else {
			view.blocks = append(view.blocks, view.newBlock(int32(index), shape))
		}
	}
	fmt.Printf("Created %d blocks.\n", len(view.blocks))

	for x := 0; x < SIZE; x++ {
		for y := 0; y < SIZE; y++ {
			for z := 0; z < world.SECTION_Z_SIZE; z++ {
				model := mgl32.Ident4()

				// translate to position
				model.Set(0, 3, float32(x-SIZE/2))
				model.Set(1, 3, float32(y-SIZE/2))
				model.Set(2, 3, float32(z))

				view.blockPos[x][y][z] = &BlockPos{
					x:     x,
					y:     y,
					z:     z,
					model: model,
					block: nil,
				}

				if z == 0 {
					edgeModel := mgl32.Ident4()

					// translate to position
					edgeModel.Set(0, 3, float32(x-SIZE/2))
					edgeModel.Set(1, 3, float32(y-SIZE/2))
					edgeModel.Set(2, 3, float32(z)+0.001)

					view.edges[x][y] = &BlockPos{
						x:             x,
						y:             y,
						z:             z,
						model:         edgeModel,
						block:         nil,
						animationType: shapes.ANIMATION_MOVE,
					}
				}
			}
		}
	}

	view.Cursor = &BlockPos{
		x:     0,
		y:     0,
		z:     0,
		model: mgl32.Ident4(),
		block: nil,
	}

	return view
}

func (view *View) newBlock(index int32, shape *shapes.Shape) *Block {
	b := &Block{
		sizeX: shape.Size[0],
		sizeY: shape.Size[1],
		sizeZ: shape.Size[2],
		shape: shape,
		index: index,
	}

	// Configure the vertex data
	gl.BindVertexArray(view.vao)

	gl.GenBuffers(1, &b.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	verts := b.vertices()
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(verts), gl.STATIC_DRAW)

	// load the texture if needed
	tex, ok := view.textures[shape.ImageIndex]
	if ok == false {
		texID, err := loadTexture(shapes.Images[shape.ImageIndex])
		if err != nil {
			panic(err)
		}
		tex = &Texture{
			texture:      texID,
			textureIndex: int32(len(view.textures)),
		}
		view.textures[shape.ImageIndex] = tex
	}
	b.texture = tex

	return b
}

func loadTexture(img image.Image) (uint32, error) {
	// img := shapes.Images[0]
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

func (b *Block) vertices() []float32 {
	// coord system is: z up, x to left, y to right
	//         z
	//         |
	//         |
	//        / \
	//       /   \
	//      x     y
	w := b.sizeX
	h := b.sizeY
	d := b.sizeZ

	// total width/height of texture
	tx := h + w
	ty := h + d + w

	// fudge factor for edges
	var f float32 = b.shape.Fudge

	points := []float32{
		w, 0, d, f, (w - f) / ty,
		w, 0, 0, f, (w + d) / ty,
		w, h, 0, h / tx, 1 - f,
		0, h, 0, 1 - f, (h + d) / ty,
		0, h, d, 1 - f, (h - f) / ty,
		0, 0, d, w / tx, f,
		w, h, d, h / tx, (w + h - f) / ty,
	}

	// scale and translate tex coords to within larger texture
	for i := 0; i < 7; i++ {
		points[i*5+3] *= b.shape.Tex.TexDim[0]
		points[i*5+3] += b.shape.Tex.TexOffset[0]

		points[i*5+4] *= b.shape.Tex.TexDim[1]
		points[i*5+4] += b.shape.Tex.TexOffset[1]
	}

	left := []int{0, 1, 2, 0, 2, 6}
	right := []int{3, 4, 2, 2, 4, 6}
	top := []int{5, 0, 4, 0, 6, 4}

	v := []float32{}
	for _, side := range [][]int{left, right, top} {
		for _, idx := range side {
			for t := 0; t < 5; t++ {
				v = append(v, points[idx*5+t])
			}
		}
	}
	return v
}

func (view *View) SetMaxZ(z int) {
	view.maxZ = z
}

func (view *View) SetUnderShape(shape *shapes.Shape) {
	view.underShape = shape
}

func (view *View) Load() {
	// reset
	view.traverse(func(x, y, z int) {
		blockPos := view.blockPos[x][y][z]
		blockPos.block = nil
		blockPos.ScrollOffset[0] = 0
		blockPos.ScrollOffset[1] = 0
		view.origins[x][y][z] = nil

		if z == 0 {
			edge := view.edges[x][y]
			edge.block = nil
		}
	})

	// load
	view.traverse(func(x, y, z int) {
		worldX, worldY, worldZ := view.toWorldPos(x, y, z)
		shapeIndex, hasShape := view.Loader.GetShape(worldX, worldY, worldZ)
		if hasShape {
			view.setShapeInner(worldX, worldY, worldZ, shapeIndex, true)
		}
		if z == 0 {
			shapeIndex, hasShape = view.Loader.GetEdge(worldX, worldY)
			view.setEdgeInner(worldX, worldY, shapeIndex, hasShape)
		}
	})
}

func (view *View) toWorldPos(viewX, viewY, viewZ int) (int, int, int) {
	return viewX + (view.Loader.X - SIZE/2), viewY + (view.Loader.Y - SIZE/2), viewZ
}

func (view *View) toViewPos(worldX, worldY, worldZ int) (int, int, int, bool) {
	viewX := worldX - (view.Loader.X - SIZE/2)
	viewY := worldY - (view.Loader.Y - SIZE/2)
	invalidPos := viewX < 0 || viewX >= SIZE || viewY < 0 || viewY >= SIZE || worldZ < 0 || worldZ >= world.SECTION_Z_SIZE
	return viewX, viewY, worldZ, !invalidPos
}

func (view *View) Inside(worldX, worldY int) bool {
	vx, vy, _, validPos := view.toViewPos(worldX, worldY, 1)
	return validPos && vx >= VIEW_BORDER && vx < SIZE-VIEW_BORDER && vy >= VIEW_BORDER && vy < SIZE-VIEW_BORDER
}

func (view *View) InView(worldX, worldY, worldZ int) bool {
	_, _, _, validPos := view.toViewPos(worldX, worldY, worldZ)
	return validPos
}

func (view *View) GetShape(worldX, worldY, worldZ int) (int, int, int, int, bool) {
	viewX, viewY, viewZ, validPos := view.toViewPos(worldX, worldY, worldZ)
	if !validPos || view.origins[viewX][viewY][viewZ] == nil {
		return 0, 0, 0, 0, false
	}
	b := view.origins[viewX][viewY][viewZ]
	if b.block == nil {
		wx, wy, wz := view.toWorldPos(viewX, viewY, viewZ)
		fmt.Printf("Error: origin points to nil block at: %d,%d,%d\n", wx, wy, wz)
		return 0, 0, 0, 0, false
	}
	originWorldX, originWorldY, originWorldZ := view.toWorldPos(b.x, b.y, b.z)
	return b.block.shape.Index, originWorldX, originWorldY, originWorldZ, true
}

func (view *View) Fits(toWorldX, toWorldY, toWorldZ int, fromWorldX, fromWorldY, fromWorldZ int) bool {
	shapeIndex, ox, oy, oz, validPos := view.GetShape(fromWorldX, fromWorldY, fromWorldZ)
	if validPos == false {
		return false
	}
	shape := shapes.Shapes[shapeIndex]

	fits := true
	shape.Traverse(func(x, y, z int) bool {
		viewX, viewY, viewZ, validPos := view.toViewPos(toWorldX+x, toWorldY+y, toWorldZ+z)
		if validPos {
			b := view.origins[viewX][viewY][viewZ]
			if b != nil && b.block != nil {
				wx, wy, wz := view.toWorldPos(b.x, b.y, b.z)
				// a shape is uniquely identified by its world coords (not its shape index)
				if !(wx == ox && wy == oy && wz == oz) {
					fits = false
					return true
				}
			}
		}
		return false
	})
	return fits
}

func (view *View) FindTop(worldX, worldY int, shape *shapes.Shape) int {
	maxZ := 0
	for x := 0; x < int(shape.Size[0]); x++ {
		for y := 0; y < int(shape.Size[1]); y++ {
			for z := view.maxZ - 1; z >= 0; z-- {
				_, _, _, _, found := view.GetShape(worldX+x, worldY+y, z)
				if found && z+1 > maxZ {
					maxZ = z + 1
				}
			}
		}
	}
	return maxZ
}

func (view *View) MoveShape(worldX, worldY, worldZ, newWorldX, newWorldY, newWorldZ int) {
	blockPos, shapeIndex := view.EraseShape(worldX, worldY, worldZ)
	if blockPos != nil {
		view.SetShape(newWorldX, newWorldY, newWorldZ, shapeIndex)
	}
}

func (view *View) SetShape(worldX, worldY, worldZ int, shapeIndex int) *BlockPos {
	view.Loader.SetShape(worldX, worldY, worldZ, shapeIndex)
	return view.setShapeInner(worldX, worldY, worldZ, shapeIndex, true)
}

func (view *View) EraseShape(worldX, worldY, worldZ int) (*BlockPos, int) {
	if shapeIndex, ox, oy, oz, hasShape := view.GetShape(worldX, worldY, worldZ); hasShape {
		view.Loader.EraseShape(ox, oy, oz)
		return view.setShapeInner(ox, oy, oz, shapeIndex, false), shapeIndex
	}

	// sometimes this is called for a shape (creature) no longer in view
	// assume the position is its origin and remove it from the sector
	view.Loader.EraseShape(worldX, worldY, worldZ)
	return nil, 0
}

func (view *View) setShapeInner(worldX, worldY, worldZ int, shapeIndex int, hasShape bool) *BlockPos {
	viewX, viewY, viewZ, validPos := view.toViewPos(worldX, worldY, worldZ)
	if validPos {
		blockPos := view.blockPos[viewX][viewY][viewZ]
		shape := shapes.Shapes[shapeIndex]
		if hasShape {
			blockPos.block = view.blocks[shapeIndex]
			blockPos.model.Set(0, 3, float32(viewX-SIZE/2)+shape.Offset[0])
			blockPos.model.Set(1, 3, float32(viewY-SIZE/2)+shape.Offset[1])
			blockPos.model.Set(2, 3, float32(viewZ)+shape.Offset[2])
		} else {
			blockPos.block = nil
		}

		shape.Traverse(func(shapeX, shapeY, shapeZ int) bool {
			if viewX+shapeX < SIZE && viewY+shapeY < SIZE && viewZ+shapeZ < world.SECTION_Z_SIZE {
				if hasShape {
					view.origins[viewX+shapeX][viewY+shapeY][viewZ+shapeZ] = blockPos
				} else {
					view.origins[viewX+shapeX][viewY+shapeY][viewZ+shapeZ] = nil
				}
			}
			return false
		})

		return blockPos
	}
	return nil
}

func (view *View) GetBlockPos(worldX, worldY, worldZ int) *BlockPos {
	viewX, viewY, viewZ, validPos := view.toViewPos(worldX, worldY, worldZ)
	if validPos {
		return view.blockPos[viewX][viewY][viewZ]
	}
	return nil
}

func (view *View) SetOffset(worldX, worldY, worldZ int, dx, dy float32) {
	blockPos := view.GetBlockPos(worldX, worldY, worldZ)
	if blockPos != nil {
		blockPos.ScrollOffset[0] = dx
		blockPos.ScrollOffset[1] = dy
	}
}

func (view *View) SetShapeAnimation(worldX, worldY, worldZ int, animationType int, dir shapes.Direction, animationSpeed float64) {
	blockPos := view.GetBlockPos(worldX, worldY, worldZ)
	if blockPos != nil {
		blockPos.dir = dir
		blockPos.animationType = animationType
		blockPos.animationSpeed = animationSpeed
	}
}

func (view *View) ClearEdge(worldX, worldY int) {
	view.Loader.ClearEdge(worldX, worldY)
	view.setEdgeInner(worldX, worldY, 0, false)
}

func (view *View) SetEdge(worldX, worldY int, shapeIndex int) {
	view.Loader.SetEdge(worldX, worldY, shapeIndex)
	view.setEdgeInner(worldX, worldY, shapeIndex, true)
}

func (view *View) setEdgeInner(worldX, worldY int, shapeIndex int, hasShape bool) {
	viewX, viewY, _, validPos := view.toViewPos(worldX, worldY, 0)
	if validPos {
		if hasShape {
			view.edges[viewX][viewY].block = view.blocks[shapeIndex]
		} else {
			view.edges[viewX][viewY].block = nil
		}
	}
}

func (view *View) traverse(fx func(x, y, z int)) {
	for x := 0; x < SIZE; x++ {
		for y := 0; y < SIZE; y++ {
			for z := 0; z < world.SECTION_Z_SIZE; z++ {
				fx(x, y, z)
			}
		}
	}
}

func (view *View) SetCursor(shapeIndex int, z int) {
	view.Cursor.model.Set(2, 3, float32(z))
	view.Cursor.block = nil
	if shapeIndex >= 0 {
		view.Cursor.block = view.blocks[shapeIndex]
	}
}

func (view *View) HideCursor() {
	view.Cursor.block = nil
}

func (view *View) Scroll(dx, dy, dz float32) {
	view.ScrollOffset[0] = dx
	view.ScrollOffset[1] = dy
	view.ScrollOffset[2] = dz
}

func (view *View) isUnderShape(x, y, z int) bool {
	if view.underShape == nil {
		return true
	}
	wx, wy, _ := view.toWorldPos(x, y, z)
	if shapeIndex, _, _, _, hasShape := view.GetShape(wx, wy, view.maxZ); hasShape {
		return shapes.Shapes[shapeIndex].Group == view.underShape.Group
	}
	return false
}

type DrawState struct {
	init    bool
	texture uint32
	vbo     uint32
	delta   float64
	time    float64
}

var state DrawState = DrawState{}

func (view *View) Draw(delta float64) {
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.UseProgram(view.program)
	gl.BindVertexArray(view.vao)
	gl.EnableVertexAttribArray(view.vertAttrib)
	gl.EnableVertexAttribArray(view.texCoordAttrib)
	gl.Uniform3fv(view.viewScrollUniform, 1, &view.ScrollOffset[0])
	state.delta = delta
	state.time += delta
	state.init = false
	view.traverse(func(x, y, z int) {
		blockPos := view.blockPos[x][y][z]
		underShape := view.isUnderShape(x, y, z)
		if blockPos.block != nil && z < view.maxZ && underShape {
			blockPos.Draw(view)
		}

		if z == 0 && underShape {
			edge := view.edges[x][y]
			if edge.block != nil {
				edge.Draw(view)
			}
		}
	})
	if view.Cursor.block != nil {
		view.Cursor.Draw(view)
	}
}

func (b *BlockPos) Draw(view *View) {
	if !state.init || state.texture != b.block.texture.texture {
		gl.BindTexture(gl.TEXTURE_2D, b.block.texture.texture)
		state.texture = b.block.texture.texture
	}
	if !state.init || state.vbo != b.block.vbo {
		gl.BindBuffer(gl.ARRAY_BUFFER, b.block.vbo)
		gl.VertexAttribPointer(view.vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
		gl.VertexAttribPointer(view.texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		state.vbo = b.block.vbo
	}
	gl.UniformMatrix4fv(view.modelUniform, 1, false, &b.model[0])
	gl.Uniform2fv(view.modelScrollUniform, 1, &b.ScrollOffset[0])
	gl.Uniform1f(view.alphaMinUniform, b.block.shape.AlphaMin)
	gl.Uniform1f(view.timeUniform, float32(state.time))
	gl.Uniform1f(view.heightUniform, b.block.shape.Size[2])
	if b.block.shape.SwayEnabled {
		gl.Uniform1i(view.swayEnabledUniform, 1)
	} else {
		gl.Uniform1i(view.swayEnabledUniform, 0)
	}

	animated := false
	if b.dir != shapes.DIR_NONE {
		if animation, ok := b.block.shape.Animations[b.animationType]; ok {
			b.incrAnimationStep(animation)
			if steps, ok := animation.Tex[b.dir]; ok {
				gl.Uniform1f(view.textureOffsetUniform, steps[animation.AnimationStep].TexOffset[0])
				animated = true
			}
		}
	}
	if !animated {
		gl.Uniform1f(view.textureOffsetUniform, 0)
	}
	gl.DrawArrays(gl.TRIANGLES, 0, 3*2*3)
	state.init = true
}

func (b *BlockPos) incrAnimationStep(animation *shapes.Animation) {
	b.animationTimer -= state.delta
	if b.animationTimer <= 0 {
		b.animationTimer = b.animationSpeed
		animation.AnimationStep++
	}
	if animation.AnimationStep >= animation.Steps {
		animation.AnimationStep = 0
	}
}

func (view *View) Zoom(zoom float64) {
	view.zoom = math.Min(math.Max(view.zoom-zoom*0.1, 0.35), 16)
	// fmt.Printf("zoom:%f\n", view.zoom)
	view.projection = getProjection(float32(view.zoom), view.shear)
	gl.UseProgram(view.program)
	gl.UniformMatrix4fv(view.projectionUniform, 1, false, &view.projection[0])
}

var vertexShader = `
#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float textureOffset;
uniform vec3 viewScroll;
uniform vec2 modelScroll;
uniform float time;
uniform float height;
uniform int swayEnabled;
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vec2(vertTexCoord.x + textureOffset, vertTexCoord.y);

	float swayX = 0;
	if(swayEnabled == 1) {
		swayX = (vert.z / height) * sin(time) / 10.0;
	}
	float swayY = 0;
	if(swayEnabled == 1) {
		swayY = (vert.z / height) * cos(time) / 10.0;
	}	
	float offsX = modelScroll.x - viewScroll.x + swayX;
	float offsY = modelScroll.y - viewScroll.y + swayY;
	float offsZ = -viewScroll.z;

	// matrix constructor is in column first order
	mat4 modelScroll = mat4(
		1.0, 0.0, 0.0, 0.0,
		0.0, 1.0, 0.0, 0.0,
		0.0, 0.0, 1.0, 0.0,
		model[3][0] + offsX, model[3][1] + offsY, model[3][2] + offsZ, 1.0
	);
    gl_Position = projection * camera * modelScroll * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
uniform sampler2D tex;
uniform float alphaMin;
in vec2 fragTexCoord;
layout(location = 0) out vec4 outputColor;
void main() {
	vec4 val = texture(tex, fragTexCoord);
	if (val.a < alphaMin) {
		discard;
	}
	outputColor = val;
}
` + "\x00"
