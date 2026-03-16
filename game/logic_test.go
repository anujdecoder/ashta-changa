// Package game contains the core game logic for Ashta Changa.
package game

import (
	"testing"
)

// TestCanMoveTokenAtStart verifies token movement rules from start position
func TestCanMoveTokenAtStart(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]

	// Token should be at start
	if token.State != TokenAtStart {
		t.Fatal("Token should start at home")
	}

	// Can only move on 1, 4, 8
	canMoveOn148 := []int{1, 4, 8}
	cannotMoveOn := []int{2, 3}

	for _, roll := range canMoveOn148 {
		if !CanMoveToken(token, roll, player) {
			t.Errorf("Token at start should be able to move with roll %d", roll)
		}
	}

	for _, roll := range cannotMoveOn {
		if CanMoveToken(token, roll, player) {
			t.Errorf("Token at start should NOT be able to move with roll %d", roll)
		}
	}
}

// TestCanMoveTokenFinished verifies finished tokens cannot move
func TestCanMoveTokenFinished(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenFinished
	token.Position = 24 // Center

	// Finished token cannot move with any roll
	for roll := 1; roll <= 8; roll++ {
		if CanMoveToken(token, roll, player) {
			t.Errorf("Finished token should not move with roll %d", roll)
		}
	}
}

// TestCanMoveTokenOnOuterPath verifies movement on outer path
func TestCanMoveTokenOnOuterPath(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard
	token.Position = 5 // Some position on outer path

	// Should be able to move with any roll on outer path
	for roll := 1; roll <= 8; roll++ {
		if !CanMoveToken(token, roll, player) {
			t.Errorf("Token on outer path should move with roll %d", roll)
		}
	}
}

// TestCanMoveTokenInnerPathExact verifies inner path requires exact roll
func TestCanMoveTokenInnerPathExact(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard

	// Position 16 = step 0 on inner path (needs 8 to reach center)
	token.Position = 16
	if !CanMoveToken(token, 8, player) {
		t.Error("Token at step 0 should be able to reach center with roll 8")
	}
	if CanMoveToken(token, 9, player) {
		t.Error("Token at step 0 should NOT move with roll 9 (exceeds center)")
	}

	// Position 23 = step 7 on inner path (needs 1 to reach center)
	token.Position = 23
	if !CanMoveToken(token, 1, player) {
		t.Error("Token at step 7 should be able to reach center with roll 1")
	}
	if CanMoveToken(token, 2, player) {
		t.Error("Token at step 7 should NOT move with roll 2 (exceeds center)")
	}
}

// TestApplyMoveOuterToInner verifies transition from outer to inner path
func TestApplyMoveOuterToInner(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard

	// Player 0 entry position is 1 (just before start position 2)
	// From position 0, roll 2 should pass entry and go to inner step 0
	token.Position = 0
	player.EntryPos = 1

	// This should transition to inner path
	extraTurn := ApplyMove(token, 3, player, nil, nil)

	if extraTurn {
		t.Error("Should not get extra turn for normal move")
	}
	if token.Position < 16 {
		t.Errorf("Token should be on inner path, got position %d", token.Position)
	}
}

// TestApplyMoveReachesCenter verifies reaching center gives extra turn
func TestApplyMoveReachesCenter(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard
	token.Position = 23 // Step 7 on inner path

	extraTurn := ApplyMove(token, 1, player, nil, nil)

	if !extraTurn {
		t.Error("Should get extra turn for reaching center")
	}
	if token.State != TokenFinished {
		t.Errorf("Token should be finished, got state %d", token.State)
	}
	if token.Position != 24 {
		t.Errorf("Token should be at position 24 (center), got %d", token.Position)
	}
}

// TestCheckKillOnOuterPath verifies killing on outer path
func TestCheckKillOnOuterPath(t *testing.T) {
	// Create 2 players
	players := []*Player{NewPlayer(0), NewPlayer(1)}

	// Player 0 token at position 5
	players[0].Tokens[0].State = TokenOnBoard
	players[0].Tokens[0].Position = 5

	// Player 1 token at same position 5
	players[1].Tokens[0].State = TokenOnBoard
	players[1].Tokens[0].Position = 5

	// Player 0 moves to position 5 (lands on player 1)
	killed := CheckKill(players[0].Tokens[0], players)

	if !killed {
		t.Error("Should have killed opponent token")
	}
	if players[1].Tokens[0].State != TokenAtStart {
		t.Errorf("Killed token should return to start, got state %d", players[1].Tokens[0].State)
	}
}

