package game

import "testing"

// TestNewPlayer_CreatesPlayerWithDefaultValues verifies that NewPlayer
// initializes a player with correct spawn position and default state.
func TestNewPlayer_CreatesPlayerWithDefaultValues(t *testing.T) {
	// Arrange
	id := 123
	name := "TestPlayer"

	// Act
	player := NewPlayer(id, name)

	// Assert
	if player.ID != id {
		t.Errorf("NewPlayer() ID = %d, want %d", player.ID, id)
	}
	if player.Name != name {
		t.Errorf("NewPlayer() Name = %s, want %s", player.Name, name)
	}
	if player.X != 100.0 {
		t.Errorf("NewPlayer() X = %.1f, want 100.0", player.X)
	}
	if player.Y != 440.0 {
		t.Errorf("NewPlayer() Y = %.1f, want 440.0", player.Y)
	}
	if player.VelocityY != 0.0 {
		t.Errorf("NewPlayer() VelocityY = %.1f, want 0.0", player.VelocityY)
	}
	if !player.IsGrounded {
		t.Error("NewPlayer() IsGrounded = false, want true")
	}
	if !player.IsAlive {
		t.Error("NewPlayer() IsAlive = false, want true")
	}
}

// TestJump_WhenGroundedAndAlive_AppliesJumpVelocity tests that Jump()
// correctly applies upward velocity when the player is on the ground.
func TestJump_WhenGroundedAndAlive_AppliesJumpVelocity(t *testing.T) {
	// Arrange
	player := NewPlayer(1, "TestPlayer")

	// Verify preconditions
	if !player.IsGrounded {
		t.Fatal("Test setup failed: player should start grounded")
	}
	if !player.IsAlive {
		t.Fatal("Test setup failed: player should start alive")
	}

	// Act
	player.Jump()

	// Assert
	if player.VelocityY != -600.0 {
		t.Errorf("Jump() VelocityY = %.1f, want -600.0", player.VelocityY)
	}
	if player.IsGrounded {
		t.Error("Jump() IsGrounded = true, want false after jumping")
	}
}

// TestJump_WhenNotGrounded_DoesNothing tests that Jump() is a no-op
// when the player is already in the air (prevents double-jumping).
func TestJump_WhenNotGrounded_DoesNothing(t *testing.T) {
	// Arrange
	player := NewPlayer(1, "TestPlayer")
	player.IsGrounded = false
	player.VelocityY = -300.0 // Already moving upward

	initialVelocity := player.VelocityY

	// Act
	player.Jump()

	// Assert - velocity should not change
	if player.VelocityY != initialVelocity {
		t.Errorf("Jump() changed velocity while airborne: got %.1f, want %.1f",
			player.VelocityY, initialVelocity)
	}
	if player.IsGrounded {
		t.Error("Jump() should not set IsGrounded to true when airborne")
	}
}

// TestJump_WhenDead_DoesNothing tests that dead players cannot jump.
func TestJump_WhenDead_DoesNothing(t *testing.T) {
	// Arrange
	player := NewPlayer(1, "TestPlayer")
	player.IsAlive = false
	player.IsGrounded = true

	initialVelocity := player.VelocityY

	// Act
	player.Jump()

	// Assert - velocity should not change
	if player.VelocityY != initialVelocity {
		t.Errorf("Jump() changed velocity while dead: got %.1f, want %.1f",
			player.VelocityY, initialVelocity)
	}
}

// TestKill_SetsIsAliveToFalse tests that Kill() marks the player as dead.
func TestKill_SetsIsAliveToFalse(t *testing.T) {
	// Arrange
	player := NewPlayer(1, "TestPlayer")

	// Verify precondition
	if !player.IsAlive {
		t.Fatal("Test setup failed: player should start alive")
	}

	// Act
	player.Kill()

	// Assert
	if player.IsAlive {
		t.Error("Kill() failed to set IsAlive to false")
	}
}

// TestJump_TableDriven tests Jump() with various player states using
// table-driven test pattern for comprehensive coverage.
func TestJump_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		isGrounded     bool
		isAlive        bool
		initialVelocity float64
		wantVelocity   float64
		wantGrounded   bool
	}{
		{
			name:           "grounded and alive - should jump",
			isGrounded:     true,
			isAlive:        true,
			initialVelocity: 0.0,
			wantVelocity:   -600.0,
			wantGrounded:   false,
		},
		{
			name:           "airborne - should not jump",
			isGrounded:     false,
			isAlive:        true,
			initialVelocity: -300.0,
			wantVelocity:   -300.0, // Unchanged
			wantGrounded:   false,
		},
		{
			name:           "dead but grounded - should not jump",
			isGrounded:     true,
			isAlive:        false,
			initialVelocity: 0.0,
			wantVelocity:   0.0, // Unchanged
			wantGrounded:   true,
		},
		{
			name:           "dead and airborne - should not jump",
			isGrounded:     false,
			isAlive:        false,
			initialVelocity: 100.0,
			wantVelocity:   100.0, // Unchanged
			wantGrounded:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			player := NewPlayer(1, "TestPlayer")
			player.IsGrounded = tt.isGrounded
			player.IsAlive = tt.isAlive
			player.VelocityY = tt.initialVelocity

			// Act
			player.Jump()

			// Assert
			if player.VelocityY != tt.wantVelocity {
				t.Errorf("Jump() VelocityY = %.1f, want %.1f",
					player.VelocityY, tt.wantVelocity)
			}
			if player.IsGrounded != tt.wantGrounded {
				t.Errorf("Jump() IsGrounded = %v, want %v",
					player.IsGrounded, tt.wantGrounded)
			}
		})
	}
}
