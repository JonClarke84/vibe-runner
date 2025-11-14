package network

// Message represents the base structure for all WebSocket messages.
// All messages follow the format: {"e": "event", "d": data}
// where "e" is the event type and "d" contains event-specific data.
//
// This short-key format minimizes bandwidth usage, which is critical
// for real-time multiplayer games sending frequent state updates.
type Message struct {
	// E is the event type (e.g., "join", "welcome", "state", "jump")
	E string `json:"e"`
	// D contains the event-specific data as a raw JSON object.
	// This will be unmarshaled into specific message types.
	D interface{} `json:"d"`
}

// JoinMessage represents a client's request to join the game.
// Sent by client immediately after WebSocket connection is established.
//
// Example JSON:
//   {"e": "join", "d": {"n": "PlayerName"}}
type JoinMessage struct {
	// N is the player's chosen display name (max 30 characters).
	// Will be sanitized server-side to prevent XSS attacks.
	N string `json:"n"`
}

// WelcomeMessage is sent by server after successful join.
// It assigns the client a unique player ID and provides game initialization data.
//
// Example JSON:
//   {"e": "welcome", "d": {"id": 1, "seed": "vibe-runner-1", "serverTime": 1700000000000}}
type WelcomeMessage struct {
	// ID is the unique player identifier assigned by the server.
	// Used to identify this player in all subsequent game state messages.
	ID int `json:"id"`

	// Seed is the master seed for procedural level generation.
	// All clients use this seed to generate identical obstacle patterns.
	// Format: "vibe-runner-{sessionID}"
	Seed string `json:"seed"`

	// ServerTime is the current server timestamp in milliseconds since Unix epoch.
	// Used for clock synchronization and latency calculation.
	ServerTime int64 `json:"serverTime"`
}

// JumpMessage represents a client's request to jump.
// Sent when player presses spacebar or jump button.
//
// Example JSON:
//   {"e": "jump", "d": {"t": 1700000000000}}
type JumpMessage struct {
	// T is the client timestamp when jump was initiated (milliseconds since Unix epoch).
	// Used for input prediction and server reconciliation.
	T int64 `json:"t"`
}

// StateMessage contains the authoritative game state broadcast by server.
// Sent at 20Hz (every 50ms) to all connected clients.
//
// Example JSON:
//   {"e": "state", "d": {"t": 1700000000000, "p": [{"i": 1, "x": 100, "y": 440}]}}
type StateMessage struct {
	// T is the server timestamp when state was generated (milliseconds since Unix epoch).
	T int64 `json:"t"`

	// P is the array of player states (positions only, alive players only).
	P []PlayerState `json:"p"`
}

// PlayerState represents a single player's position in the game world.
// Only includes alive players. Dead players are excluded from state broadcasts.
type PlayerState struct {
	// I is the player ID (matches ID from WelcomeMessage).
	I int `json:"i"`

	// X is the player's horizontal position in pixels.
	X float64 `json:"x"`

	// Y is the player's vertical position in pixels (0 = top, increases downward).
	Y float64 `json:"y"`
}

// DeathMessage notifies a client that their player has died.
// Sent immediately when player collides with an obstacle.
//
// Example JSON:
//   {"e": "death", "d": {"s": 1234}}
type DeathMessage struct {
	// S is the player's final score (distance traveled in pixels).
	S int `json:"s"`
}

// ChunkMessage delivers a procedurally generated level chunk to clients.
// Sent when player approaches a new chunk boundary (within 2 screen widths).
//
// Example JSON:
//   {"e": "chunk", "d": {"id": 10, "obs": [{"t": 1, "x": 15000, "y": 0}]}}
type ChunkMessage struct {
	// ID is the chunk identifier (sequential integer starting at 0).
	// Chunk N covers X range [N*5000, (N+1)*5000)
	ID int `json:"id"`

	// Obs is the array of obstacles in this chunk.
	Obs []ObstacleData `json:"obs"`
}

// ObstacleData represents a single obstacle within a level chunk.
type ObstacleData struct {
	// T is the obstacle type (1=tall, 2=low, 3=spike).
	T int `json:"t"`

	// X is the absolute horizontal position in world coordinates (pixels).
	X float64 `json:"x"`

	// Y is the vertical position (0=ground level, positive=elevated).
	Y float64 `json:"y"`
}
