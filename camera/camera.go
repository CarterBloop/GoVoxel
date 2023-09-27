package camera

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

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
    velocity := 25 * deltaTime
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