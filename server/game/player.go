package game

// Player represents a single player in the game world.
// Each player has a unique ID, position, velocity, and state flags.
//
// The player's position is measured in pixels with the origin (0,0)
// at the top-left corner of the screen. Y increases downward.
//
// Physics constants used:
//   - Gravity: 1200 pixels/second²
//   - Jump velocity: -600 pixels/second (negative = upward)
//   - Ground Y position: 440 pixels
type Player struct {
	// ID is the unique player identifier assigned on join.
	// Used to identify this player in network messages.
	ID int

	// Name is the player's display name (sanitized for XSS prevention).
	// Maximum 30 characters.
	Name string

	// X is the horizontal position in pixels.
	// Players spawn at X=100 and move right at constant speed.
	X float64

	// Y is the vertical position in pixels (0=top, increases downward).
	// Players spawn at Y=440 (ground level).
	Y float64

	// VelocityY is the vertical velocity in pixels/second.
	// Positive = moving down, negative = moving up.
	// Affected by gravity each tick.
	VelocityY float64

	// IsGrounded indicates if the player is standing on the ground.
	// True when Y >= 440 (ground level).
	// Must be true to allow jumping.
	IsGrounded bool

	// IsAlive indicates if the player is still in the game.
	// Set to false when player collides with an obstacle.
	// Dead players are excluded from state broadcasts.
	IsAlive bool
}

// NewPlayer creates a new player with default spawn values.
// The player spawns at position (100, 440) on the ground,
// with zero velocity and alive state.
//
// Parameters:
//   - id: Unique player identifier
//   - name: Player's display name (should already be sanitized)
//
// Returns:
//   - *Player: New player instance ready for gameplay
func NewPlayer(id int, name string) *Player {
	return &Player{
		ID:         id,
		Name:       name,
		X:          100.0,  // Spawn position X
		Y:          440.0,  // Spawn at ground level
		VelocityY:  0.0,    // No initial vertical velocity
		IsGrounded: true,   // Start on ground
		IsAlive:    true,   // Start alive
	}
}

// Jump applies upward velocity to make the player jump.
// Only works if the player is grounded and alive.
//
// When called successfully:
//   - Sets VelocityY to -600 (upward)
//   - Sets IsGrounded to false
//
// If the player is not grounded or dead, this does nothing.
// This prevents double-jumping and jumping after death.
func (p *Player) Jump() {
	// Only allow jumping if grounded and alive
	if p.IsGrounded && p.IsAlive {
		p.VelocityY = -600.0 // Jump velocity (pixels/second, upward)
		p.IsGrounded = false
	}
}

// Kill marks the player as dead.
// Dead players are excluded from state broadcasts and cannot perform actions.
// This is called when the player collides with an obstacle.
func (p *Player) Kill() {
	p.IsAlive = false
}
