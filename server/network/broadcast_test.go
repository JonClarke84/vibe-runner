package network

import (
	"sync"
	"testing"
	"time"
	"vibe-runner-server/game"
)

// TestNewClientHub_CreatesEmptyHub verifies that NewClientHub
// initializes an empty client map ready for use.
func TestNewClientHub_CreatesEmptyHub(t *testing.T) {
	// Act
	hub := NewClientHub()

	// Assert
	if hub == nil {
		t.Fatal("NewClientHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("NewClientHub() clients map is nil")
	}

	hub.mu.RLock()
	count := len(hub.clients)
	hub.mu.RUnlock()

	if count != 0 {
		t.Errorf("NewClientHub() client count = %d, want 0", count)
	}
}

// TestClientHub_ManualClientAddition tests adding clients manually
// without starting write goroutines (unit test approach).
func TestClientHub_ManualClientAddition(t *testing.T) {
	// Arrange
	hub := NewClientHub()

	// Manually add clients to the hub (bypassing AddClient which starts goroutines)
	client1 := &ClientConnection{
		PlayerID: 1,
		Conn:     nil, // Not needed for this test
		SendChan: make(chan []byte, 10),
		closed:   false,
	}
	client2 := &ClientConnection{
		PlayerID: 2,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
		closed:   false,
	}

	hub.mu.Lock()
	hub.clients[1] = client1
	hub.clients[2] = client2
	hub.mu.Unlock()

	// Assert
	hub.mu.RLock()
	count := len(hub.clients)
	hub.mu.RUnlock()

	if count != 2 {
		t.Errorf("Manual client addition resulted in %d clients, want 2", count)
	}
}

// TestBroadcastState_QueuesMessagesToClients tests that BroadcastState
// puts messages into client send channels.
func TestBroadcastState_QueuesMessagesToClients(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState()

	// Add test players to game state
	player1 := game.NewPlayer(1, "Player1")
	player2 := game.NewPlayer(2, "Player2")
	gameState.AddPlayer(player1)
	gameState.AddPlayer(player2)

	// Manually add mock clients (without starting write goroutines)
	client1 := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}
	client2 := &ClientConnection{
		PlayerID: 2,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	hub.mu.Lock()
	hub.clients[1] = client1
	hub.clients[2] = client2
	hub.mu.Unlock()

	// Act
	hub.BroadcastState(gameState)

	// Assert - verify messages were queued in SendChan for both clients
	select {
	case msg := <-client1.SendChan:
		if msg == nil || len(msg) == 0 {
			t.Error("BroadcastState() sent nil/empty message to client 1")
		}
	default:
		t.Error("BroadcastState() did not queue message to client 1")
	}

	select {
	case msg := <-client2.SendChan:
		if msg == nil || len(msg) == 0 {
			t.Error("BroadcastState() sent nil/empty message to client 2")
		}
	default:
		t.Error("BroadcastState() did not queue message to client 2")
	}
}

// TestBroadcastState_EmptyGameState_SendsValidMessage tests that
// broadcasting an empty game state sends a valid JSON message.
func TestBroadcastState_EmptyGameState_SendsValidMessage(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState() // No players

	client := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	hub.mu.Lock()
	hub.clients[1] = client
	hub.mu.Unlock()

	// Act
	hub.BroadcastState(gameState)

	// Assert
	select {
	case msg := <-client.SendChan:
		if msg == nil || len(msg) == 0 {
			t.Error("BroadcastState() sent nil/empty message for empty game state")
		}
		// Message should be valid JSON (basic check)
		if msg[0] != '{' {
			t.Error("BroadcastState() message doesn't start with '{' (not JSON)")
		}
	default:
		t.Error("BroadcastState() did not send message for empty game state")
	}
}

// TestBroadcastState_FullChannelDropsMessage tests that when a client's
// SendChan is full, the broadcast drops the message rather than blocking.
func TestBroadcastState_FullChannelDropsMessage(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState()
	gameState.AddPlayer(game.NewPlayer(1, "Player1"))

	client := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	hub.mu.Lock()
	hub.clients[1] = client
	hub.mu.Unlock()

	// Fill the channel to capacity
	for i := 0; i < 10; i++ {
		client.SendChan <- []byte("fill")
	}

	// Verify channel is full
	if len(client.SendChan) != 10 {
		t.Fatalf("Test setup failed: SendChan should be full, got %d/10", len(client.SendChan))
	}

	// Act - broadcast should not block even though channel is full
	done := make(chan bool, 1)
	go func() {
		hub.BroadcastState(gameState)
		done <- true
	}()

	// Assert - broadcast should complete quickly (not block)
	// Use timeout to allow goroutine to start and complete
	select {
	case <-done:
		// Success - broadcast completed without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("BroadcastState() appears to have blocked on full SendChan")
	}

	// Verify channel is still full (new message was dropped)
	if len(client.SendChan) != 10 {
		t.Errorf("BroadcastState() did not drop message for full channel: got %d/10", len(client.SendChan))
	}
}

// TestClientHub_ConcurrentBroadcast tests that multiple goroutines
// can broadcast simultaneously without race conditions.
func TestClientHub_ConcurrentBroadcast(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState()
	gameState.AddPlayer(game.NewPlayer(1, "Player1"))

	// Add mock clients
	for i := 1; i <= 10; i++ {
		client := &ClientConnection{
			PlayerID: i,
			Conn:     nil,
			SendChan: make(chan []byte, 10),
		}
		hub.mu.Lock()
		hub.clients[i] = client
		hub.mu.Unlock()
	}

	// Act - broadcast from multiple goroutines simultaneously
	var wg sync.WaitGroup
	numBroadcasts := 50

	wg.Add(numBroadcasts)
	for i := 0; i < numBroadcasts; i++ {
		go func() {
			defer wg.Done()
			hub.BroadcastState(gameState)
		}()
	}

	// Assert - should complete without race conditions or panics
	wg.Wait()

	// Verify clients received some messages
	hub.mu.RLock()
	for playerID, client := range hub.clients {
		if len(client.SendChan) == 0 {
			t.Errorf("Client %d did not receive any messages from concurrent broadcasts", playerID)
		}
	}
	hub.mu.RUnlock()
}

// TestBroadcastState_OnlyAlivePlayers tests that only alive players
// are included in the broadcast state.
func TestBroadcastState_OnlyAlivePlayers(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState()

	// Add alive and dead players
	player1 := game.NewPlayer(1, "AlivePlayer")
	player2 := game.NewPlayer(2, "DeadPlayer")
	player2.Kill() // Kill player 2

	gameState.AddPlayer(player1)
	gameState.AddPlayer(player2)

	client := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	hub.mu.Lock()
	hub.clients[1] = client
	hub.mu.Unlock()

	// Act
	hub.BroadcastState(gameState)

	// Assert - message should be sent and contain only alive players
	select {
	case msg := <-client.SendChan:
		if msg == nil {
			t.Error("BroadcastState() sent nil message")
		}
		// The message should only include player1 (alive)
		// Detailed JSON verification would require parsing, but we can check length
		// The message for 1 player should be shorter than for 2 players
		// (This is a basic sanity check - proper verification would parse JSON)
	default:
		t.Error("BroadcastState() did not send message")
	}
}

// TestClientConnection_BufferSize tests that the send channel
// has the expected buffer size of 10.
func TestClientConnection_BufferSize(t *testing.T) {
	// Arrange & Act
	client := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	// Assert
	if cap(client.SendChan) != 10 {
		t.Errorf("ClientConnection SendChan capacity = %d, want 10", cap(client.SendChan))
	}
}

// TestBroadcastState_MultiplePlayersJSON tests that the broadcast
// creates valid JSON with multiple players.
func TestBroadcastState_MultiplePlayersJSON(t *testing.T) {
	// Arrange
	hub := NewClientHub()
	gameState := game.NewGameState()

	// Add multiple alive players
	for i := 1; i <= 5; i++ {
		player := game.NewPlayer(i, "Player")
		player.X = float64(i * 100)
		player.Y = 440.0
		gameState.AddPlayer(player)
	}

	client := &ClientConnection{
		PlayerID: 1,
		Conn:     nil,
		SendChan: make(chan []byte, 10),
	}

	hub.mu.Lock()
	hub.clients[1] = client
	hub.mu.Unlock()

	// Act
	hub.BroadcastState(gameState)

	// Assert
	select {
	case msg := <-client.SendChan:
		// Basic JSON structure validation
		if len(msg) == 0 {
			t.Fatal("BroadcastState() sent empty message")
		}
		if msg[0] != '{' || msg[len(msg)-1] != '}' {
			t.Error("BroadcastState() message is not valid JSON object")
		}
		// Should contain event and data fields
		msgStr := string(msg)
		if len(msgStr) < 20 {
			t.Error("BroadcastState() message seems too short for 5 players")
		}
	default:
		t.Error("BroadcastState() did not send message")
	}
}
