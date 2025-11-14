package main

import (
	"log"
	"net/http"
	"vibe-runner-server/game"
	"vibe-runner-server/network"

	"github.com/gorilla/websocket"
)

// upgrader configures the WebSocket connection upgrade from HTTP.
// It sets buffer sizes for read/write operations and allows connections
// from any origin (CORS). In production, CheckOrigin should validate
// the origin to prevent unauthorized connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development. In production, implement proper origin checking.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// makeWebSocketHandler creates a WebSocket upgrade handler with access to game state and client hub.
// This returns a closure that captures the game state and client hub for use in HandleClient.
//
// Parameters:
//   - gameState: The shared game state for player management
//   - clientHub: The client hub for state broadcasting
//
// Returns:
//   - http.HandlerFunc: Handler function for WebSocket upgrades
func makeWebSocketHandler(gameState *game.GameState, clientHub *network.ClientHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket protocol
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		// Log successful connection establishment
		log.Printf("WebSocket upgrade: client connected from %s", r.RemoteAddr)

		// Delegate connection handling to network package
		// HandleClient manages message parsing, event routing, player state, and cleanup
		// This call blocks until the client disconnects
		network.HandleClient(conn, gameState, clientHub)
	}
}

// main initializes and starts the HTTP server with WebSocket support.
// It creates the game state, sets up routing for the WebSocket endpoint,
// and starts listening on port 8080.
//
// The server registers a single endpoint:
//   - /ws: WebSocket upgrade endpoint for game client connections
//
// The function blocks indefinitely, serving incoming HTTP requests.
// If the server fails to start, the application exits with a fatal error.
func main() {
	// Create game state (shared across all client connections)
	gameState := game.NewGameState()
	log.Printf("Game state initialized")

	// Create client hub for broadcasting state updates
	clientHub := network.NewClientHub()
	log.Printf("Client hub initialized")

	// Start game ticker (20Hz physics loop with state broadcasting)
	game.StartGameTicker(gameState, clientHub)
	log.Printf("Game ticker started")

	// Register WebSocket handler at /ws endpoint with game state and client hub
	http.HandleFunc("/ws", makeWebSocketHandler(gameState, clientHub))

	// Start HTTP server on port 8080
	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	log.Printf("WebSocket endpoint available at ws://localhost%s/ws", addr)

	// Start listening and serving requests
	// This blocks until the server encounters a fatal error
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
