package chai

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"syscall/js"
	"unsafe"
)

const SHAPES_SHADER_VERTEX = `#version 300 es

precision mediump float;

in vec2 coordinates;
in vec4 colors;

out vec4 vertex_FragColor;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;

void main(void) {
	vec4 global_position = vec4(0.0);
	global_position = view_matrix * vec4(coordinates, 0.0, 1.0);
	global_position.z = 0.0;
	global_position.w = 1.0;		
	gl_Position = global_position;
	
	vertex_FragColor = colors;
}
`
const SHAPES_SHADER_FRAGMENT = `#version 300 es

precision mediump float;

in vec4 vertex_FragColor;

out vec4 fragColor;
void main(void) {
    fragColor = vertex_FragColor;
}
`

var canvasContext js.Value

type RGBA8 struct {
	r, g, b, a uint8
}

func (rgba8 *RGBA8) SetColorR(r uint8) {
	rgba8.r = r
}

func (rgba8 *RGBA8) SetColorRFloat32(r float32) {
	rgba8.r = Float1ToUint8_255(r)
}

func (rgba8 *RGBA8) SetColorG(g uint8) {
	rgba8.g = g
}

func (rgba8 *RGBA8) SetColorGFloat32(g float32) {
	rgba8.g = Float1ToUint8_255(g)
}

func (rgba8 *RGBA8) SetColorB(b uint8) {
	rgba8.b = b
}

func (rgba8 *RGBA8) SetColorBFloat32(b float32) {
	rgba8.b = Float1ToUint8_255(b)
}

func (rgba8 *RGBA8) SetColorA(a uint8) {
	rgba8.a = a
}

func (rgba8 *RGBA8) SetColorAFloat32(a float32) {
	rgba8.a = Float1ToUint8_255(a)
}

func (rgba8 *RGBA8) GetColorRFloat32() float32 {
	return Uint8ToFloat1(rgba8.r)
}
func (rgba8 *RGBA8) GetColorGFloat32() float32 {
	return Uint8ToFloat1(rgba8.g)
}
func (rgba8 *RGBA8) GetColorBFloat32() float32 {
	return Uint8ToFloat1(rgba8.b)
}
func (rgba8 *RGBA8) GetColorAFloat32() float32 {
	return Uint8ToFloat1(rgba8.a)
}

func NewRGBA8(r, g, b, a uint8) RGBA8 {
	return RGBA8{r, g, b, a}
}

func NewRGBA8Float(r, g, b, a float32) RGBA8 {
	return RGBA8{Float1ToUint8_255(r), Float1ToUint8_255(g), Float1ToUint8_255(b), Float1ToUint8_255(a)}
}

func GetRandomRGBA8() RGBA8 {
	return NewRGBA8(Float1ToUint8_255(rand.Float32()), Float1ToUint8_255(rand.Float32()), Float1ToUint8_255(rand.Float32()), 255)
}

var WHITE = RGBA8{255, 255, 255, 255}

const VertexSize = 20

type Vertex struct {
	Coordinates Vector2f
	Color       RGBA8
	UV          Vector2f
}

func NewVertex(_coords, _uv Vector2f, _color RGBA8) Vertex {
	return Vertex{_coords, _color, _uv}
}

const vertexByteSize uintptr = unsafe.Sizeof(Vertex{})

type ShapeBatch struct {
	vao              js.Value
	vbo, ibo         js.Value
	Vertices         []Vertex
	Indices          []int32
	NumberOfElements int
	Shader           ShaderProgram
	LineWidth        float32
	Initialized      bool
}

