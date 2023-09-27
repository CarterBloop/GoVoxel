package main

import (
	"fmt"
	_ "image/png"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"GoVoxel/camera"
	"GoVoxel/shaders"
	"GoVoxel/world"
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

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "GoVoxel", nil, nil)
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
    program, err := newProgram(shaders.VertexShader, shaders.FragmentShader)
    if err != nil {
        panic(err)
    }

    gl.UseProgram(program)

    projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 1000.0)
    projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
    gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

    cam := camera.NewCamera(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 1, 0}, -90, 0)
	cameraMatrix := cam.GetViewMatrix()
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &cameraMatrix[0])

	var lastX, lastY float64 = windowWidth / 2, windowHeight / 2
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		xOffset := float32(xpos - lastX)
		yOffset := float32(lastY - ypos) // Reversed since y-coordinates range from bottom to top
		lastX = xpos
		lastY = ypos

		cam.ProcessMouseMovement(xOffset, yOffset, true)
	})

    textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
    gl.Uniform1i(textureUniform, 0)

    modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))

    gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Create a new world
	newWorld := world.NewWorld()
	
    if err != nil {
        log.Fatalln(err)
    }

    // Configure global settings
    gl.Enable(gl.DEPTH_TEST)
    gl.DepthFunc(gl.LESS)
    gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW) // Assuming your vertices are counter-clockwise

	// Main loop
	var lastFrame float64 = 0.0   // Time of the last frame
	var deltaTime float32 = 0.0   // Time difference between the current and the last frame
    for !window.ShouldClose() {

		// User Input
		currentFrame := glfw.GetTime()
		deltaTime = float32(currentFrame - lastFrame)
		lastFrame = currentFrame

		if window.GetKey(glfw.KeyW) == glfw.Press {
			cam.ProcessKeyboard(camera.FORWARD, deltaTime)
		}
		if window.GetKey(glfw.KeyS) == glfw.Press {
			cam.ProcessKeyboard(camera.BACKWARD, deltaTime)
		}
		if window.GetKey(glfw.KeyA) == glfw.Press {
			cam.ProcessKeyboard(camera.LEFT, deltaTime)
		}
		if window.GetKey(glfw.KeyD) == glfw.Press {
			cam.ProcessKeyboard(camera.RIGHT, deltaTime)
		}

        gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

        // Render chunks
		newWorld.RenderChunks(modelUniform)

        gl.UseProgram(program)

        // Set the block's model matrix
		cameraMatrix = cam.GetViewMatrix()
    	gl.UniformMatrix4fv(cameraUniform, 1, false, &cameraMatrix[0])

        // Maintenance
        window.SwapBuffers()
        glfw.PollEvents()
    }
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := shaders.CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := shaders.CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
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