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
                chunk.blocks[x][y][z], err = block.NewBlock(mgl32.Vec3{(float32(x)*2), (float32(y)*2), (float32(z)*2)}, "block.png")
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