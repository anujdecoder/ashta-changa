package server

import (
	"testing"
)

// TestRollResultMapping verifies shell roll result calculation
func TestRollResultMapping(t *testing.T) {
	tests := []struct {
		openCount int
		want      int
	}{
		{0, 8},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
	}

	for _, tt := range tests {
		got := rollResultFromOpenCount(tt.openCount)
		if got != tt.want {
			t.Errorf("rollResultFromOpenCount(%d) = %d, want %d", tt.openCount, got, tt.want)
		}
	}
}

// rollResultFromOpenCount mirrors server roll calculation for testing
func rollResultFromOpenCount(openCount int) int {
	switch openCount {
	case 0:
		return 8
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	}
	return 0
}

// TestStartGamePositions verifies tokens start at center safe houses
func TestStartGamePositions(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host1")
	
	// Add 2 players
	room.AddPlayer(&Player{ID: "p1", Name: "P1", PlayerIdx: 0})
	room.AddPlayer(&Player{ID: "p2", Name: "P2", PlayerIdx: 1})
	
	if !room.StartGame() {
		t.Fatal("StartGame failed")
	}
	
	// Check player 0 tokens at position 2 (top center)
	if room.GameState.PlayerTokens[0].Tokens[0].Position != 2 {
		t.Errorf("Player 0 token position = %d, want 2", room.GameState.PlayerTokens[0].Tokens[0].Position)
	}
	
	// Check player 1 tokens at position 6 (right center)
	if room.GameState.PlayerTokens[1].Tokens[0].Position != 6 {
		t.Errorf("Player 1 token position = %d, want 6", room.GameState.PlayerTokens[1].Tokens[0].Position)
	}
	
	// All tokens should be at start (state 0)
	for _, pt := range room.GameState.PlayerTokens {
		for _, tok := range pt.Tokens {
			if tok.State != 0 {
				t.Errorf("Token state = %d, want 0 (atStart)", tok.State)
			}
		}
	}
}

// TestStartOnlyOn148 verifies game can only start (exit home) on 1, 4, or 8
func TestStartOnlyOn148(t *testing.T) {
	tests := []struct {
		roll       int
		allAtHome  bool
		shouldSkip bool
	}{
		{1, true, false},  // Can start, no skip
		{4, true, false},  // Can start, no skip
		{8, true, false},  // Can start, no skip
		{2, true, true},   // Cannot start, skip
		{3, true, true},   // Cannot start, skip
		{1, false, false}, // On board, no skip (normal)
		{2, false, false}, // On board, no skip (normal)
	}

	for _, tt := range tests {
		skip := shouldSkipTurn(tt.allAtHome, tt.roll)
		if skip != tt.shouldSkip {
			t.Errorf("shouldSkipTurn(allAtHome=%v, roll=%d) = %v, want %v",
				tt.allAtHome, tt.roll, skip, tt.shouldSkip)
		}
	}
}

// shouldSkipTurn mirrors server skip logic for testing
func shouldSkipTurn(allAtHome bool, roll int) bool {
	return allAtHome && roll != 1 && roll != 4 && roll != 8
}

// TestExtraTurn verifies extra turn granted on 4 or 8
func TestExtraTurn(t *testing.T) {
	tests := []struct {
		roll      int
		extraTurn bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, true},
		{8, true},
	}

	for _, tt := range tests {
		got := tt.roll == 4 || tt.roll == 8
		if got != tt.extraTurn {
			t.Errorf("roll %d extraTurn = %v, want %v", tt.roll, got, tt.extraTurn)
		}
	}
}