func (_shapesB *ShapeBatch) Init() {
	_shapesB.LineWidth = 50

	_shapesB.Vertices = make([]Vertex, 0)
	_shapesB.Indices = make([]int32, 0)

	_shapesB.vao = canvasContext.Call("createVertexArray")
	canvasContext.Call("bindVertexArray", _shapesB.vao)

	_shapesB.vbo = canvasContext.Call("createBuffer")
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), _shapesB.vbo)

	canvasContext.Call("enableVertexAttribArray", 0)
	canvasContext.Call("enableVertexAttribArray", 1)
	canvasContext.Call("enableVertexAttribArray", 2)

	canvasContext.Call("vertexAttribPointer", 0, 2, canvasContext.Get("FLOAT"), false, VertexSize, 0)
	canvasContext.Call("vertexAttribPointer", 1, 4, canvasContext.Get("UNSIGNED_BYTE"), true, VertexSize, 8)
	canvasContext.Call("vertexAttribPointer", 2, 2, canvasContext.Get("FLOAT"), false, VertexSize, 12)

	_shapesB.ibo = canvasContext.Call("createBuffer")

	canvasContext.Call("bindVertexArray", js.Null())
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), js.Null())
	canvasContext.Call("bindBuffer", canvasContext.Get("ELEMENT_ARRAY_BUFFER"), js.Null())

	canvasContext.Call("disableVertexAttribArray", 0)
	canvasContext.Call("disableVertexAttribArray", 1)
	canvasContext.Call("disableVertexAttribArray", 2)

	_shapesB.Shader.ParseShader(SHAPES_SHADER_VERTEX, SHAPES_SHADER_FRAGMENT)
	//_shapesB.Shader.ParseShaderFromFile("shapes.shader")
	_shapesB.Shader.CreateShaderProgram()
	_shapesB.Shader.AddAttribute("coordinates")
	_shapesB.Shader.AddAttribute("colors")
	_shapesB.Initialized = true
}

func (_sp *ShapeBatch) DrawLine(_from, _to Vector2f, _color RGBA8) {
	//var offset Vector2f = NewVector2f(_sp.LineWidth/2.0, 0.0)
	var offset Vector2f

	offset = _to.Subtract(_from)
	offset = offset.Perpendicular()
	offset = offset.Normalize()
	offset = offset.Scale(_sp.LineWidth)

	vertsSize := len(_sp.Vertices)
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _from.Subtract(offset), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _to.Subtract(offset), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _from.Add(offset), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _to.Add(offset), Color: _color})

	_sp.Indices = append(_sp.Indices, int32(vertsSize))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+3))
}

func (_sp *ShapeBatch) DrawRect(_center, _dimensions Vector2f, _color RGBA8) {
	offsetX := _dimensions.Scale(0.5).X
	offsetY := _dimensions.Scale(0.5).Y

	lineDifference := _sp.LineWidth

	_sp.DrawLine(NewVector2f(-offsetX-lineDifference, offsetY).Add(_center), NewVector2f(offsetX+lineDifference, offsetY).Add(_center), _color)
	_sp.DrawLine(NewVector2f(-offsetX-lineDifference, -offsetY).Add(_center), NewVector2f(offsetX+lineDifference, -offsetY).Add(_center), _color)
	_sp.DrawLine(NewVector2f(offsetX, offsetY+lineDifference).Add(_center), NewVector2f(offsetX, -offsetY-lineDifference).Add(_center), _color)
	_sp.DrawLine(NewVector2f(-offsetX, offsetY+lineDifference).Add(_center), NewVector2f(-offsetX, -offsetY-lineDifference).Add(_center), _color)
}

