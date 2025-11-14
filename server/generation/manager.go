package generation

import (
	"sync"
)

// ChunkManager manages procedural chunk generation and caching.
// It generates chunks on-demand, caches them in memory, and provides
// garbage collection for chunks that are no longer needed.
//
// The manager is thread-safe and can be accessed from multiple goroutines.
type ChunkManager struct {
	// masterSeed is the global seed for this game session.
	// All chunks are derived from this seed to ensure determinism.
	masterSeed string

	// chunks stores generated chunks by their ID.
	// Access must be protected by mutex.
	chunks map[int]*Chunk

	// mu protects concurrent access to the chunks map.
	mu sync.RWMutex
}

// NewChunkManager creates a new chunk manager with the given master seed.
//
// Parameters:
//   - masterSeed: The global seed for this game session. All chunks
//     are derived from this seed to ensure deterministic generation.
//
// Returns:
//   - *ChunkManager: A new manager ready to generate chunks
//
// Example:
//
//	manager := NewChunkManager("vibe-runner-12345")
//	chunk := manager.GetOrGenerateChunk(0)
func NewChunkManager(masterSeed string) *ChunkManager {
	return &ChunkManager{
		masterSeed: masterSeed,
		chunks:     make(map[int]*Chunk),
	}
}

// GetOrGenerateChunk retrieves a chunk from cache or generates it if needed.
// This method is thread-safe and can be called from multiple goroutines.
//
// The chunk is cached after generation, so subsequent calls with the same
// chunkID will return the cached instance.
//
// Parameters:
//   - chunkID: The zero-based index of the chunk to retrieve
//
// Returns:
//   - *Chunk: The requested chunk (either cached or newly generated)
//
// Example:
//
//	chunk := manager.GetOrGenerateChunk(5)
//	// Returns chunk 5 covering X range [25000, 30000)
func (cm *ChunkManager) GetOrGenerateChunk(chunkID int) *Chunk {
	// Try to get from cache first (read lock)
	cm.mu.RLock()
	chunk, exists := cm.chunks[chunkID]
	cm.mu.RUnlock()

	if exists {
		return chunk
	}

	// Generate new chunk (write lock)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check in case another goroutine generated it
	chunk, exists = cm.chunks[chunkID]
	if exists {
		return chunk
	}

	// Generate and cache
	chunk = GenerateChunk(cm.masterSeed, chunkID)
	cm.chunks[chunkID] = chunk

	return chunk
}

// GetChunksInRange returns all chunks that overlap the given X range.
// Chunks are generated if they don't already exist in the cache.
// This method is thread-safe.
//
// Parameters:
//   - startX: The starting X position (in world coordinates)
//   - endX: The ending X position (in world coordinates)
//
// Returns:
//   - []*Chunk: Slice of chunks in range, sorted by chunk ID
//
// Example:
//
//	chunks := manager.GetChunksInRange(5000.0, 15000.0)
//	// Returns chunks 1 and 2
func (cm *ChunkManager) GetChunksInRange(startX, endX float64) []*Chunk {
	if startX >= endX {
		return []*Chunk{}
	}

	// Calculate chunk IDs for range
	startChunkID := int(startX / ChunkSize)
	endChunkID := int(endX / ChunkSize)

	chunks := make([]*Chunk, 0, endChunkID-startChunkID+1)

	// Generate/retrieve all chunks in range
	for chunkID := startChunkID; chunkID <= endChunkID; chunkID++ {
		chunk := cm.GetOrGenerateChunk(chunkID)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// GenerateAheadForPlayer pre-generates chunks ahead of a player's position.
// This ensures chunks are ready before players reach them, preventing
// load-time stutters. This method is thread-safe.
//
// Parameters:
//   - playerX: The player's current X position in world coordinates
//   - chunksAhead: Number of chunks to generate ahead of the player
//
// Example:
//
//	// Player at X=5000, generate 2 chunks ahead
//	manager.GenerateAheadForPlayer(5000.0, 2)
//	// Generates chunks 1, 2, 3
func (cm *ChunkManager) GenerateAheadForPlayer(playerX float64, chunksAhead int) {
	currentChunkID := int(playerX / ChunkSize)

	// Generate current chunk and chunks ahead
	for i := 0; i <= chunksAhead; i++ {
		cm.GetOrGenerateChunk(currentChunkID + i)
	}
}

// CleanupBehind removes chunks that are behind all players to save memory.
// This method is thread-safe.
//
// It keeps a specified number of chunks behind the trailing player for safety
// (in case of server reconciliation or lag).
//
// Parameters:
//   - minPlayerX: The X position of the furthest-behind player
//   - keepBehind: Number of chunks to keep behind the trailing player
//
// Example:
//
//	// Remove chunks more than 1 chunk behind trailing player at X=15000
//	manager.CleanupBehind(15000.0, 1)
//	// Chunks 0, 1 removed; chunks 2+ kept
func (cm *ChunkManager) CleanupBehind(minPlayerX float64, keepBehind int) {
	trailingChunkID := int(minPlayerX / ChunkSize)
	cleanupThreshold := trailingChunkID - keepBehind

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove chunks below threshold
	for chunkID := range cm.chunks {
		if chunkID < cleanupThreshold {
			delete(cm.chunks, chunkID)
		}
	}
}

// GetAllChunks returns a snapshot of all currently cached chunks.
// This method is thread-safe and returns a copy of the chunk slice.
//
// Returns:
//   - []*Chunk: All cached chunks (order not guaranteed)
//
// Example:
//
//	chunks := manager.GetAllChunks()
//	fmt.Printf("Cached %d chunks\n", len(chunks))
func (cm *ChunkManager) GetAllChunks() []*Chunk {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	chunks := make([]*Chunk, 0, len(cm.chunks))
	for _, chunk := range cm.chunks {
		chunks = append(chunks, chunk)
	}

	return chunks
}

// GetOrGenerateChunkInterface is an interface-compatible version of GetOrGenerateChunk.
// This exists to satisfy the game.ChunkManager interface without circular dependencies.
//
// Parameters:
//   - chunkID: The zero-based index of the chunk to retrieve
//
// Returns:
//   - interface{}: The requested chunk as an interface{} (actually *Chunk)
func (cm *ChunkManager) GetOrGenerateChunkInterface(chunkID int) interface{} {
	return cm.GetOrGenerateChunk(chunkID)
}
