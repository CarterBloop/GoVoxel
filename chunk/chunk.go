package chunk

import (
	"GoVoxel/block"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)
type Chunk struct {
    blocks [16][16][16]*block.Block
}

func NewChunk() *Chunk {
    chunk := &Chunk{}
    for x := 0; x < 16; x++ {
        for y := 0; y < 16; y++{
            for z := 0; z < 16; z++{
				err := error(nil)
                chunk.blocks[x][y][z], err = block.NewBlock(mgl32.Vec3{(float32(x)*2), (float32(y)*2), (float32(z)*2)}, "textures/Grass.png")
				if err != nil {
					panic(err)
				}
            }
        }
    }
    return chunk
}

func RenderChunk(chunk *Chunk, modelUniform int32) {
    for x := 0; x < 16; x++ {
        for y := 0; y < 16; y++ {
            for z := 0; z < 16; z++ {
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
    // Check the block on each side. If all sides have blocks, don't render.
    directions := [][3]int{
        {1, 0, 0},  // right
        {-1, 0, 0}, // left
        {0, 1, 0},  // above
        {0, -1, 0}, // below
        {0, 0, 1},  // front
        {0, 0, -1}, // back
    }

    for _, dir := range directions {
        nx, ny, nz := x+dir[0], y+dir[1], z+dir[2]
        if nx < 0 || nx >= 16 || ny < 0 || ny >= 16 || nz < 0 || nz >= 16 || c.blocks[nx][ny][nz] == nil {
            return true
        }
    }

    return false
}