func (_sp *ShapeBatch) DrawRectRotated(_center, _dimensions Vector2f, _color RGBA8, _angle float32) {
	offsetX := _dimensions.Scale(0.5).X
	offsetY := _dimensions.Scale(0.5).Y

	lineDifference := _sp.LineWidth

	_sp.DrawLine(NewVector2f(-offsetX-lineDifference, offsetY).Add(_center).Rotate(_angle, _center), NewVector2f(offsetX+lineDifference, offsetY).Add(_center).Rotate(_angle, _center), _color)
	_sp.DrawLine(NewVector2f(-offsetX-lineDifference, -offsetY).Add(_center).Rotate(_angle, _center), NewVector2f(offsetX+lineDifference, -offsetY).Add(_center).Rotate(_angle, _center), _color)
	_sp.DrawLine(NewVector2f(offsetX, offsetY+lineDifference).Add(_center).Rotate(_angle, _center), NewVector2f(offsetX, -offsetY-lineDifference).Add(_center).Rotate(_angle, _center), _color)
	_sp.DrawLine(NewVector2f(-offsetX, offsetY+lineDifference).Add(_center).Rotate(_angle, _center), NewVector2f(-offsetX, -offsetY-lineDifference).Add(_center).Rotate(_angle, _center), _color)
}

func (_sp *ShapeBatch) DrawTriangle(_center, _dimensions Vector2f, _color RGBA8) {
	numOfVertices := 3

	pos := [3]Vector2f{}
	for i := 0; i < numOfVertices; i++ {
		angle := (float32(i) / float32(3.0)) * 2.0 * PI
		pos[i] = NewVector2f(_center.X+float32(math.Cos(float64(angle)))*_dimensions.Y, _center.Y+float32(math.Sin(float64(angle)))*_dimensions.X)
	}

	for i := 0; i < numOfVertices-1; i++ {
		_sp.DrawLine(pos[i], pos[i+1], _color)
	}
	_sp.DrawLine(pos[numOfVertices-1], pos[0], _color)
}

func (_sp *ShapeBatch) DrawTriangleRotated(_center, _dimensions Vector2f, _color RGBA8, rotation float32) {
	numOfVertices := 3

	pos := [3]Vector2f{}
	for i := 0; i < numOfVertices; i++ {
		angle := (float32(i) / float32(3.0)) * 2.0 * PI
		pos[i] = NewVector2f(_center.X+float32(math.Cos(float64(angle)))*_dimensions.Y, _center.Y+float32(math.Sin(float64(angle)))*_dimensions.X)
		pos[i] = pos[i].Rotate(rotation, _center)
	}

	for i := 0; i < numOfVertices-1; i++ {
		_sp.DrawLine(pos[i], pos[i+1], _color)
	}
	_sp.DrawLine(pos[numOfVertices-1], pos[0], _color)
}

func (_sp *ShapeBatch) DrawCircle(_center Vector2f, _radius float32, _color RGBA8) {
	numOfVertices := 16

	pos := [16]Vector2f{}
	for i := 0; i < numOfVertices; i++ {
		angle := (float32(i) / float32(numOfVertices)) * 2.0 * PI
		pos[i] = NewVector2f(_center.X+float32(math.Cos(float64(angle)))*_radius, _center.Y+float32(math.Sin(float64(angle)))*_radius)
	}

	for i := 0; i < numOfVertices-1; i++ {
		_sp.DrawLine(pos[i], pos[i+1], _color)
	}
	_sp.DrawLine(pos[numOfVertices-1], pos[0], _color)
}

func (_sp *ShapeBatch) DrawFillRect(_center, _dimensions Vector2f, _color RGBA8) {
	offset := _dimensions.Scale(0.5)

	vertsSize := len(_sp.Vertices)
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.Subtract(offset), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.SubtractXY(offset.X, -offset.Y), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.AddXY(offset.X, -offset.Y), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.Add(offset), Color: _color})

	_sp.Indices = append(_sp.Indices, int32(vertsSize))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+3))
}

func (_sp *ShapeBatch) DrawFillRectRotated(_center, _dimensions Vector2f, _color RGBA8, _rotation float32) {
	offset := _dimensions.Scale(0.5)

	vertsSize := len(_sp.Vertices)
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.Subtract(offset).Rotate(_rotation, _center), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.SubtractXY(offset.X, -offset.Y).Rotate(_rotation, _center), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.AddXY(offset.X, -offset.Y).Rotate(_rotation, _center), Color: _color})
	_sp.Vertices = append(_sp.Vertices, Vertex{Coordinates: _center.Add(offset).Rotate(_rotation, _center), Color: _color})

	_sp.Indices = append(_sp.Indices, int32(vertsSize))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+2))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+1))
	_sp.Indices = append(_sp.Indices, int32(vertsSize+3))
}

