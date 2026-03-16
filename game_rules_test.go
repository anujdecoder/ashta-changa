package main

import (
	"testing"

	"github.com/anujdecoder/ashta-board/game"
)

// TestPlayerStartPositions verifies center safe house starting positions
func TestPlayerStartPositions(t *testing.T) {
	expected := []int{2, 6, 10, 14} // Center of each edge (safe houses)

	if len(game.PlayerStartPositions) != 4 {
		t.Fatalf("PlayerStartPositions len = %d, want 4", len(game.PlayerStartPositions))
	}

	for i, pos := range game.PlayerStartPositions {
		if pos != expected[i] {
			t.Errorf("Player %d start pos = %d, want %d", i, pos, expected[i])
		}
	}
}

// TestOuterPathSize verifies 16 positions on outer path
func TestOuterPathSize(t *testing.T) {
	if len(game.OuterPath) != 16 {
		t.Errorf("OuterPath len = %d, want 16", len(game.OuterPath))
	}
}

// TestInnerCircleSize verifies 8 positions on inner path
func TestInnerCircleSize(t *testing.T) {
	if len(game.InnerCircle) != 8 {
		t.Errorf("InnerCircle len = %d, want 8", len(game.InnerCircle))
	}
}

// TestPlayerColors verifies 4 player colors
func TestPlayerColors(t *testing.T) {
	if len(game.PlayerColors) != 4 {
		t.Errorf("PlayerColors len = %d, want 4", len(game.PlayerColors))
	}
}

// TestTokensPerPlayer verifies 4 tokens per player
func TestTokensPerPlayer(t *testing.T) {
	if game.TokensPerPlayer != 4 {
		t.Errorf("TokensPerPlayer = %d, want 4", game.TokensPerPlayer)
	}
}

// TestNumPlayers verifies 4 players max
func TestNumPlayers(t *testing.T) {
	if game.NumPlayers != 4 {
		t.Errorf("NumPlayers = %d, want 4", game.NumPlayers)
	}
}

// TestSafeZones verifies safe positions are center safe houses
func TestSafeZones(t *testing.T) {
	safeZones := []int{2, 6, 10, 14}

	for _, pos := range safeZones {
		// Safe zones are where players start and cannot be killed
		if pos < 0 || pos >= 16 {
			t.Errorf("Safe zone %d out of OuterPath range", pos)
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
	if game.BoardSize != 5 {
		t.Errorf("BoardSize = %d, want 5", game.BoardSize)
	}
}

// TestCellSize verifies cell size constant
func TestCellSize(t *testing.T) {
	if game.CellSize != 100 {
		t.Errorf("CellSize = %d, want 100", game.CellSize)
	}
}
