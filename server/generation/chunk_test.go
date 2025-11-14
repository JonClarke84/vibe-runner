package generation_test

import (
	"testing"

	"vibe-runner-server/generation"
)

// TestGenerateChunk_SameSeed_ProducesSameObstacles verifies deterministic generation.
// This is critical for multiplayer - all clients must see identical obstacle layouts.
func TestGenerateChunk_SameSeed_ProducesSameObstacles(t *testing.T) {
	// Arrange
	masterSeed := "test-seed-123"
	chunkID := 5

	// Act
	chunk1 := generation.GenerateChunk(masterSeed, chunkID)
	chunk2 := generation.GenerateChunk(masterSeed, chunkID)

	// Assert
	if chunk1.ID != chunk2.ID {
		t.Errorf("chunk IDs differ: %d vs %d", chunk1.ID, chunk2.ID)
	}

	if len(chunk1.Obstacles) != len(chunk2.Obstacles) {
		t.Errorf("obstacle count differs: %d vs %d", len(chunk1.Obstacles), len(chunk2.Obstacles))
	}

	// Verify each obstacle is identical
	for i := range chunk1.Obstacles {
		obs1 := chunk1.Obstacles[i]
		obs2 := chunk2.Obstacles[i]

		if obs1.Type != obs2.Type {
			t.Errorf("obstacle %d type differs: %d vs %d", i, obs1.Type, obs2.Type)
		}
		if obs1.X != obs2.X {
			t.Errorf("obstacle %d X differs: %.2f vs %.2f", i, obs1.X, obs2.X)
		}
		if obs1.Y != obs2.Y {
			t.Errorf("obstacle %d Y differs: %.2f vs %.2f", i, obs1.Y, obs2.Y)
		}
	}
}

// TestGenerateChunk_DifferentSeeds_ProducesDifferentObstacles ensures
// different seeds create different layouts.
func TestGenerateChunk_DifferentSeeds_ProducesDifferentObstacles(t *testing.T) {
	// Arrange
	seed1 := "seed-A"
	seed2 := "seed-B"
	chunkID := 5

	// Act
	chunk1 := generation.GenerateChunk(seed1, chunkID)
	chunk2 := generation.GenerateChunk(seed2, chunkID)

	// Assert - layouts should be different
	if len(chunk1.Obstacles) == 0 || len(chunk2.Obstacles) == 0 {
		t.Fatal("chunks should contain obstacles")
	}

	// At least one obstacle should differ
	allSame := true
	for i := 0; i < len(chunk1.Obstacles) && i < len(chunk2.Obstacles); i++ {
		if chunk1.Obstacles[i].X != chunk2.Obstacles[i].X ||
			chunk1.Obstacles[i].Type != chunk2.Obstacles[i].Type {
			allSame = false
			break
		}
	}

	if allSame && len(chunk1.Obstacles) == len(chunk2.Obstacles) {
		t.Error("different seeds should produce different obstacle layouts")
	}
}

// TestGenerateChunk_DifferentChunkIDs_ProducesDifferentObstacles ensures
// different chunk IDs create unique layouts even with same master seed.
func TestGenerateChunk_DifferentChunkIDs_ProducesDifferentObstacles(t *testing.T) {
	// Arrange
	masterSeed := "test-seed"
	chunk1ID := 1
	chunk2ID := 2

	// Act
	chunk1 := generation.GenerateChunk(masterSeed, chunk1ID)
	chunk2 := generation.GenerateChunk(masterSeed, chunk2ID)

	// Assert - layouts should be different
	if len(chunk1.Obstacles) == 0 || len(chunk2.Obstacles) == 0 {
		t.Fatal("chunks should contain obstacles")
	}

	// At least one obstacle should differ
	allSame := true
	for i := 0; i < len(chunk1.Obstacles) && i < len(chunk2.Obstacles); i++ {
		if chunk1.Obstacles[i].X != chunk2.Obstacles[i].X ||
			chunk1.Obstacles[i].Type != chunk2.Obstacles[i].Type {
			allSame = false
			break
		}
	}

	if allSame && len(chunk1.Obstacles) == len(chunk2.Obstacles) {
		t.Error("different chunk IDs should produce different obstacle layouts")
	}
}

// TestGenerateChunk_ValidatesChunkID sets the correct ID on returned chunk.
func TestGenerateChunk_ValidatesChunkID(t *testing.T) {
	// Arrange
	masterSeed := "test-seed"
	chunkID := 42

	// Act
	chunk := generation.GenerateChunk(masterSeed, chunkID)

	// Assert
	if chunk.ID != chunkID {
		t.Errorf("chunk ID = %d, want %d", chunk.ID, chunkID)
	}
}

// TestGenerateChunk_ContainsObstacles ensures chunks aren't empty.
func TestGenerateChunk_ContainsObstacles(t *testing.T) {
	// Arrange
	masterSeed := "test-seed"
	chunkID := 10

	// Act
	chunk := generation.GenerateChunk(masterSeed, chunkID)

	// Assert
	if len(chunk.Obstacles) == 0 {
		t.Error("chunk should contain at least one obstacle")
	}
}

// TestGenerateChunk_ObstaclesWithinBounds verifies obstacles are positioned
// within the chunk's boundaries.
func TestGenerateChunk_ObstaclesWithinBounds(t *testing.T) {
	// Arrange
	masterSeed := "test-seed"
	chunkID := 5
	chunkSize := 5000.0 // 5000 pixels as per spec

	// Act
	chunk := generation.GenerateChunk(masterSeed, chunkID)

	// Assert
	minX := float64(chunkID) * chunkSize
	maxX := float64(chunkID+1) * chunkSize

	for i, obs := range chunk.Obstacles {
		if obs.X < minX || obs.X >= maxX {
			t.Errorf("obstacle %d X position %.2f outside chunk bounds [%.2f, %.2f)",
				i, obs.X, minX, maxX)
		}

		// Y should be at ground level (0) or slightly above
		if obs.Y < 0 || obs.Y > 100 {
			t.Errorf("obstacle %d Y position %.2f outside reasonable bounds [0, 100]", i, obs.Y)
		}
	}
}

// TestGenerateChunk_ValidObstacleTypes ensures all obstacle types are valid.
func TestGenerateChunk_ValidObstacleTypes(t *testing.T) {
	// Arrange
	masterSeed := "test-seed"
	chunkID := 10

	// Act
	chunk := generation.GenerateChunk(masterSeed, chunkID)

	// Assert
	validTypes := map[int]bool{
		1: true, // Tall
		2: true, // Low
		3: true, // Spike
	}

	for i, obs := range chunk.Obstacles {
		if !validTypes[obs.Type] {
			t.Errorf("obstacle %d has invalid type %d, expected 1-3", i, obs.Type)
		}
	}
}
