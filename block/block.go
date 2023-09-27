package block

import (
	"fmt"
	"image"
	"image/draw"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Block struct {
	Model   mgl32.Mat4
	Texture uint32
	Vao     uint32
	Vbo     uint32
}

func NewBlock(pos mgl32.Vec3, texturePath string) (*Block, error) {
	block := &Block{
		Model: mgl32.Translate3D(pos[0], pos[1], pos[2]),
	}
	
	// Load texture and handle error
	tex, err := loadTexture(texturePath)
	if err != nil {
		return nil, err
	}
	block.Texture = tex

	// Initialize VAO and VBO
	if err := block.initVAOandVBO(); err != nil {
		return nil, err
	}

	return block, nil
}

func (b *Block) initVAOandVBO() error {
	var vao, vbo uint32

	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.BufferData(gl.ARRAY_BUFFER, len(CubeVertices)*4, gl.Ptr(CubeVertices), gl.STATIC_DRAW)

	// Define vertex attributes
	defineVertexAttributes()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0) // Unbind buffer
	gl.BindVertexArray(0)             // Unbind VAO

	b.Vao = vao
	b.Vbo = vbo

	return nil
}

func defineVertexAttributes() {
	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coordinate attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
}

func loadTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	defer imgFile.Close()

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

	// Set texture parameters
	setTextureParameters()

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

func setTextureParameters() {
	// Nearest neighbor filtering for pixellated look
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Texture wrapping/clamping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}