func (_sp *ShapeBatch) finalize() {
	if len(_sp.Vertices) == 0 {
		_sp.NumberOfElements = 0
		return
	}
	canvasContext.Call("bindVertexArray", _sp.vao)

	jsVerts := vertexBufferToJsVertexBuffer(_sp.Vertices)
	jsElem := int32BufferToJsInt32Buffer(_sp.Indices)

	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), _sp.vbo)
	canvasContext.Call("bufferData", canvasContext.Get("ARRAY_BUFFER"), jsVerts, canvasContext.Get("STATIC_DRAW"))

	canvasContext.Call("bindBuffer", canvasContext.Get("ELEMENT_ARRAY_BUFFER"), _sp.ibo)
	canvasContext.Call("bufferData", canvasContext.Get("ELEMENT_ARRAY_BUFFER"), jsElem, canvasContext.Get("STATIC_DRAW"))

	canvasContext.Call("enableVertexAttribArray", 0)
	canvasContext.Call("enableVertexAttribArray", 1)
	canvasContext.Call("enableVertexAttribArray", 2)

	canvasContext.Call("vertexAttribPointer", 0, 2, canvasContext.Get("FLOAT"), false, VertexSize, 0)
	canvasContext.Call("vertexAttribPointer", 1, 4, canvasContext.Get("UNSIGNED_BYTE"), true, VertexSize, 8)
	canvasContext.Call("vertexAttribPointer", 2, 2, canvasContext.Get("FLOAT"), false, VertexSize, 12)

	canvasContext.Call("bindVertexArray", js.Null())
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), js.Null())
	canvasContext.Call("bindBuffer", canvasContext.Get("ELEMENT_ARRAY_BUFFER"), js.Null())

	_sp.NumberOfElements = len(_sp.Indices)

	_sp.Vertices = _sp.Vertices[:0]
	_sp.Indices = _sp.Indices[:0]
}

func (_sp *ShapeBatch) Render(cam *Camera2D) {
	_sp.finalize()

	UseShader(&_sp.Shader)

	viewMatrix := cam.viewMatrix.Data()

	viewBuffer := new(bytes.Buffer)
	{
		err := binary.Write(viewBuffer, binary.LittleEndian, viewMatrix[:])
		Assert(err == nil, "Error writing viewBuffer")
	}

	byte_slice2 := viewBuffer.Bytes()

	b2 := js.Global().Get("Uint8Array").New(len(viewMatrix) * 4)
	js.CopyBytesToJS(b2, byte_slice2)

	viewMatrixJS := js.Global().Get("Float32Array").New(b2.Get("buffer"), b2.Get("byteOffset"), b2.Get("byteLength").Int()/4)
	for i := 0; i < len(viewMatrix); i++ {
		viewMatrixJS.SetIndex(i, js.ValueOf(viewMatrix[i]))
	}

	viewmatrix_loc := canvasContext.Call("getUniformLocation", _sp.Shader.ShaderProgramID, "view_matrix")
	canvasContext.Call("uniformMatrix4fv", viewmatrix_loc, false, viewMatrixJS)
	canvasContext.Call("bindVertexArray", _sp.vao)

	canvasContext.Call("drawElements", canvasContext.Get("TRIANGLES"), _sp.NumberOfElements, canvasContext.Get("UNSIGNED_INT"), 0)
	UnuseShader()
}

/*
##############################################################
################ Sprite Batch - Sprite Batch #################
##############################################################
*/