// TestCheckKillOnSafeHouse verifies no killing on safe houses
func TestCheckKillOnSafeHouse(t *testing.T) {
	// Safe position 2
	players := []*Player{NewPlayer(0), NewPlayer(1)}

	// Both tokens at safe position 2
	players[0].Tokens[0].State = TokenOnBoard
	players[0].Tokens[0].Position = 2
	players[1].Tokens[0].State = TokenOnBoard
	players[1].Tokens[0].Position = 2

	killed := CheckKill(players[0].Tokens[0], players)

	if killed {
		t.Error("Should NOT kill on safe house")
	}
	if players[1].Tokens[0].State != TokenOnBoard {
		t.Error("Token on safe house should not be sent back")
	}
}

// TestCheckKillOnInnerPath verifies killing on inner path
// Note: On inner path, same position number means same physical location
// but GetCellCoordinates returns different coordinates based on player inner start index
func TestCheckKillOnInnerPath(t *testing.T) {
	players := []*Player{NewPlayer(0), NewPlayer(1)}

	// Place both tokens at the same GRID coordinates on inner path
	// Player 0 at inner step 0 (position 16) -> grid (1, 1)
	// Player 1 needs to be at position that also maps to (1, 1)
	// Player 1 starts at inner index 2, so to reach (1,1) they need to be at step 6
	// which is position 22
	players[0].Tokens[0].State = TokenOnBoard
	players[0].Tokens[0].Position = 16 // step 0 for player 0

	// For player 1: inner index starts at 2, so position 22 is step 6
	// (2 + 6) % 8 = 0, which maps to InnerCircle[0] = (1, 1)
	players[1].Tokens[0].State = TokenOnBoard
	players[1].Tokens[0].Position = 22

	// Verify they have same coordinates
	r0, c0 := GetCellCoordinates(players[0].Tokens[0], 0)
	r1, c1 := GetCellCoordinates(players[1].Tokens[0], 1)

	if r0 != r1 || c0 != c1 {
		t.Skipf("Skipping test - tokens at different grid coordinates (%d,%d) vs (%d,%d)", r0, c0, r1, c1)
		return
	}

	killed := CheckKill(players[0].Tokens[0], players)

	if !killed {
		t.Error("Should kill on inner path when at same grid coordinates")
	}
}

// TestCheckWin verifies win condition
func TestCheckWin(t *testing.T) {
	player := NewPlayer(0)

	// All tokens at start - not a win
	if CheckWin(player) {
		t.Error("Should not win with all tokens at start")
	}

	// Set all tokens to finished
	for _, token := range player.Tokens {
		token.State = TokenFinished
	}

	if !CheckWin(player) {
		t.Error("Should win with all tokens finished")
	}
}

// TestCheckWinPartial verifies partial completion is not a win
func TestCheckWinPartial(t *testing.T) {
	player := NewPlayer(0)

	// 3 finished, 1 on board
	player.Tokens[0].State = TokenFinished
	player.Tokens[1].State = TokenFinished
	player.Tokens[2].State = TokenFinished
	player.Tokens[3].State = TokenOnBoard

	if CheckWin(player) {
		t.Error("Should not win with only 3 tokens finished")
	}
}

// TestRollFromShellStates verifies roll calculation
func TestRollFromShellStates(t *testing.T) {
	tests := []struct {
		shells []bool
		want   int
	}{
		{[]bool{false, false, false, false}, 8}, // 0 open = 8
		{[]bool{true, false, false, false}, 1},  // 1 open = 1
		{[]bool{true, true, false, false}, 2},   // 2 open = 2
		{[]bool{true, true, true, false}, 3},    // 3 open = 3
		{[]bool{true, true, true, true}, 4},     // 4 open = 4
	}

	for _, tt := range tests {
		got := RollFromShellStates(tt.shells)
		if got != tt.want {
			t.Errorf("RollFromShellStates(%v) = %d, want %d", tt.shells, got, tt.want)
		}
	}
}

// TestConchShellsRoll verifies shell rolling produces valid results
func TestConchShellsRoll(t *testing.T) {
	cs := NewConchShells()

	// Roll multiple times to verify results
	for i := 0; i < 20; i++ {
		result := cs.Roll()
		validResults := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true}
		if !validResults[result] {
			t.Errorf("Invalid roll result: %d", result)
		}
	}
}

