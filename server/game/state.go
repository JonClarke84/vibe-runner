package game

import (
	"sync"
)

// GameState holds the authoritative game state for all connected players.
// It provides thread-safe operations for adding, removing, and accessing players.
//
// The game state is shared across multiple goroutines:
//   - One goroutine per connected client (reads/writes players)
//   - One game ticker goroutine (updates physics for all players)
//   - One broadcast goroutine (reads all players for state messages)
//
// All operations use a mutex to prevent race conditions.
type GameState struct {
	// mu protects concurrent access to the players map.
	// All operations that read or modify players must hold this lock.
	mu sync.RWMutex

	// players maps player ID to Player instance.
	// Only alive players remain in this map.
	// Dead players are removed on disconnect.
	players map[int]*Player
}

// NewGameState creates a new game state with an empty player list.
// This should be called once at server startup.
//
// Returns:
//   - *GameState: New game state instance ready for use
func NewGameState() *GameState {
	return &GameState{
		players: make(map[int]*Player),
	}
}

// AddPlayer adds a new player to the game state.
// This is called when a client successfully joins the game.
//
// The operation is thread-safe and can be called from multiple
// client goroutines concurrently.
//
// Parameters:
//   - player: The player to add (should be created with NewPlayer)
//
// The player is added at their spawn position (100, 440) and
// will be included in the next state broadcast.
func (g *GameState) AddPlayer(player *Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.players[player.ID] = player
}

// RemovePlayer removes a player from the game state.
// This is called when a client disconnects.
//
// The operation is thread-safe and can be called from multiple
// client goroutines concurrently.
//
// Parameters:
//   - playerID: ID of the player to remove
//
// If the player doesn't exist, this does nothing.
// After removal, the player will no longer appear in state broadcasts.
func (g *GameState) RemovePlayer(playerID int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.players, playerID)
}

// GetPlayer retrieves a player by ID.
// This is used when processing player actions (e.g., jump requests).
//
// The operation is thread-safe and uses a read lock to allow
// multiple concurrent reads without blocking.
//
// Parameters:
//   - playerID: ID of the player to retrieve
//
// Returns:
//   - *Player: The player instance, or nil if not found
func (g *GameState) GetPlayer(playerID int) *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.players[playerID]
}

// GetAllPlayers returns a slice of all active players.
// This is used by the game ticker for physics updates and
// by the broadcast system for state messages.
//
// The operation is thread-safe and creates a snapshot of the
// current player list. The returned slice is safe to iterate
// over even if players are added/removed concurrently.
//
// Returns:
//   - []*Player: Slice of all active players (empty if none)
//
// Note: The returned slice contains pointers to the actual Player
// instances, so modifications to individual players will affect
// the game state. Use the player's methods to ensure thread-safety.
func (g *GameState) GetAllPlayers() []*Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Create snapshot of current players
	players := make([]*Player, 0, len(g.players))
	for _, player := range g.players {
		players = append(players, player)
	}
	return players
}

// GetPlayerCount returns the number of active players.
// This is used for logging and monitoring purposes.
//
// The operation is thread-safe and uses a read lock.
//
// Returns:
//   - int: Number of active players
func (g *GameState) GetPlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.players)
}
