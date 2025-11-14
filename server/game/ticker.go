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
// then broadcasts the updated state to all clients.
//
// The ticker performs these operations each tick:
//  1. Gets all active players from game state
//  2. Applies gravity to each player
//  3. Updates vertical velocity and position
//  4. Checks for ground collision
//  5. Updates grounded state
//  6. Broadcasts state to all connected clients
//
// This function does not block. It launches a goroutine that runs indefinitely.
// To stop the ticker, cancel the returned stop function (future enhancement).
//
// Parameters:
//   - gameState: The shared game state containing all players
//   - broadcaster: The broadcaster for sending state updates to clients
//
// The function logs tick rate information on startup.
// In production, consider adding a context parameter for graceful shutdown.
func StartGameTicker(gameState *GameState, broadcaster Broadcaster) {
	log.Printf("Game ticker starting at %d Hz (%.1f ms per tick)", TickRate, float64(TickDuration.Milliseconds()))

	// Launch ticker in separate goroutine
	go func() {
		// Create ticker for 20Hz updates (50ms intervals)
		ticker := time.NewTicker(TickDuration)
		defer ticker.Stop()

		tickCount := 0

		// Main game loop - runs indefinitely
		for range ticker.C {
			tickCount++

			// Get all active players
			players := gameState.GetAllPlayers()

			// Update physics for each player
			for _, player := range players {
				// Only update alive players
				if !player.IsAlive {
					continue
				}

				// Apply physics update
				updatePlayerPhysics(player)
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