const SPRITES_SHADER_VERTEX = `#version 300 es

precision mediump float;

in vec2 coordinates;
in vec4 colors;
in vec2 uv;

out vec4 vertex_FragColor;
out vec2 vertex_UV;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;

void main(void) {
	vec4 global_position = vec4(0.0);
	global_position = view_matrix * vec4(coordinates, 0.0, 1.0);
	global_position.z = 0.0;
	global_position.w = 1.0;		
	gl_Position = global_position;
	
	
	vertex_FragColor = colors;
	vertex_UV = uv;
}`
const SPRITES_SHADER_FRAGMENT = `#version 300 es

precision mediump float;

in vec4 vertex_FragColor;
in vec2 vertex_UV;

uniform sampler2D genericSampler;

out vec4 fragColor;

void main(void) {
	vec4 thisColor = vertex_FragColor * texture(genericSampler, vertex_UV);
	fragColor = thisColor;
}`

type SpriteGlyph struct {
	bottomleft, topleft, topright, bottomright Vertex
	texture                                    *Texture2D
}

func NewSpriteGlyph(_pos, _dimensions, _uv1 Vector2f, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) SpriteGlyph {
	var tempGlyph SpriteGlyph

	halfDim := _dimensions.Scale(0.5)

	tempGlyph.bottomleft = NewVertex(_pos.Subtract(halfDim), NewVector2f(_uv1.X, _uv2.Y), _tint)
	tempGlyph.topleft = NewVertex(_pos.Add(NewVector2f(-halfDim.X, halfDim.Y)), _uv1, _tint)
	tempGlyph.topright = NewVertex(_pos.Add(halfDim), NewVector2f(_uv2.X, _uv1.Y), _tint)
	tempGlyph.bottomright = NewVertex(_pos.Add(NewVector2f(halfDim.X, -halfDim.Y)), _uv2, _tint)

	tempGlyph.texture = _texture

	return tempGlyph
}

func NewSpriteGlyphRotated(_pos, _dimensions, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8, _rotation float32) SpriteGlyph {
	var tempGlyph SpriteGlyph

	halfDim := _dimensions.Scale(0.5)

	tempGlyph.bottomleft = NewVertex(_pos.Subtract(halfDim).Rotate(_rotation, _pos), NewVector2f(_uv1.X, _uv2.Y), _tint)
	tempGlyph.topleft = NewVertex(_pos.Add(NewVector2f(-halfDim.X, halfDim.Y)).Rotate(_rotation, _pos), _uv1, _tint)
	tempGlyph.topright = NewVertex(_pos.Add(halfDim).Rotate(_rotation, _pos), NewVector2f(_uv2.X, _uv1.Y), _tint)
	tempGlyph.bottomright = NewVertex(_pos.Add(NewVector2f(halfDim.X, -halfDim.Y)).Rotate(_rotation, _pos), _uv2, _tint)

	tempGlyph.texture = _texture

	return tempGlyph
}

type RenderBatch struct {
	offset, numberOfVertices int
	texture                  *Texture2D
}

func NewRenderBatch(_offset, _numberOfVertices int, _texture *Texture2D) RenderBatch {
	return RenderBatch{_offset, _numberOfVertices, _texture}
}

type SpriteBatch struct {
	vbo js.Value
	vao js.Value

	shader ShaderProgram

	renderBatches []RenderBatch
	spriteGlyphs  []SpriteGlyph
}

func (self *SpriteBatch) Reset() {
	self.renderBatches = self.renderBatches[:0]
	self.spriteGlyphs = self.spriteGlyphs[:0]
	//`self.Init("")
}

