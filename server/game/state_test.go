package game

import (
	"sync"
	"testing"
)

// TestNewGameState_CreatesEmptyState verifies that NewGameState
// initializes an empty player map ready for use.
func TestNewGameState_CreatesEmptyState(t *testing.T) {
	// Act
	state := NewGameState()

	// Assert
	if state == nil {
		t.Fatal("NewGameState() returned nil")
	}
	if state.players == nil {
		t.Error("NewGameState() players map is nil")
	}
	if len(state.players) != 0 {
		t.Errorf("NewGameState() player count = %d, want 0", len(state.players))
	}
}

// TestAddPlayer_AddsPlayerToState tests that AddPlayer correctly
// stores a player and makes it retrievable.
func TestAddPlayer_AddsPlayerToState(t *testing.T) {
	// Arrange
	state := NewGameState()
	player := NewPlayer(1, "TestPlayer")

	// Act
	state.AddPlayer(player)

	// Assert
	count := state.GetPlayerCount()
	if count != 1 {
		t.Errorf("AddPlayer() player count = %d, want 1", count)
	}

	retrieved := state.GetPlayer(1)
	if retrieved == nil {
		t.Fatal("AddPlayer() player not retrievable after adding")
	}
	if retrieved.ID != player.ID {
		t.Errorf("AddPlayer() retrieved player ID = %d, want %d", retrieved.ID, player.ID)
	}
	if retrieved.Name != player.Name {
		t.Errorf("AddPlayer() retrieved player Name = %s, want %s", retrieved.Name, player.Name)
	}
}

// TestAddPlayer_MultiplePlayers tests adding multiple players to the state.
func TestAddPlayer_MultiplePlayers(t *testing.T) {
	// Arrange
	state := NewGameState()
	player1 := NewPlayer(1, "Player1")
	player2 := NewPlayer(2, "Player2")
	player3 := NewPlayer(3, "Player3")

	// Act
	state.AddPlayer(player1)
	state.AddPlayer(player2)
	state.AddPlayer(player3)

	// Assert
	count := state.GetPlayerCount()
	if count != 3 {
		t.Errorf("AddPlayer() player count = %d, want 3", count)
	}

	// Verify each player is retrievable
	for i := 1; i <= 3; i++ {
		player := state.GetPlayer(i)
		if player == nil {
			t.Errorf("GetPlayer(%d) returned nil", i)
		}
	}
}

// TestGetPlayer_NonExistentID_ReturnsNil tests that GetPlayer returns
// nil for player IDs that don't exist in the state.
func TestGetPlayer_NonExistentID_ReturnsNil(t *testing.T) {
	// Arrange
	state := NewGameState()
	state.AddPlayer(NewPlayer(1, "Player1"))

	// Act
	player := state.GetPlayer(999)

	// Assert
	if player != nil {
		t.Error("GetPlayer() returned non-nil for non-existent player")
	}
}

// TestRemovePlayer_RemovesPlayerFromState tests that RemovePlayer
// correctly removes a player and decrements the count.
func TestRemovePlayer_RemovesPlayerFromState(t *testing.T) {
	// Arrange
	state := NewGameState()
	player := NewPlayer(1, "TestPlayer")
	state.AddPlayer(player)

	// Verify player exists before removal
	if state.GetPlayer(1) == nil {
		t.Fatal("Test setup failed: player not added")
	}

	// Act
	state.RemovePlayer(1)

	// Assert
	count := state.GetPlayerCount()
	if count != 0 {
		t.Errorf("RemovePlayer() player count = %d, want 0", count)
	}

	retrieved := state.GetPlayer(1)
	if retrieved != nil {
		t.Error("RemovePlayer() player still retrievable after removal")
	}
}

// TestRemovePlayer_NonExistentID_DoesNotPanic tests that removing
// a non-existent player doesn't cause errors.
func TestRemovePlayer_NonExistentID_DoesNotPanic(t *testing.T) {
	// Arrange
	state := NewGameState()

	// Act & Assert - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RemovePlayer() panicked with non-existent ID: %v", r)
		}
	}()

	state.RemovePlayer(999)
}

