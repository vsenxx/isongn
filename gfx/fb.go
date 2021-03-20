package gfx

import "github.com/go-gl/gl/all-core/gl"

type FrameBuffer struct {
	FrameBuffer      uint32
	texture          uint32
	depthBuffer      uint32
	vao              uint32
	vbo              uint32
	program          uint32
	frameBufferTexID int32
	vertAttrib       uint32
	aspectUniform    int32
	fadeUniform      int32
	texCoordAttrib   uint32
	useAspectRatio   bool
}

func NewFrameBuffer(width, height int32, useAspectRatio bool) *FrameBuffer {
	fb := &FrameBuffer{
		useAspectRatio: useAspectRatio,
	}
	gl.GenFramebuffers(1, &fb.FrameBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.FrameBuffer)
	gl.GenTextures(1, &fb.texture)
	gl.BindTexture(gl.TEXTURE_2D, fb.texture)
	// Give an empty image to OpenGL ( the last "0" )
	pix := []uint8{0}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pix))
	// Poor filtering. Needed !
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	// configure a depth buffer for the frame buffer
	gl.GenRenderbuffers(1, &fb.depthBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, fb.depthBuffer)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, width, height)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, fb.depthBuffer)
	// Set "renderedTexture" as our colour attachement #0
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, fb.texture, 0)
	// Set the list of draw buffers.
	attachments := []uint32{gl.COLOR_ATTACHMENT0}
	gl.DrawBuffers(int32(len(attachments)), &attachments[0])
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Can't configure framebuffer!")
	}

	// The fullscreen quad's FBO
	gl.GenVertexArrays(1, &fb.vao)
	gl.BindVertexArray(fb.vao)
	gl.GenBuffers(1, &fb.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, fb.vbo)
	verts := []float32{
		-1, -1, 0, 0, 0,
		-1, 1, 0, 0, 1,
		1, 1, 0, 1, 1,
		1, 1, 0, 1, 1,
		1, -1, 0, 1, 0,
		-1, -1, 0, 0, 0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(verts), gl.STATIC_DRAW)

	// Create and compile our GLSL program from the shaders
	var err error
	fb.program, err = NewProgram(vertexShaderFb, fragmentShaderFb)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(fb.program)
	fb.frameBufferTexID = gl.GetUniformLocation(fb.program, gl.Str("renderedTexture\x00"))
	gl.BindFragDataLocation(fb.program, 0, gl.Str("outputColor\x00"))
	fb.aspectUniform = gl.GetUniformLocation(fb.program, gl.Str("aspect\x00"))
	fb.fadeUniform = gl.GetUniformLocation(fb.program, gl.Str("fade\x00"))
	fb.vertAttrib = uint32(gl.GetAttribLocation(fb.program, gl.Str("vert\x00")))
	fb.texCoordAttrib = uint32(gl.GetAttribLocation(fb.program, gl.Str("vertTexCoord\x00")))
	gl.GenVertexArrays(1, &fb.vao)

	return fb
}

func (fb *FrameBuffer) Enable(width, height int) {
	// render to the framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.FrameBuffer)
	// Render on the whole framebuffer, complete from the lower left corner to the upper right
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (fb *FrameBuffer) Draw(windowWidth, windowHeight int, fade float32) {
	// render to screen
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	// window size
	gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
	// gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Disable(gl.DEPTH_TEST)
	gl.BindVertexArray(fb.vao)
	gl.EnableVertexAttribArray(fb.vertAttrib)
	gl.EnableVertexAttribArray(fb.texCoordAttrib)
	gl.UseProgram(fb.program)
	gl.Uniform1i(fb.frameBufferTexID, 0)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, fb.texture)
	gl.BindBuffer(gl.ARRAY_BUFFER, fb.vbo)
	gl.VertexAttribPointer(fb.vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.VertexAttribPointer(fb.texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	if fb.useAspectRatio {
		gl.Uniform1f(fb.aspectUniform, float32(windowWidth)/float32(windowHeight))
	} else {
		gl.Uniform1f(fb.aspectUniform, 1)
	}
	gl.Uniform1f(fb.fadeUniform, fade)
	gl.DrawArrays(gl.TRIANGLES, 0, 3*2*3)
}

var vertexShaderFb = `
#version 330
uniform float aspect;
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = vec4(vert.x, vert.y * aspect, vert.z, 1);
}
` + "\x00"

var fragmentShaderFb = `
#version 330
uniform sampler2D tex;
uniform float fade;
in vec2 fragTexCoord;
layout(location = 0) out vec4 outputColor;
void main() {
	vec4 val = texture(tex, fragTexCoord);
	if (val.a < 0.1) {
		discard;
	}
	outputColor = val * fade;
}
` + "\x00"