func (self *SpriteBatch) Init(_shader_path string) {

	self.renderBatches = make([]RenderBatch, 0)
	self.spriteGlyphs = make([]SpriteGlyph, 0)

	self.vao = canvasContext.Call("createVertexArray")
	canvasContext.Call("bindVertexArray", self.vao)

	self.vbo = canvasContext.Call("createBuffer")
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), self.vbo)

	canvasContext.Call("enableVertexAttribArray", 0)
	canvasContext.Call("enableVertexAttribArray", 1)
	canvasContext.Call("enableVertexAttribArray", 2)

	canvasContext.Call("vertexAttribPointer", 0, 2, canvasContext.Get("FLOAT"), false, VertexSize, 0)
	canvasContext.Call("vertexAttribPointer", 1, 4, canvasContext.Get("UNSIGNED_BYTE"), true, VertexSize, 8)
	canvasContext.Call("vertexAttribPointer", 2, 2, canvasContext.Get("FLOAT"), false, VertexSize, 12)

	canvasContext.Call("bindVertexArray", js.Null())
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), js.Null())

	canvasContext.Call("disableVertexAttribArray", 0)
	canvasContext.Call("disableVertexAttribArray", 1)
	canvasContext.Call("disableVertexAttribArray", 2)

	if _shader_path == "" {
		self.shader.ParseShader(SPRITES_SHADER_VERTEX, SPRITES_SHADER_FRAGMENT)
	} else {
		self.shader.ParseShaderFromFile(_shader_path)
	}
	self.shader.CreateShaderProgram()
	self.shader.AddAttribute("coordinates")
	self.shader.AddAttribute("colors")
	self.shader.AddAttribute("uv")

}

func (self *SpriteBatch) DrawSprite(_center, _dimensions, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_center, _dimensions, _uv1, _uv2, _texture, _tint))
}

func (self *SpriteBatch) DrawSpriteOrigin(_center, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_center, NewVector2f(float32(_texture.Width), float32(_texture.Height)), _uv1, _uv2, _texture, _tint))

}
func (self *SpriteBatch) DrawSpriteOriginScaled(_center, _uv1, _uv2 Vector2f, _scale float32, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_center, NewVector2f(float32(_texture.Width), float32(_texture.Height)).Scale(_scale), _uv1, _uv2, _texture, _tint))

}
func (self *SpriteBatch) DrawSpriteBottomLeft(_pos, _dimensions, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_pos.Add(_dimensions.Scale(0.5)), _dimensions, _uv1, _uv2, _texture, _tint))
}
func (self *SpriteBatch) DrawSpriteBottomRight(_pos, _dimensions, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_pos.AddXY(-_dimensions.Scale(0.5).X, _dimensions.Scale(0.5).Y), _dimensions, _uv1, _uv2, _texture, _tint))
}
func (self *SpriteBatch) DrawSpriteBottomLeftOrigin(_pos, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyph(_pos.Subtract(NewVector2f(float32(_texture.Width), float32(_texture.Height)).Scale(0.5)), NewVector2f(float32(_texture.Width), float32(_texture.Height)), _uv1, _uv2, _texture, _tint))

}

func (self *SpriteBatch) DrawSpriteOriginRotated(_center, _uv1, _uv2 Vector2f, _texture *Texture2D, _tint RGBA8, _rotation float32) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyphRotated(_center, NewVector2f(float32(_texture.Width), float32(_texture.Height)), _uv1, _uv2, _texture, _tint, _rotation))
}

func (self *SpriteBatch) DrawSpriteOriginScaledRotated(_center, _uv1, _uv2 Vector2f, _scale float32, _texture *Texture2D, _tint RGBA8, _rotation float32) {
	self.spriteGlyphs = append(self.spriteGlyphs, NewSpriteGlyphRotated(_center, NewVector2f(float32(_texture.Width), float32(_texture.Height)).Scale(_scale), _uv1, _uv2, _texture, _tint, _rotation))
}

func (self *SpriteBatch) finalize() {
	self.createRenderBatches()

}