// TestGetAllPlayers_ReturnsAllPlayers tests that GetAllPlayers returns
// a slice containing all players in the state.
func TestGetAllPlayers_ReturnsAllPlayers(t *testing.T) {
	// Arrange
	state := NewGameState()
	player1 := NewPlayer(1, "Player1")
	player2 := NewPlayer(2, "Player2")
	player3 := NewPlayer(3, "Player3")

	state.AddPlayer(player1)
	state.AddPlayer(player2)
	state.AddPlayer(player3)

	// Act
	players := state.GetAllPlayers()

	// Assert
	if len(players) != 3 {
		t.Errorf("GetAllPlayers() returned %d players, want 3", len(players))
	}

	// Verify all players are present (order doesn't matter)
	foundIDs := make(map[int]bool)
	for _, p := range players {
		foundIDs[p.ID] = true
	}

	for i := 1; i <= 3; i++ {
		if !foundIDs[i] {
			t.Errorf("GetAllPlayers() missing player ID %d", i)
		}
	}
}

// TestGetAllPlayers_EmptyState_ReturnsEmptySlice tests that GetAllPlayers
// returns an empty slice (not nil) when no players exist.
func TestGetAllPlayers_EmptyState_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	state := NewGameState()

	// Act
	players := state.GetAllPlayers()

	// Assert
	if players == nil {
		t.Error("GetAllPlayers() returned nil, want empty slice")
	}
	if len(players) != 0 {
		t.Errorf("GetAllPlayers() returned %d players, want 0", len(players))
	}
}

// TestGetPlayerCount_ReflectsActualCount tests that GetPlayerCount
// accurately reflects the number of players.
func TestGetPlayerCount_ReflectsActualCount(t *testing.T) {
	// Arrange
	state := NewGameState()

	// Test empty state
	if count := state.GetPlayerCount(); count != 0 {
		t.Errorf("GetPlayerCount() = %d, want 0 for empty state", count)
	}

	// Add players and verify count increases
	state.AddPlayer(NewPlayer(1, "Player1"))
	if count := state.GetPlayerCount(); count != 1 {
		t.Errorf("GetPlayerCount() = %d, want 1", count)
	}

	state.AddPlayer(NewPlayer(2, "Player2"))
	if count := state.GetPlayerCount(); count != 2 {
		t.Errorf("GetPlayerCount() = %d, want 2", count)
	}

	// Remove player and verify count decreases
	state.RemovePlayer(1)
	if count := state.GetPlayerCount(); count != 1 {
		t.Errorf("GetPlayerCount() = %d, want 1 after removal", count)
	}
}

// TestGameState_ConcurrentAccess tests that GameState is thread-safe
// when accessed by multiple goroutines simultaneously.
// This test verifies that RWMutex correctly prevents race conditions.
func TestGameState_ConcurrentAccess(t *testing.T) {
	// Arrange
	state := NewGameState()
	numGoroutines := 100
	var wg sync.WaitGroup

	// Act - spawn many goroutines that add, read, and remove players concurrently
	wg.Add(numGoroutines * 3)

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			player := NewPlayer(id, "Player")
			state.AddPlayer(player)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_ = state.GetPlayer(id)
			_ = state.GetAllPlayers()
			_ = state.GetPlayerCount()
		}(i)
	}

	// Concurrent removes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			state.RemovePlayer(id)
		}(i)
	}

	// Assert - should complete without race conditions or panics
	wg.Wait()

	// Final state should be valid (exact count depends on goroutine timing)
	count := state.GetPlayerCount()
	if count < 0 {
		t.Errorf("Concurrent access resulted in negative player count: %d", count)
	}
}

// TestGameState_GetAllPlayers_ReturnsSnapshot tests that GetAllPlayers
// returns a snapshot of players that won't be affected by concurrent modifications.
func TestGameState_GetAllPlayers_ReturnsSnapshot(t *testing.T) {
	// Arrange
	state := NewGameState()
	state.AddPlayer(NewPlayer(1, "Player1"))
	state.AddPlayer(NewPlayer(2, "Player2"))

	// Act - get snapshot
	snapshot := state.GetAllPlayers()
	initialLen := len(snapshot)

	// Modify state after getting snapshot
	state.AddPlayer(NewPlayer(3, "Player3"))
	state.RemovePlayer(1)

	// Assert - snapshot should be unchanged
	if len(snapshot) != initialLen {
		t.Errorf("GetAllPlayers() snapshot length changed: got %d, want %d",
			len(snapshot), initialLen)
	}
}
