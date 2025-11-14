package generation_test

import (
	"testing"

	"vibe-runner-server/generation"
)

// TestNewChunkManager_InitializesWithSeed creates a manager with given seed.
func TestNewChunkManager_InitializesWithSeed(t *testing.T) {
	// Arrange
	masterSeed := "test-seed-123"

	// Act
	manager := generation.NewChunkManager(masterSeed)

	// Assert
	if manager == nil {
		t.Fatal("NewChunkManager returned nil")
	}
}

// TestGetOrGenerateChunk_FirstCall_GeneratesChunk verifies chunk generation on first request.
func TestGetOrGenerateChunk_FirstCall_GeneratesChunk(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")
	chunkID := 5

	// Act
	chunk := manager.GetOrGenerateChunk(chunkID)

	// Assert
	if chunk == nil {
		t.Fatal("GetOrGenerateChunk returned nil")
	}
	if chunk.ID != chunkID {
		t.Errorf("chunk ID = %d, want %d", chunk.ID, chunkID)
	}
	if len(chunk.Obstacles) == 0 {
		t.Error("chunk should contain obstacles")
	}
}

// TestGetOrGenerateChunk_SecondCall_ReturnsCachedChunk verifies chunk caching.
func TestGetOrGenerateChunk_SecondCall_ReturnsCachedChunk(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")
	chunkID := 3

	// Act
	chunk1 := manager.GetOrGenerateChunk(chunkID)
	chunk2 := manager.GetOrGenerateChunk(chunkID)

	// Assert - should be same pointer (cached)
	if chunk1 != chunk2 {
		t.Error("expected same chunk instance (cached), got different instances")
	}
}

// TestGetChunksInRange_ReturnsChunksForRange verifies range-based chunk retrieval.
func TestGetChunksInRange_ReturnsChunksForRange(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")
	startX := 5000.0  // Middle of chunk 1
	endX := 15000.0   // Middle of chunk 3

	// Act - this should generate chunks 1 and 2
	chunks := manager.GetChunksInRange(startX, endX)

	// Assert
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks in range, got %d", len(chunks))
	}

	// Verify chunks are in order
	for i := 0; i < len(chunks)-1; i++ {
		if chunks[i].ID >= chunks[i+1].ID {
			t.Error("chunks should be returned in ascending ID order")
		}
	}
}

// TestGetChunksInRange_EmptyRange_ReturnsEmpty verifies empty range handling.
func TestGetChunksInRange_EmptyRange_ReturnsEmpty(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")
	startX := 1000.0
	endX := 1000.0 // Same as start

	// Act
	chunks := manager.GetChunksInRange(startX, endX)

	// Assert
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty range, got %d", len(chunks))
	}
}

// TestGenerateAheadForPlayer_GeneratesUpcomingChunks verifies ahead generation.
func TestGenerateAheadForPlayer_GeneratesUpcomingChunks(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")
	playerX := 5000.0 // Player at middle of chunk 1

	// Act - generate 2 chunks ahead
	manager.GenerateAheadForPlayer(playerX, 2)

	// Assert - chunks 1, 2, 3 should exist
	chunk1 := manager.GetOrGenerateChunk(1)
	chunk2 := manager.GetOrGenerateChunk(2)
	chunk3 := manager.GetOrGenerateChunk(3)

	if chunk1 == nil || chunk2 == nil || chunk3 == nil {
		t.Error("expected chunks 1, 2, 3 to be generated")
	}
}

// TestCleanupBehind_RemovesOldChunks verifies garbage collection.
func TestCleanupBehind_RemovesOldChunks(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")

	// Generate chunks 0-5
	for i := 0; i <= 5; i++ {
		manager.GetOrGenerateChunk(i)
	}

	// Act - cleanup chunks behind X=15000 (chunk 3+), keep 1 behind
	minPlayerX := 15000.0
	manager.CleanupBehind(minPlayerX, 1)

	// Assert - chunks 0 and 1 should be removed, 2-5 should remain
	chunks := manager.GetAllChunks()

	for _, chunk := range chunks {
		if chunk.ID < 2 {
			t.Errorf("chunk %d should have been cleaned up", chunk.ID)
		}
	}
}

// TestGetAllChunks_ReturnsAllCachedChunks returns all stored chunks.
func TestGetAllChunks_ReturnsAllCachedChunks(t *testing.T) {
	// Arrange
	manager := generation.NewChunkManager("test-seed")

	// Generate some chunks
	manager.GetOrGenerateChunk(0)
	manager.GetOrGenerateChunk(2)
	manager.GetOrGenerateChunk(5)

	// Act
	chunks := manager.GetAllChunks()

	// Assert
	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}

	// Verify all expected chunks are present
	foundIDs := make(map[int]bool)
	for _, chunk := range chunks {
		foundIDs[chunk.ID] = true
	}

	if !foundIDs[0] || !foundIDs[2] || !foundIDs[5] {
		t.Error("not all generated chunks were returned")
	}
}

// TestChunkManager_DeterministicAcrossInstances verifies different manager
// instances with same seed produce identical chunks.
func TestChunkManager_DeterministicAcrossInstances(t *testing.T) {
	// Arrange
	seed := "test-seed-456"
	manager1 := generation.NewChunkManager(seed)
	manager2 := generation.NewChunkManager(seed)

	// Act
	chunk1 := manager1.GetOrGenerateChunk(10)
	chunk2 := manager2.GetOrGenerateChunk(10)

	// Assert - chunks should have identical obstacles
	if len(chunk1.Obstacles) != len(chunk2.Obstacles) {
		t.Errorf("obstacle count differs: %d vs %d",
			len(chunk1.Obstacles), len(chunk2.Obstacles))
	}

	for i := range chunk1.Obstacles {
		obs1 := chunk1.Obstacles[i]
		obs2 := chunk2.Obstacles[i]

		if obs1.Type != obs2.Type || obs1.X != obs2.X || obs1.Y != obs2.Y {
			t.Errorf("obstacle %d differs between manager instances", i)
		}
	}
}
