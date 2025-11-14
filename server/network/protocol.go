package network

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"strings"
	"sync"
	"time"
	"vibe-runner-server/game"

	"github.com/gorilla/websocket"
)

// playerIDCounter is a thread-safe counter for assigning unique player IDs.
// Each new player receives an incrementing ID starting from 1.
var (
	playerIDCounter int
	counterMutex    sync.Mutex
)

// getNextPlayerID generates the next unique player ID in a thread-safe manner.
// Multiple goroutines can call this concurrently without race conditions.
//
// Returns:
//   - int: The next available player ID (starting at 1)
func getNextPlayerID() int {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	playerIDCounter++
	return playerIDCounter
}

// sanitizePlayerName cleans and validates a player name to prevent XSS attacks.
// It performs the following operations:
//  1. Trims leading/trailing whitespace
//  2. Limits length to 30 characters
//  3. Escapes HTML entities (prevents <script> injection)
//  4. Provides default name if empty
//
// Parameters:
//   - name: The raw player name from client input
//
// Returns:
//   - string: Sanitized player name safe for display
func sanitizePlayerName(name string) string {
	// Remove leading/trailing whitespace
	name = strings.TrimSpace(name)

	// Limit to 30 characters
	if len(name) > 30 {
		name = name[:30]
	}

	// Escape HTML entities to prevent XSS
	name = html.EscapeString(name)

	// Provide default if empty after sanitization
	if name == "" {
		name = "Player"
	}

	return name
}

// HandleClient manages the WebSocket connection lifecycle for a single client.
// It handles message parsing, event routing, player state management, broadcasting, and cleanup.
//
// The function runs in its own goroutine (one per connected client).
// It blocks until the client disconnects or an error occurs.
//
// Parameters:
//   - conn: The WebSocket connection to manage
//   - gameState: The shared game state for adding/removing players
//   - clientHub: The client hub for registering this connection for broadcasts
//
// The function performs these steps:
//  1. Waits for join message
//  2. Creates player and adds to game state
//  3. Registers client with hub for state broadcasts
//  4. Assigns player ID and sends welcome
//  5. Enters message handling loop
//  6. Removes player from game state and hub on disconnect
func HandleClient(conn *websocket.Conn, gameState *game.GameState, clientHub *ClientHub) {
	// Player ID will be assigned after join message
	var playerID int
	var playerName string

	defer func() {
		// Remove player from game state and client hub on disconnect
		if playerID != 0 {
			gameState.RemovePlayer(playerID)
			clientHub.RemoveClient(playerID)
			log.Printf("Player removed from game state: ID=%d, Name=%s, Active players: %d",
				playerID, playerName, gameState.GetPlayerCount())
		}
		conn.Close()
		log.Printf("Client disconnected: %s", conn.RemoteAddr())
	}()

	log.Printf("Client connected: %s", conn.RemoteAddr())

	// Message handling loop
	for {
		// Read message from client
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			// Connection closed or error occurred
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error from %s: %v", conn.RemoteAddr(), err)
			}
			break
		}

		// Parse base message structure
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message from %s: %v", conn.RemoteAddr(), err)
			// Send error response (optional - could skip to ignore malformed messages)
			continue
		}

		// Route message based on event type
		switch msg.E {
		case "join":
			// Handle join event
			playerID, playerName, err = handleJoin(conn, msg, gameState)
			if err != nil {
				log.Printf("Join failed for %s: %v", conn.RemoteAddr(), err)
				return // Close connection on join failure
			}

			// Register client with hub for state broadcasts
			clientHub.AddClient(playerID, conn)

			log.Printf("Player joined: ID=%d, Name=%s, Position=(%.1f, %.1f), Active players: %d",
				playerID, playerName, 100.0, 440.0, gameState.GetPlayerCount())

		case "jump":
			// Handle jump event - apply jump to player in game state
			if playerID != 0 {
				player := gameState.GetPlayer(playerID)
				if player != nil {
					player.Jump()
					log.Printf("Player %d (%s) jumped", playerID, playerName)
				}
			}

		default:
			// Unknown event type
			log.Printf("Unknown event '%s' from player %d (%s)", msg.E, playerID, playerName)
		}
	}
}

// handleJoin processes a join request from a newly connected client.
// It validates the join message, creates a player entity, adds it to game state,
// assigns a player ID, and sends the welcome response.
//
// Parameters:
//   - conn: The WebSocket connection to send welcome message on
//   - msg: The parsed base message containing join data
//   - gameState: The game state to add the new player to
//
// Returns:
//   - int: Assigned player ID
//   - string: Sanitized player name
//   - error: Non-nil if join processing failed
func handleJoin(conn *websocket.Conn, msg Message, gameState *game.GameState) (int, string, error) {
	// Parse join-specific data
	joinDataBytes, err := json.Marshal(msg.D)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal join data: %w", err)
	}

	var joinMsg JoinMessage
	if err := json.Unmarshal(joinDataBytes, &joinMsg); err != nil {
		return 0, "", fmt.Errorf("failed to parse join message: %w", err)
	}

	// Sanitize player name
	playerName := sanitizePlayerName(joinMsg.N)

	// Assign unique player ID
	playerID := getNextPlayerID()

	// Create new player entity at spawn position (100, 440)
	player := game.NewPlayer(playerID, playerName)

	// Add player to game state
	gameState.AddPlayer(player)

	// Generate master seed (for now, simple seed - will be improved in Chunk 4)
	seed := fmt.Sprintf("vibe-runner-%d", playerID)

	// Get current server time in milliseconds
	serverTime := time.Now().UnixMilli()

	// Create welcome message
	welcomeData := WelcomeMessage{
		ID:         playerID,
		Seed:       seed,
		ServerTime: serverTime,
	}

	welcomeMsg := Message{
		E: "welcome",
		D: welcomeData,
	}

	// Send welcome message
	if err := sendMessage(conn, welcomeMsg); err != nil {
		return 0, "", fmt.Errorf("failed to send welcome message: %w", err)
	}

	return playerID, playerName, nil
}

// sendMessage sends a message to a client over the WebSocket connection.
// It marshals the message to JSON and writes it to the connection.
//
// Parameters:
//   - conn: The WebSocket connection to send on
//   - msg: The message to send (will be JSON-encoded)
//
// Returns:
//   - error: Non-nil if sending failed
func sendMessage(conn *websocket.Conn, msg Message) error {
	// Marshal message to JSON
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write to WebSocket connection
	if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
