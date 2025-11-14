// Package generation provides deterministic procedural level generation.
// It uses seeded PRNGs to create identical obstacle layouts across all clients.
package generation

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
)

const (
	// ChunkSize is the width of each chunk in pixels (~5 screen widths).
	ChunkSize = 5000.0

	// ObstacleTypeTall represents a tall, thin "firewall" obstacle.
	ObstacleTypeTall = 1

	// ObstacleTypeLow represents a low, wide "data-block" obstacle.
	ObstacleTypeLow = 2

	// ObstacleTypeSpike represents a small "glitch" spike obstacle.
	ObstacleTypeSpike = 3

	// MinObstaclesPerChunk is the minimum number of obstacles in a chunk.
	MinObstaclesPerChunk = 3

	// MaxObstaclesPerChunk is the maximum number of obstacles in a chunk.
	MaxObstaclesPerChunk = 8

	// ObstacleSpacing is the minimum spacing between obstacles in pixels.
	ObstacleSpacing = 300.0
)

// Obstacle represents a single obstacle in the game world.
type Obstacle struct {
	// Type identifies the obstacle variant (1=tall, 2=low, 3=spike).
	Type int `json:"t"`

	// X is the horizontal position in pixels (absolute world coordinates).
	X float64 `json:"x"`

	// Y is the vertical position in pixels (0=ground level).
	Y float64 `json:"y"`
}

// Chunk represents a segment of the procedurally generated level.
// Each chunk contains obstacles positioned deterministically based on
// the master seed and chunk ID.
type Chunk struct {
	// ID is the unique identifier for this chunk (0, 1, 2, ...).
	ID int `json:"id"`

	// Obstacles is the list of obstacles in this chunk.
	Obstacles []Obstacle `json:"obs"`
}

// GenerateChunk creates a deterministic chunk of level obstacles.
// The same masterSeed and chunkID will always produce the same obstacle layout.
// This ensures all connected clients see identical levels.
//
// The algorithm:
//  1. Computes a unique seed from hash(masterSeed + chunkID)
//  2. Initializes a PRNG with that seed
//  3. Generates 3-8 obstacles with random types and positions
//  4. Ensures obstacles are spaced appropriately
//
// Parameters:
//   - masterSeed: The global seed for the entire game session
//   - chunkID: The zero-based index of this chunk (0, 1, 2, ...)
//
// Returns:
//   - *Chunk: A chunk with deterministically generated obstacles
//
// Example:
//
//	chunk := GenerateChunk("vibe-runner-12345", 5)
//	// chunk.ID == 5
//	// chunk.Obstacles contains 3-8 obstacles at X positions [25000, 30000)
func GenerateChunk(masterSeed string, chunkID int) *Chunk {
	// Compute deterministic seed for this specific chunk
	// Using SHA-256 ensures good distribution and no collisions
	seedSource := fmt.Sprintf("%s-%d", masterSeed, chunkID)
	hash := sha256.Sum256([]byte(seedSource))

	// Convert first 8 bytes of hash to int64 seed
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	// Initialize PRNG with computed seed
	rng := rand.New(rand.NewSource(seed))

	// Determine number of obstacles for this chunk
	obstacleCount := MinObstaclesPerChunk + rng.Intn(MaxObstaclesPerChunk-MinObstaclesPerChunk+1)

	// Calculate chunk boundaries
	chunkStartX := float64(chunkID) * ChunkSize
	chunkEndX := float64(chunkID+1) * ChunkSize

	obstacles := make([]Obstacle, 0, obstacleCount)

	// Generate obstacles with spacing
	currentX := chunkStartX + 500.0 // Start 500px into chunk for safety

	for i := 0; i < obstacleCount && currentX < chunkEndX-500.0; i++ {
		// Choose random obstacle type (1-3)
		obstacleType := 1 + rng.Intn(3)

		// Randomize X position with spacing
		xOffset := rng.Float64() * 200.0 // Add up to 200px variation
		obstacleX := currentX + xOffset

		// Ensure obstacle stays within chunk bounds
		if obstacleX >= chunkEndX {
			break
		}

		// Y position: 0 for ground level (could add elevated platforms later)
		obstacleY := 0.0

		obstacles = append(obstacles, Obstacle{
			Type: obstacleType,
			X:    obstacleX,
			Y:    obstacleY,
		})

		// Move to next obstacle position with spacing
		currentX += ObstacleSpacing + rng.Float64()*200.0
	}

	return &Chunk{
		ID:        chunkID,
		Obstacles: obstacles,
	}
}
