package world

import (
	"GoVoxel/chunk"

	"github.com/aquilax/go-perlin"
)

const (
	alpha   = 2.
	beta    = 2.
	n       = 3
	seedVal = int64(1000) // You can change this for different worlds
	WorldSize = 1
)

type World struct {
	Chunks   [][WorldSize]*chunk.Chunk
	NoiseGen *perlin.Perlin
}

func NewWorld() *World {
	noiseGen := perlin.NewPerlin(alpha, beta, n, seedVal)

	world := &World{
		NoiseGen: noiseGen,
	}

	// Create a 5x5 world of chunks
	for i := 0; i < WorldSize; i++ {
		var chunkRow [WorldSize]*chunk.Chunk
		for j := 0; j < WorldSize; j++ {
			chk, _ := chunk.NewChunk(i*chunk.ChunkWidth, j*chunk.ChunkWidth, world.NoiseGen)
			chunkRow[j] = chk
		}
		world.Chunks = append(world.Chunks, chunkRow)
	}

	return world
}

func (w *World) RenderChunks(modelUniform int32) {
	for x := 0; x < WorldSize; x++ {
		for z := 0; z < WorldSize; z++ {
			chunk.RenderChunk(w.Chunks[x][z], modelUniform)
		}
	}
}