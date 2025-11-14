package network

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"vibe-runner-server/game"

	"github.com/gorilla/websocket"
)

// ClientConnection represents a connected client with write capabilities.
// Each client has a dedicated write goroutine that reads from a buffered channel.
// This prevents slow clients from blocking the broadcast.
type ClientConnection struct {
	// PlayerID is the unique identifier for this client's player
	PlayerID int

	// Conn is the WebSocket connection
	Conn *websocket.Conn

	// SendChan is the buffered channel for outgoing messages
	// Buffer size of 10 allows some tolerance for slow clients
	SendChan chan []byte

	// closed indicates if this connection has been closed
	closed bool

	// mu protects the closed flag
	mu sync.Mutex
}

// ClientHub manages all connected clients and broadcasts game state.
// It provides thread-safe add/remove operations and a broadcast function
// for sending state updates to all clients.
type ClientHub struct {
	// clients maps player ID to client connection
	clients map[int]*ClientConnection

	// mu protects concurrent access to the clients map
	mu sync.RWMutex
}

// NewClientHub creates a new client hub for managing connections.
//
// Returns:
//   - *ClientHub: New hub instance ready for use
func NewClientHub() *ClientHub {
	return &ClientHub{
		clients: make(map[int]*ClientConnection),
	}
}

// AddClient registers a new client connection and starts its write goroutine.
// The write goroutine reads from the client's send channel and writes to
// the WebSocket connection.
//
// Parameters:
//   - playerID: Unique player identifier
//   - conn: WebSocket connection for this client
//
// The function starts a goroutine that handles all writes for this client.
// The goroutine exits when the send channel is closed.
func (h *ClientHub) AddClient(playerID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create client connection with buffered send channel
	client := &ClientConnection{
		PlayerID: playerID,
		Conn:     conn,
		SendChan: make(chan []byte, 10), // Buffer 10 messages
		closed:   false,
	}

	h.clients[playerID] = client

	// Start write goroutine for this client
	go client.writeLoop()

	log.Printf("Client added to hub: PlayerID=%d, Total clients: %d", playerID, len(h.clients))
}

// RemoveClient unregisters a client connection and cleans up resources.
// This closes the send channel, which causes the write goroutine to exit.
//
// Parameters:
//   - playerID: Player ID of the client to remove
//
// If the client doesn't exist, this does nothing.
func (h *ClientHub) RemoveClient(playerID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exists := h.clients[playerID]
	if !exists {
		return
	}

	// Mark as closed and close send channel
	client.mu.Lock()
	if !client.closed {
		client.closed = true
		close(client.SendChan)
	}
	client.mu.Unlock()

	delete(h.clients, playerID)

	log.Printf("Client removed from hub: PlayerID=%d, Total clients: %d", playerID, len(h.clients))
}

// BroadcastState sends the current game state to all connected clients.
// This is called by the game ticker at 20Hz.
//
// The function creates a state message with current server time and all
// alive player positions, then sends it to all clients via their send channels.
//
// Parameters:
//   - gameState: The game state containing all players
//
// Slow clients with full send buffers will have messages dropped (non-blocking).
// This prevents slow clients from degrading performance for other clients.
func (h *ClientHub) BroadcastState(gameState *game.GameState) {
	// Get all active players
	players := gameState.GetAllPlayers()

	// Build player state array (only alive players)
	playerStates := make([]PlayerState, 0, len(players))
	for _, player := range players {
		if player.IsAlive {
			playerStates = append(playerStates, PlayerState{
				I: player.ID,
				X: player.X,
				Y: player.Y,
			})
		}
	}

	// Create state message
	stateMsg := Message{
		E: "state",
		D: StateMessage{
			T: time.Now().UnixMilli(),
			P: playerStates,
		},
	}

	// Marshal to JSON once (more efficient than per-client)
	messageBytes, err := json.Marshal(stateMsg)
	if err != nil {
		log.Printf("Failed to marshal state message: %v", err)
		return
	}

	// Broadcast to all clients
	h.mu.RLock()
	defer h.mu.RUnlock()

	for playerID, client := range h.clients {
		// Non-blocking send
		// If client's channel is full, skip this update for them
		select {
		case client.SendChan <- messageBytes:
			// Message queued successfully
		default:
			// Channel full - client is too slow
			log.Printf("Dropped state update for slow client: PlayerID=%d", playerID)
		}
	}
}

