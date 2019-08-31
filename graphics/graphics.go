package graphics

import (
	"fmt"
	"gopengl/graphics/opengl"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

/*
All opengl commands must be executed in the main thread, thus all execution must occur in this file,
graphics enqueues tasks that are then performed by this file and execute in the go context.
TODO: Add some sync functionality if needed (eg for lighting)

All render objects are also stored here so that they can be cleaned up on program closure.
*/

/*
Setup
*/

func Init() {
	err := gl.Init()

	if err != nil {
		panic(err)
	}
}

var windowWidth float32 = 800
var windowHeight float32 = 600

func SetWindowSize(width, height float32) {
	windowWidth = width
	windowHeight = height
}

/*
Render object handling
*/

/*
Render objects are abstractions on a VAO, each render object can only be used for a single texture,
General transformations can be applied to the entire render object which are performed on every vert
*/

type RenderObject struct {
	vao      *opengl.VAO
	texture  *opengl.Texture
	freeVert int
	maxVert  int
}

var renderObjects = make([]*RenderObject, 0)
var window *glfw.Window

//Creation and deletion

func SetWindow(newWindow *glfw.Window) {
	window = newWindow
}

func CreateRenderObject(obj *RenderObject, size int, texture string, defaultShader bool) {
	vao := opengl.CreateVAO(uint32(size), texture, defaultShader)
	vao.CreateBuffers()

	obj.vao = vao
	obj.texture = vao.Texture
	obj.freeVert = 0
	obj.maxVert = size

	renderObjects = append(renderObjects, obj)
}

func DeleteRenderObjects() {
	for _, obj := range renderObjects {
		obj.Delete()
	}
}

/*
Render Object methods
*/

func Render() {
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	for _, obj := range renderObjects {
		obj.Render()
	}

	Poll(window)
}

func (obj *RenderObject) Render() {
	vertNum := obj.PrepRender()
	gl.DrawArrays(gl.TRIANGLES, 0, vertNum)
	obj.FinishRender()
}

func (obj *RenderObject) PrepRender() int32 {
	return obj.vao.PrepRender()
}

func (obj *RenderObject) FinishRender() {
	obj.vao.FinishRender()
}

func (obj *RenderObject) Delete() {
	obj.vao.Delete()
}

// AddSquare ... add a square to the render object, position is from the top left in pixels
// Returns index of new objects first vertex
func (obj *RenderObject) AddSquare(x, y, xTex, yTex, width, widthTex float32) int {
	verts := []float32{
		// Upper right triangle
		x, y,
		x + width, y,
		x + width, y - width,

		// Lower left triangle
		x, y,
		x + width, y - width,
		x, y - width,
	}

	texs := []float32{
		// Upper right triangle
		xTex, yTex,
		xTex + widthTex, yTex,
		xTex + widthTex, yTex + widthTex,

		// Lower left triangle
		xTex, yTex,
		xTex + widthTex, yTex + widthTex,
		xTex, yTex + widthTex,
	}

	verts = PixToScreen(verts)
	texs = obj.texture.PixToTex(texs)

	if obj.freeVert+6 > obj.maxVert {
		panic("Render Object Buffer overflow")
	}

	obj.vao.UpdateBufferIndex(obj.freeVert, verts, texs)
	obj.freeVert += 6

	return obj.freeVert - 6
}

func (obj *RenderObject) ModifyVertSquare(index int, x, y, width float32) {
	verts := []float32{
		// Upper right triangle
		x, y,
		x + width, y,
		x + width, y + width,

		// Lower left triangle
		x, y,
		x + width, y + width,
		x, y + width,
	}

	verts = PixToScreen(verts)

	obj.vao.UpdateVertBufferIndex(index, verts)
}

func (obj *RenderObject) ModifyTexSquare(index int, xTex, yTex, widthTex float32) {
	texs := []float32{
		// Upper right triangle
		xTex, yTex,
		xTex + widthTex, yTex,
		xTex + widthTex, yTex + widthTex,

		// Lower left triangle
		xTex, yTex,
		xTex + widthTex, yTex + widthTex,
		xTex, yTex + widthTex,
	}

	texs = obj.texture.PixToTex(texs)

	obj.vao.UpdateTexBufferIndex(index, texs)
}

func (obj *RenderObject) ModifySquare(index int, x, y, xTex, yTex, width, widthTex float32) {
	obj.ModifyVertSquare(index, x, y, width)
	obj.ModifyTexSquare(index, xTex, yTex, widthTex)
}

// Clear a square, does not delete the object.
func (obj *RenderObject) ClearSquare(index int) {
	obj.ModifyVertSquare(index, 0, 0, 0)
}

func (obj *RenderObject) RotateSquare(x, y, rad float32) {
	obj.vao.SetRotation(x, y, rad)
}

func (obj *RenderObject) TranslateSquare(x, y float32) {
	obj.vao.SetTranslation(x, y)
}

/*
Utility methods
*/

func PixToScreen(coords []float32) []float32 {
	normedCoords := make([]float32, len(coords))
	even := false

	/*
		In opengl the centre of the screen is 0,0 so need to normalize about that point
	*/

	halfWidth := windowWidth / 2
	halfHeight := windowHeight / 2

	for i, coord := range coords {
		even = !even

		if even {
			normedCoords[i] = (coord - halfWidth) / halfWidth

			fmt.Println(normedCoords[i])
			continue
		}
		normedCoords[i] = (coord - halfHeight) / halfHeight
	}

	return normedCoords
}