func (self *SpriteBatch) Render(cam *Camera2D) {
	self.finalize()

	UseShader(&self.shader)

	viewMatrix := cam.viewMatrix.Data()

	viewBuffer := new(bytes.Buffer)
	{
		err := binary.Write(viewBuffer, binary.LittleEndian, viewMatrix[:])
		Assert(err == nil, "Error writing viewBuffer")
	}

	byte_slice2 := viewBuffer.Bytes()

	b2 := js.Global().Get("Uint8Array").New(len(viewMatrix) * 4)
	js.CopyBytesToJS(b2, byte_slice2)

	viewMatrixJS := js.Global().Get("Float32Array").New(b2.Get("buffer"), b2.Get("byteOffset"), b2.Get("byteLength").Int()/4)
	for i := 0; i < len(viewMatrix); i++ {
		viewMatrixJS.SetIndex(i, js.ValueOf(viewMatrix[i]))
	}

	viewmatrix_loc := canvasContext.Call("getUniformLocation", self.shader.ShaderProgramID, "view_matrix")
	canvasContext.Call("uniformMatrix4fv", viewmatrix_loc, false, viewMatrixJS)

	canvasContext.Call("activeTexture", canvasContext.Get("TEXTURE0"))
	canvasContext.Call("uniform1i", self.shader.GetUniformLocation("genericSampler"), 0)

	canvasContext.Call("bindVertexArray", self.vao)
	for i := 0; i < len(self.renderBatches); i++ {
		canvasContext.Call("bindTexture", canvasContext.Get("TEXTURE_2D"), self.renderBatches[i].texture.textureId)
		canvasContext.Call("drawArrays", canvasContext.Get("TRIANGLES"), self.renderBatches[i].offset, self.renderBatches[i].numberOfVertices)
	}
	canvasContext.Call("bindVertexArray", js.Null())

	self.renderBatches = self.renderBatches[:0]
	self.spriteGlyphs = self.spriteGlyphs[:0]
}

func (self *SpriteBatch) createRenderBatches() {
	vertices := make([]Vertex, len(self.spriteGlyphs)*6)

	if len(vertices) == 0 {
		return
	}

	offset := 0
	vertexNum := 0

	self.renderBatches = append(self.renderBatches, NewRenderBatch(offset, 6, self.spriteGlyphs[0].texture))
	vertices[vertexNum] = self.spriteGlyphs[0].bottomleft
	vertexNum++
	vertices[vertexNum] = self.spriteGlyphs[0].topright
	vertexNum++
	vertices[vertexNum] = self.spriteGlyphs[0].bottomright
	vertexNum++
	vertices[vertexNum] = self.spriteGlyphs[0].bottomleft
	vertexNum++
	vertices[vertexNum] = self.spriteGlyphs[0].topright
	vertexNum++
	vertices[vertexNum] = self.spriteGlyphs[0].topleft
	vertexNum++
	offset += 6

	for i := 1; i < len(self.spriteGlyphs); i++ {
		if !self.spriteGlyphs[i-1].texture.textureId.Equal(self.spriteGlyphs[i].texture.textureId) {
			self.renderBatches = append(self.renderBatches, NewRenderBatch(offset, 6, self.spriteGlyphs[i].texture))
		} else {
			self.renderBatches[len(self.renderBatches)-1].numberOfVertices += 6
		}

		vertices[vertexNum] = self.spriteGlyphs[i].bottomleft
		vertexNum++
		vertices[vertexNum] = self.spriteGlyphs[i].topright
		vertexNum++
		vertices[vertexNum] = self.spriteGlyphs[i].bottomright
		vertexNum++
		vertices[vertexNum] = self.spriteGlyphs[i].bottomleft
		vertexNum++
		vertices[vertexNum] = self.spriteGlyphs[i].topright
		vertexNum++
		vertices[vertexNum] = self.spriteGlyphs[i].topleft
		vertexNum++
		offset += 6
	}
	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), self.vbo)

	jsVerts := vertexBufferToJsVertexBuffer(vertices)

	canvasContext.Call("bufferData", canvasContext.Get("ARRAY_BUFFER"), jsVerts, canvasContext.Get("STATIC_DRAW"))

	canvasContext.Call("bindBuffer", canvasContext.Get("ARRAY_BUFFER"), js.Null())

}
