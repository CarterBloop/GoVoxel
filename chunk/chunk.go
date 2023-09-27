package chunk

import (
	"GoVoxel/block"

	"github.com/aquilax/go-perlin"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	ChunkWidth  = 16
	ChunkHeight = 32
)

type Chunk struct {
	blocks [ChunkWidth][ChunkHeight][ChunkWidth]*block.Block
}

func NewChunk(xOffset, zOffset int, noiseGen *perlin.Perlin) (*Chunk, error) {
	chunk := &Chunk{}

	for x := 0; x < ChunkWidth; x++ {
		for z := 0; z < ChunkWidth; z++ {
			worldX := float64(xOffset + x)
			worldZ := float64(zOffset + z)
			height := int(noiseGen.Noise2D(worldX/20.0, worldZ/20.0)*float64(ChunkHeight/2) + float64(ChunkHeight/2))

            for y := 0; y < ChunkHeight; y++ {
                texture := "textures/Stone.png"
                if y == height {
                    texture = "textures/Grass.png"
                } else if y > height-4 && y < height {
                    texture = "textures/Dirt.png"
                } else if y < 20 {
                    texture = "textures/Water.png"
                }

                b, err := block.NewBlock(mgl32.Vec3{float32(worldX) * 2, float32(y) * 2, float32(worldZ) * 2}, texture)
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
	for x := 0; x < ChunkWidth; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkWidth; z++ {
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
		if nx < 0 || nx >= ChunkWidth || ny < 0 || ny >= ChunkHeight || nz < 0 || nz >= ChunkWidth || c.blocks[nx][ny][nz] == nil {
			return true
		}
	}

	return false
}