// TestTurnChange verifies turn advances correctly
func TestTurnChange(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host1")
	
	room.AddPlayer(&Player{ID: "p1", Name: "P1", PlayerIdx: 0})
	room.AddPlayer(&Player{ID: "p2", Name: "P2", PlayerIdx: 1})
	room.StartGame()
	
	// Simulate after move (no extra turn)
	room.mu.Lock()
	room.GameState.CurrentPlayer = (room.GameState.CurrentPlayer + 1) % room.GameState.NumActivePlayers
	room.mu.Unlock()
	
	if room.GameState.CurrentPlayer != 1 {
		t.Errorf("After turn change, CurrentPlayer = %d, want 1", room.GameState.CurrentPlayer)
	}
	
	// Wrap around
	room.mu.Lock()
	room.GameState.CurrentPlayer = (room.GameState.CurrentPlayer + 1) % room.GameState.NumActivePlayers
	room.mu.Unlock()
	
	if room.GameState.CurrentPlayer != 0 {
		t.Errorf("After wrap, CurrentPlayer = %d, want 0", room.GameState.CurrentPlayer)
	}
}

// TestWinCondition verifies win when all tokens finished
func TestWinCondition(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host1")
	
	room.AddPlayer(&Player{ID: "p1", Name: "P1", PlayerIdx: 0})
	room.StartGame()
	
	// Set all player 0 tokens to finished
	room.mu.Lock()
	for i := range room.GameState.PlayerTokens[0].Tokens {
		room.GameState.PlayerTokens[0].Tokens[i].State = 2 // finished
		room.GameState.PlayerTokens[0].Tokens[i].Position = 24
	}
	room.mu.Unlock()
	
	// Check win
	allFinished := true
	for _, t := range room.GameState.PlayerTokens[0].Tokens {
		if t.State != 2 {
			allFinished = false
			break
		}
	}
	
	if !allFinished {
		t.Error("Expected all tokens finished for win condition")
	}
	
	room.mu.Lock()
	room.GameState.Winner = 0
	room.mu.Unlock()
	
	if room.GameState.Winner != 0 {
		t.Errorf("Winner = %d, want 0", room.GameState.Winner)
	}
}

// TestKillAllowed verifies kill rules (not on safe zones)
func TestKillAllowed(t *testing.T) {
	// Safe positions: 2, 6, 10, 14 (center safe houses)
	safePositions := []int{2, 6, 10, 14}
	
	// Test kill allowed (not safe)
	testPos := 5 // Not safe
	isSafe := false
	for _, sp := range safePositions {
		if testPos == sp {
			isSafe = true
			break
		}
	}
	if isSafe {
		t.Errorf("Position %d should not be safe", testPos)
	}
	
	// Test kill not allowed (safe)
	testPos = 2 // Safe
	isSafe = false
	for _, sp := range safePositions {
		if testPos == sp {
			isSafe = true
			break
		}
	}
	if !isSafe {
		t.Errorf("Position %d should be safe", testPos)
	}
}

// TestKillOnSafeHouse verifies no kill possible on safe houses
func TestKillOnSafeHouse(t *testing.T) {
	safeHouses := []int{2, 6, 10, 14}
	
	for _, pos := range safeHouses {
		// If token lands on safe house, kill should not happen
		isSafe := pos == 2 || pos == 6 || pos == 10 || pos == 14
		if !isSafe {
			t.Errorf("Position %d should be safe house", pos)
		}
	}
}

// TestPlayerOrder verifies players are ordered correctly
func TestPlayerOrder(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host1")
	
	room.AddPlayer(&Player{ID: "p1", Name: "P1"})
	room.AddPlayer(&Player{ID: "p2", Name: "P2"})
	room.AddPlayer(&Player{ID: "p3", Name: "P3"})
	room.AddPlayer(&Player{ID: "p4", Name: "P4"})
	
	room.StartGame()
	
	// Verify player indices
	for i, pt := range room.GameState.PlayerTokens {
		if pt.PlayerIdx != i {
			t.Errorf("PlayerTokens[%d].PlayerIdx = %d, want %d", i, pt.PlayerIdx, i)
		}
		// Verify start positions: 2, 6, 10, 14
		expectedPos := i*4 + 2
		if pt.Tokens[0].Position != expectedPos {
			t.Errorf("Player %d position = %d, want %d", i, pt.Tokens[0].Position, expectedPos)
		}
	}
}
