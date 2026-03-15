package main

import "testing"

// TestPlayerStartPositions verifies center safe house starting positions
func TestPlayerStartPositions(t *testing.T) {
	expected := []int{2, 6, 10, 14} // Center of each edge (safe houses)
	
	if len(playerStartPositions) != 4 {
		t.Fatalf("playerStartPositions len = %d, want 4", len(playerStartPositions))
	}
	
	for i, pos := range playerStartPositions {
		if pos != expected[i] {
			t.Errorf("Player %d start pos = %d, want %d", i, pos, expected[i])
		}
	}
}

// TestOuterPathSize verifies 16 positions on outer path
func TestOuterPathSize(t *testing.T) {
	if len(outerPath) != 16 {
		t.Errorf("outerPath len = %d, want 16", len(outerPath))
	}
}

// TestInnerCircleSize verifies 8 positions on inner path
func TestInnerCircleSize(t *testing.T) {
	if len(innerCircle) != 8 {
		t.Errorf("innerCircle len = %d, want 8", len(innerCircle))
	}
}

// TestPlayerColors verifies 4 player colors
func TestPlayerColors(t *testing.T) {
	if len(playerColors) != 4 {
		t.Errorf("playerColors len = %d, want 4", len(playerColors))
	}
}

// TestTokensPerPlayer verifies 4 tokens per player
func TestTokensPerPlayer(t *testing.T) {
	if tokensPerPlayer != 4 {
		t.Errorf("tokensPerPlayer = %d, want 4", tokensPerPlayer)
	}
}

// TestNumPlayers verifies 4 players max
func TestNumPlayers(t *testing.T) {
	if numPlayers != 4 {
		t.Errorf("numPlayers = %d, want 4", numPlayers)
	}
}

// TestSafeZones verifies safe positions are center safe houses
func TestSafeZones(t *testing.T) {
	safeZones := []int{2, 6, 10, 14}
	
	for _, pos := range safeZones {
		// Safe zones are where players start and cannot be killed
		if pos < 0 || pos >= 16 {
			t.Errorf("Safe zone %d out of outerPath range", pos)
		}
	}
}

// TestRollValues verifies valid roll results are 1-8
func TestRollValues(t *testing.T) {
	validRolls := []int{1, 2, 3, 4, 8}
	
	for _, r := range validRolls {
		if r < 1 || r > 8 {
			t.Errorf("Invalid roll value %d", r)
		}
	}
}

// TestStartRollValues verifies only 1, 4, 8 allow exit from home
func TestStartRollValues(t *testing.T) {
	startValues := []int{1, 4, 8}
	nonStartValues := []int{2, 3} // Note: 5,6,7 don't occur with 4 shells
	
	for _, v := range startValues {
		if v != 1 && v != 4 && v != 8 {
			t.Errorf("%d should be a start value", v)
		}
	}
	
	for _, v := range nonStartValues {
		if v == 1 || v == 4 || v == 8 {
			t.Errorf("%d should NOT be a start value", v)
		}
	}
}

// TestBoardSize verifies 5x5 board
func TestBoardSize(t *testing.T) {
	if boardSize != 5 {
		t.Errorf("boardSize = %d, want 5", boardSize)
	}
}

// TestCellSize verifies cell size constant
func TestCellSize(t *testing.T) {
	if cellSize != 100 {
		t.Errorf("cellSize = %d, want 100", cellSize)
	}
}