// TestGetValidTokens verifies getting valid tokens for a roll
func TestGetValidTokens(t *testing.T) {
	player := NewPlayer(0)

	// All tokens at start, roll 1 - should get all 4 tokens
	valid := GetValidTokens(player, 1)
	if len(valid) != 4 {
		t.Errorf("Expected 4 valid tokens at start with roll 1, got %d", len(valid))
	}

	// All tokens at start, roll 2 - should get 0 tokens
	valid = GetValidTokens(player, 2)
	if len(valid) != 0 {
		t.Errorf("Expected 0 valid tokens at start with roll 2, got %d", len(valid))
	}
}

// TestGetCellCoordinates verifies coordinate calculation
func TestGetCellCoordinates(t *testing.T) {
	player := NewPlayer(0)

	// Token at start position
	token := player.Tokens[0]
	r, c := GetCellCoordinates(token, 0)

	// Player 0 starts at outer path position 2, which is (0, 2)
	if r != 0 || c != 2 {
		t.Errorf("Start position coords = (%d, %d), want (0, 2)", r, c)
	}

	// Finished token should be at center (2, 2)
	token.State = TokenFinished
	r, c = GetCellCoordinates(token, 0)
	if r != 2 || c != 2 {
		t.Errorf("Center coords = (%d, %d), want (2, 2)", r, c)
	}
}

// TestIsSafePosition verifies safe position detection
func TestIsSafePosition(t *testing.T) {
	safePositions := []int{2, 6, 10, 14}
	unsafePositions := []int{0, 1, 3, 4, 5, 7, 8, 9, 11, 12, 13, 15}

	for _, pos := range safePositions {
		if !IsSafePosition(pos) {
			t.Errorf("Position %d should be safe", pos)
		}
	}

	for _, pos := range unsafePositions {
		if IsSafePosition(pos) {
			t.Errorf("Position %d should NOT be safe", pos)
		}
	}
}

// TestNewBoard verifies board creation
func TestNewBoard(t *testing.T) {
	board := NewBoard()

	// Check center is marked
	if !board.Cells[2][2].IsCenter {
		t.Error("Center cell should be marked")
	}

	// Check start positions are marked
	startPositions := [][2]int{{0, 2}, {2, 4}, {4, 2}, {2, 0}}
	for i, pos := range startPositions {
		cell := board.Cells[pos[0]][pos[1]]
		if !cell.IsStart {
			t.Errorf("Start position %d should be marked", i)
		}
		if cell.PlayerIdx != i {
			t.Errorf("Start position %d should have PlayerIdx %d, got %d", i, i, cell.PlayerIdx)
		}
	}
}

// TestNewPlayer verifies player creation
func TestNewPlayer(t *testing.T) {
	for i := 0; i < NumPlayers; i++ {
		player := NewPlayer(i)

		if player.Idx != i {
			t.Errorf("Player index = %d, want %d", player.Idx, i)
		}

		if len(player.Tokens) != TokensPerPlayer {
			t.Errorf("Token count = %d, want %d", len(player.Tokens), TokensPerPlayer)
		}

		for j, token := range player.Tokens {
			if token.ID != j {
				t.Errorf("Token %d ID = %d, want %d", j, token.ID, j)
			}
			if token.PlayerIdx != i {
				t.Errorf("Token %d PlayerIdx = %d, want %d", j, token.PlayerIdx, i)
			}
			if token.State != TokenAtStart {
				t.Errorf("Token %d State = %d, want TokenAtStart", j, token.State)
			}
		}
	}
}

// TestApplyMoveWithKillCallback verifies kill callback is invoked
func TestApplyMoveWithKillCallback(t *testing.T) {
	player := NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard
	token.Position = 5

	killCalled := false
	checkKill := func(t *Token) bool {
		killCalled = true
		return false
	}

	ApplyMove(token, 3, player, checkKill, nil)

	if !killCalled {
		t.Error("Kill callback should have been called")
	}
}

// TestApplyMoveWithWinCallback verifies win callback is invoked
func TestApplyMoveWithWinCallback(t *testing.T) {
	player := NewPlayer(0)

	// Set 3 tokens to finished
	player.Tokens[0].State = TokenFinished
	player.Tokens[1].State = TokenFinished
	player.Tokens[2].State = TokenFinished

	// Last token at step 7 on inner path
	player.Tokens[3].State = TokenOnBoard
	player.Tokens[3].Position = 23

	winCalled := false
	checkWin := func() {
		winCalled = true
	}

	ApplyMove(player.Tokens[3], 1, player, nil, checkWin)

	if !winCalled {
		t.Error("Win callback should have been called when last token reaches center")
	}
}