// BroadcastChunk sends a level chunk to all connected clients.
// This is called when a new chunk needs to be sent to players (typically
// when the leading player approaches a new chunk boundary).
//
// Parameters:
//   - chunkID: The ID of the chunk to broadcast
//   - chunkData: The chunk data (interface{} to avoid circular dependency)
//
// The function creates a chunk message and sends it to all clients.
// The chunkData is expected to be a *generation.Chunk but we use interface{}
// to avoid import cycles.
func (h *ClientHub) BroadcastChunk(chunkID int, chunkData interface{}) {
	// Convert chunk data to network format
	// We use reflection to extract obstacles without importing generation package
	obstacles := convertChunkToObstacles(chunkData)
	// Create chunk message
	chunkMsg := Message{
		E: "chunk",
		D: ChunkMessage{
			ID:  chunkID,
			Obs: obstacles,
		},
	}

	// Marshal to JSON once
	messageBytes, err := json.Marshal(chunkMsg)
	if err != nil {
		log.Printf("Failed to marshal chunk message: %v", err)
		return
	}

	// Broadcast to all clients
	h.mu.RLock()
	defer h.mu.RUnlock()

	for playerID, client := range h.clients {
		// Non-blocking send
		select {
		case client.SendChan <- messageBytes:
			// Message queued successfully
		default:
			// Channel full
			log.Printf("Dropped chunk update for slow client: PlayerID=%d", playerID)
		}
	}

	log.Printf("Broadcasted chunk %d with %d obstacles to %d clients", chunkID, len(obstacles), len(h.clients))
}

// convertChunkToObstacles converts a generation.Chunk to network ObstacleData format.
// This uses reflection to avoid circular import between network and generation packages.
//
// Parameters:
//   - chunkData: Expected to be *generation.Chunk
//
// Returns:
//   - []ObstacleData: Converted obstacles for network transmission
func convertChunkToObstacles(chunkData interface{}) []ObstacleData {
	// Use type assertion with reflection to extract obstacles
	// The chunk has structure: {ID int, Obstacles []Obstacle}
	// Each Obstacle has: {Type int, X float64, Y float64}

	// Type switch to handle the conversion
	type chunkLike struct {
		ID        int `json:"id"`
		Obstacles []struct {
			Type int     `json:"t"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		} `json:"obs"`
	}

	// Try to marshal and unmarshal to extract data generically
	jsonBytes, err := json.Marshal(chunkData)
	if err != nil {
		log.Printf("Failed to marshal chunk data: %v", err)
		return []ObstacleData{}
	}

	var chunk chunkLike
	if err := json.Unmarshal(jsonBytes, &chunk); err != nil {
		log.Printf("Failed to unmarshal chunk data: %v", err)
		return []ObstacleData{}
	}

	// Convert to network format
	obstacles := make([]ObstacleData, len(chunk.Obstacles))
	for i, obs := range chunk.Obstacles {
		obstacles[i] = ObstacleData{
			T: obs.Type,
			X: obs.X,
			Y: obs.Y,
		}
	}

	return obstacles
}

// writeLoop handles writing messages to the WebSocket connection.
// This runs in a dedicated goroutine per client.
//
// The function reads from the SendChan and writes each message to the WebSocket.
// It exits when SendChan is closed (on client disconnect).
//
// Write errors (e.g., connection closed) are logged but don't crash the goroutine.
// The connection will be cleaned up by the main HandleClient function.
func (c *ClientConnection) writeLoop() {
	// Set write deadline for all writes
	// If a write takes longer than 10 seconds, consider client dead
	writeTimeout := 10 * time.Second

	for messageBytes := range c.SendChan {
		// Check if connection is closed
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			break
		}
		c.mu.Unlock()

		// Set write deadline
		if err := c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			log.Printf("Failed to set write deadline for PlayerID=%d: %v", c.PlayerID, err)
			break
		}

		// Write message to WebSocket
		if err := c.Conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("Failed to write to PlayerID=%d: %v", c.PlayerID, err)
			break
		}
	}

	log.Printf("Write loop exited for PlayerID=%d", c.PlayerID)
}
