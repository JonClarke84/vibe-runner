package game

import (
	"log"
	"time"
)

// Broadcaster is an interface for broadcasting game state to clients.
// This interface prevents circular dependencies between game and network packages.
type Broadcaster interface {
	BroadcastState(gameState *GameState)
}

// ChunkBroadcaster is an interface for broadcasting chunk data to clients.
// This extends Broadcaster to support procedural level generation in Phase 4.
type ChunkBroadcaster interface {
	Broadcaster
	// BroadcastChunk sends a chunk to all connected clients
	// The obstacles parameter uses the network.ObstacleData type
	BroadcastChunk(chunkID int, obstacles interface{})
}

// ChunkManager is an interface for procedural chunk generation.
// This prevents circular dependencies between game and generation packages.
type ChunkManager interface {
	// GenerateAheadForPlayer pre-generates chunks ahead of player position
	GenerateAheadForPlayer(playerX float64, chunksAhead int)

	// CleanupBehind removes chunks behind all players
	CleanupBehind(minPlayerX float64, keepBehind int)

	// GetOrGenerateChunkInterface retrieves or generates a chunk by ID
	GetOrGenerateChunkInterface(chunkID int) interface{}
}

// Physics constants matching the Phase 1 client implementation.
// These values determine how the game feels and must remain synchronized
// with any client-side prediction in later phases.
const (
	// Gravity is the downward acceleration in pixels/second²
	Gravity = 1200.0

	// JumpVelocity is the initial upward velocity when jumping (pixels/second, negative = upward)
	JumpVelocity = -600.0

	// GroundY is the Y coordinate of the ground (pixels from top)
	GroundY = 440.0

	// PlayerWidth is the player's hitbox width in pixels
	PlayerWidth = 40.0

	// PlayerHeight is the player's hitbox height in pixels
	PlayerHeight = 60.0

	// TickRate is the server update frequency (Hz)
	TickRate = 20

	// TickDuration is the time between ticks (50ms for 20Hz)
	TickDuration = time.Second / TickRate

	// DeltaTime is the time step for physics calculations (seconds)
	// This is 0.05 seconds (50ms) for 20Hz
	DeltaTime = 1.0 / float64(TickRate)
)

// StartGameTicker launches the main game loop in a goroutine.
// The game loop runs at 20Hz (50ms per tick) and updates all player physics,
// manages chunk generation/broadcasting, then broadcasts the updated state to all clients.
//
// The ticker performs these operations each tick:
//  1. Gets all active players from game state
//  2. Applies gravity to each player
//  3. Updates vertical velocity and position
//  4. Checks for ground collision
//  5. Updates grounded state
//  6. Generates chunks ahead of leading player
//  7. Broadcasts new chunks to clients
//  8. Cleans up old chunks behind all players
//  9. Broadcasts state to all connected clients
//
// This function does not block. It launches a goroutine that runs indefinitely.
// To stop the ticker, cancel the returned stop function (future enhancement).
//
// Parameters:
//   - gameState: The shared game state containing all players
//   - broadcaster: The broadcaster for sending state and chunk updates to clients
//   - chunkManager: The chunk manager for procedural level generation (nil to disable)
//
// The function logs tick rate information on startup.
// In production, consider adding a context parameter for graceful shutdown.
func StartGameTicker(gameState *GameState, broadcaster Broadcaster, chunkManager ChunkManager) {
	log.Printf("Game ticker starting at %d Hz (%.1f ms per tick)", TickRate, float64(TickDuration.Milliseconds()))

	// Launch ticker in separate goroutine
	go func() {
		// Create ticker for 20Hz updates (50ms intervals)
		ticker := time.NewTicker(TickDuration)
		defer ticker.Stop()

		tickCount := 0
		lastBroadcastedChunk := -1

		// Main game loop - runs indefinitely
		for range ticker.C {
			tickCount++

			// Get all active players
			players := gameState.GetAllPlayers()

			// Track player positions for chunk management
			var maxPlayerX, minPlayerX float64
			if len(players) > 0 {
				maxPlayerX = players[0].X
				minPlayerX = players[0].X
			}

			// Update physics for each player
			for _, player := range players {
				// Only update alive players
				if !player.IsAlive {
					continue
				}

				// Apply physics update
				updatePlayerPhysics(player)

				// Track leading and trailing player positions
				if player.X > maxPlayerX {
					maxPlayerX = player.X
				}
				if player.X < minPlayerX {
					minPlayerX = player.X
				}
			}

			// Phase 4: Chunk management (if chunk manager provided)
			if chunkManager != nil && len(players) > 0 {
				// Generate chunks ahead of leading player
				// Generate 2 chunks ahead (within 2 screen widths as per spec)
				chunkManager.GenerateAheadForPlayer(maxPlayerX, 2)

				// Broadcast new chunks to clients
				// Determine which chunk the leading player is approaching
				leadingChunkID := int(maxPlayerX / 5000.0)

				// Broadcast next chunk if we haven't sent it yet
				nextChunkID := leadingChunkID + 1
				if nextChunkID > lastBroadcastedChunk {
					chunk := chunkManager.GetOrGenerateChunkInterface(nextChunkID)
					if chunk != nil && broadcaster != nil {
						// Type assert to ChunkBroadcaster if supported
						if chunkBroadcaster, ok := broadcaster.(ChunkBroadcaster); ok {
							chunkBroadcaster.BroadcastChunk(nextChunkID, chunk)
							lastBroadcastedChunk = nextChunkID
						}
					}
				}

				// Cleanup old chunks (every 4 seconds = 80 ticks)
				if tickCount%80 == 0 {
					// Keep 1 chunk behind trailing player for safety
					chunkManager.CleanupBehind(minPlayerX, 1)
				}
			}

			// Broadcast state to all clients at 20Hz
			broadcaster.BroadcastState(gameState)

			// Log debug info every 2 seconds (20 ticks/sec * 2 = 40 ticks)
			if tickCount%40 == 0 {
				log.Printf("[Tick %d] Active players: %d", tickCount, len(players))
			}
		}
	}()
}

// updatePlayerPhysics applies physics calculations to a single player for one tick.
// This function updates vertical velocity due to gravity, updates position,
// and handles ground collision detection.
//
// Physics equations used:
//   - velocityY += gravity * deltaTime (acceleration due to gravity)
//   - y += velocityY * deltaTime (position update from velocity)
//
// Ground collision:
//   - If y >= GroundY (440): player hits ground
//   - Set y = GroundY, velocityY = 0, isGrounded = true
//
// Parameters:
//   - player: The player to update (modified in place)
//
// The function does not acquire any locks. The caller (game ticker) is
// responsible for thread-safety when accessing player state.
func updatePlayerPhysics(player *Player) {
	// Apply gravity to vertical velocity
	// velocityY increases (more downward) each tick due to gravity
	player.VelocityY += Gravity * DeltaTime

	// Update vertical position based on velocity
	// y increases (moves down) when velocityY is positive
	player.Y += player.VelocityY * DeltaTime

	// Check ground collision
	if player.Y >= GroundY {
		// Player has hit or passed through ground
		// Snap to ground level
		player.Y = GroundY

		// Stop vertical movement
		player.VelocityY = 0.0

		// Mark as grounded (allows jumping again)
		player.IsGrounded = true
	} else {
		// Player is in the air
		player.IsGrounded = false
	}

	// Horizontal movement (constant speed, no acceleration)
	// Players move right at fixed speed
	// TODO: Implement in Chunk 4 or later - for now, players stay at spawn X=100
	// player.X += PlayerSpeed * DeltaTime
}
