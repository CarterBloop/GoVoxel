package chunk

import (
	"GoVoxel/block"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	ChunkSize = 16
)

type Chunk struct {
	blocks [ChunkSize][ChunkSize][ChunkSize]*block.Block
}

func NewChunk() (*Chunk, error) {
	chunk := &Chunk{}

	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				b, err := block.NewBlock(mgl32.Vec3{float32(x) * 2, float32(y) * 2, float32(z) * 2}, "textures/Grass.png")
				if err != nil {
					return nil, err
				}
				chunk.blocks[x][y][z] = b
			}
		}
	}

	return chunk, nil
}

func RenderChunk(chunk *Chunk, modelUniform int32) {
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				if !chunk.shouldRenderBlock(x, y, z) {
					continue
				}
				b := chunk.blocks[x][y][z]
				gl.UniformMatrix4fv(modelUniform, 1, false, &b.Model[0])
				gl.BindVertexArray(b.Vao)
				gl.ActiveTexture(gl.TEXTURE0)
				gl.BindTexture(gl.TEXTURE_2D, b.Texture)
				gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3) 
			}
		}
	}
}

func (c *Chunk) shouldRenderBlock(x, y, z int) bool {
	directions := [][3]int{
		{1, 0, 0},
		{-1, 0, 0},
		{0, 1, 0},
		{0, -1, 0},
		{0, 0, 1},
		{0, 0, -1},
	}

	for _, dir := range directions {
		nx, ny, nz := x+dir[0], y+dir[1], z+dir[2]
		if nx < 0 || nx >= ChunkSize || ny < 0 || ny >= ChunkSize || nz < 0 || nz >= ChunkSize || c.blocks[nx][ny][nz] == nil {
			return true
		}
	}

	return false
}