package main

import (
	"fmt"
	"go/build"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 800
const windowHeight = 600

func init() {
	runtime.LockOSThread()
}

func main() {
	window := Setup()
	defer glfw.Terminate()
	Render(window)
}

func Setup() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	return window
}

func Render(window *glfw.Window) {
    program, err := newProgram(vertexShader, fragmentShader)
    if err != nil {
        panic(err)
    }

    gl.UseProgram(program)

    projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)
    projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
    gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

    camera := NewCamera(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 1, 0}, -90, 0)
	cameraMatrix := camera.GetViewMatrix()
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &cameraMatrix[0])

	var lastX, lastY float64 = windowWidth / 2, windowHeight / 2
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		xOffset := float32(xpos - lastX)
		yOffset := float32(lastY - ypos) // Reversed since y-coordinates range from bottom to top
		lastX = xpos
		lastY = ypos

		camera.ProcessMouseMovement(xOffset, yOffset, true)
	})

    textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
    gl.Uniform1i(textureUniform, 0)

    modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))

    gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

    // Create a new block
    block, err := NewBlock(mgl32.Vec3{0, 0, 0}, "square.png")
    if err != nil {
        log.Fatalln(err)
    }

    // Configure global settings
    gl.Enable(gl.DEPTH_TEST)
    gl.DepthFunc(gl.LESS)
    gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	// Main loop
	var lastFrame float64 = 0.0   // Time of the last frame
	var deltaTime float32 = 0.0   // Time difference between the current and the last frame
    for !window.ShouldClose() {

		currentFrame := glfw.GetTime()
		deltaTime = float32(currentFrame - lastFrame)
		lastFrame = currentFrame

		if window.GetKey(glfw.KeyW) == glfw.Press {
			camera.ProcessKeyboard("FORWARD", deltaTime)
		}
		if window.GetKey(glfw.KeyS) == glfw.Press {
			camera.ProcessKeyboard("BACKWARD", deltaTime)
		}
		if window.GetKey(glfw.KeyA) == glfw.Press {
			camera.ProcessKeyboard("LEFT", deltaTime)
		}
		if window.GetKey(glfw.KeyD) == glfw.Press {
			camera.ProcessKeyboard("RIGHT", deltaTime)
		}
        gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

        // Render block
        gl.UseProgram(program)

        // Set the block's model matrix
        gl.UniformMatrix4fv(modelUniform, 1, false, &block.model[0])

		cameraMatrix = camera.GetViewMatrix()
    	gl.UniformMatrix4fv(cameraUniform, 1, false, &cameraMatrix[0])


        gl.BindVertexArray(block.vao)
        gl.ActiveTexture(gl.TEXTURE0)
        gl.BindTexture(gl.TEXTURE_2D, block.texture)

        gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3) // assuming the cubeVertices is setup for 12 triangles (6 faces * 2 triangles per face)

        // Maintenance
        window.SwapBuffers()
        glfw.PollEvents()
    }
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

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

var vertexShader = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330

uniform sampler2D tex;

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    outputColor = texture(tex, fragTexCoord);
}
` + "\x00"

var cubeVertices = []float32{
	//  X, Y, Z, U, V
	// Bottom
	-1.0, -1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,

	// Top
	-1.0, 1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,

	// Front
	-1.0, -1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,

	// Back
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 1.0,

	// Left
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,

	// Right
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
}

// Set the working directory to the root of Go package, so that its assets can be accessed.
func init() {
	dir, err := importPathToDir("github.com/go-gl/example/gl41core-cube")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}

type Block struct {
	model   mgl32.Mat4
	texture uint32
	vao     uint32
}

func NewBlock(pos mgl32.Vec3, texturePath string) (*Block, error) {
	block := &Block{}
	// set the model matrix to move the block to the specified position
	block.model = mgl32.Translate3D(pos[0], pos[1], pos[2])

	// Load the texture
	tex, err := newTexture(texturePath)
	if err != nil {
		return nil, err
	}
	block.texture = tex

	// Define the standard cube vertex data
	vertices := cubeVertices

	// Create VAO
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coordinate attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0) // Unbind VAO

	block.vao = vao

	return block, nil
}

type Camera struct {
    Position mgl32.Vec3
    Front    mgl32.Vec3
    Up       mgl32.Vec3
    Right    mgl32.Vec3
    WorldUp  mgl32.Vec3
    Yaw      float32
    Pitch    float32
}

func NewCamera(position, up mgl32.Vec3, yaw, pitch float32) *Camera {
    camera := &Camera{Position: position, WorldUp: up, Yaw: yaw, Pitch: pitch}
    camera.updateCameraVectors()
    return camera
}

func (c *Camera) GetViewMatrix() mgl32.Mat4 {
    return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

func (c *Camera) ProcessKeyboard(direction string, deltaTime float32) {
    velocity := 2.5 * deltaTime
    if direction == "FORWARD" {
        c.Position = c.Position.Add(c.Front.Mul(velocity))
    }
    if direction == "BACKWARD" {
        c.Position = c.Position.Sub(c.Front.Mul(velocity))
    }
    if direction == "LEFT" {
        c.Position = c.Position.Sub(c.Right.Mul(velocity))
    }
    if direction == "RIGHT" {
        c.Position = c.Position.Add(c.Right.Mul(velocity))
    }
}

func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32, constrainPitch bool) {
    xoffset *= 0.1
    yoffset *= 0.1

    c.Yaw += xoffset
    c.Pitch += yoffset

    if constrainPitch {
        if c.Pitch > 89.0 {
            c.Pitch = 89.0
        }
        if c.Pitch < -89.0 {
            c.Pitch = -89.0
        }
    }
    c.updateCameraVectors()
}

func (c *Camera) updateCameraVectors() {
    front := mgl32.Vec3{
        float32(math.Cos(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
        float32(math.Sin(float64(mgl32.DegToRad(c.Pitch)))),
        float32(math.Sin(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
    }
    c.Front = front.Normalize()
    c.Right = c.Front.Cross(c.WorldUp).Normalize()
    c.Up = c.Right.Cross(c.Front).Normalize()
}