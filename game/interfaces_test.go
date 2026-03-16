// Package game contains the core game logic for Ashta Changa.
package game

import (
	"testing"
)

// TestGameLogicInterface verifies GameLogic interface is properly implemented
func TestGameLogicInterface(t *testing.T) {
	var _ GameLogic = (*DefaultGameLogic)(nil)
}

// TestDefaultGameLogicNewPlayer verifies DefaultGameLogic.NewPlayer
func TestDefaultGameLogicNewPlayer(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)

	if player == nil {
		t.Fatal("NewPlayer should not return nil")
	}
	if player.Idx != 0 {
		t.Errorf("Player index = %d, want 0", player.Idx)
	}
}

// TestDefaultGameLogicNewBoard verifies DefaultGameLogic.NewBoard
func TestDefaultGameLogicNewBoard(t *testing.T) {
	logic := NewGameLogic()
	board := logic.NewBoard()

	if board == nil {
		t.Fatal("NewBoard should not return nil")
	}
	if !board.Cells[2][2].IsCenter {
		t.Error("Board center should be marked")
	}
}

// TestDefaultGameLogicCanMoveToken verifies DefaultGameLogic.CanMoveToken
func TestDefaultGameLogicCanMoveToken(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)
	token := player.Tokens[0]

	// At start, can move on 1, 4, 8
	if !logic.CanMoveToken(token, 1, player) {
		t.Error("Should be able to move with roll 1 from start")
	}
	if logic.CanMoveToken(token, 2, player) {
		t.Error("Should NOT be able to move with roll 2 from start")
	}
}

// TestDefaultGameLogicGetValidTokens verifies DefaultGameLogic.GetValidTokens
func TestDefaultGameLogicGetValidTokens(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)

	valid := logic.GetValidTokens(player, 1)
	if len(valid) != 4 {
		t.Errorf("Expected 4 valid tokens, got %d", len(valid))
	}
}

// TestDefaultGameLogicApplyMove verifies DefaultGameLogic.ApplyMove
func TestDefaultGameLogicApplyMove(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)
	token := player.Tokens[0]
	token.State = TokenOnBoard
	token.Position = 23

	extraTurn := logic.ApplyMove(token, 1, player, nil, nil)

	if !extraTurn {
		t.Error("Should get extra turn for reaching center")
	}
	if token.State != TokenFinished {
		t.Error("Token should be finished")
	}
}

// TestDefaultGameLogicGetCellCoordinates verifies DefaultGameLogic.GetCellCoordinates
func TestDefaultGameLogicGetCellCoordinates(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)
	token := player.Tokens[0]

	r, c := logic.GetCellCoordinates(token, 0)
	if r != 0 || c != 2 {
		t.Errorf("Start coords = (%d, %d), want (0, 2)", r, c)
	}
}

// TestDefaultGameLogicIsSafePosition verifies DefaultGameLogic.IsSafePosition
func TestDefaultGameLogicIsSafePosition(t *testing.T) {
	logic := NewGameLogic()

	if !logic.IsSafePosition(2) {
		t.Error("Position 2 should be safe")
	}
	if logic.IsSafePosition(0) {
		t.Error("Position 0 should NOT be safe")
	}
}

// TestDefaultGameLogicCheckKill verifies DefaultGameLogic.CheckKill
func TestDefaultGameLogicCheckKill(t *testing.T) {
	logic := NewGameLogic()
	players := []*Player{logic.NewPlayer(0), logic.NewPlayer(1)}

	// Both at same unsafe position
	players[0].Tokens[0].State = TokenOnBoard
	players[0].Tokens[0].Position = 5
	players[1].Tokens[0].State = TokenOnBoard
	players[1].Tokens[0].Position = 5

	killed := logic.CheckKill(players[0].Tokens[0], players)
	if !killed {
		t.Error("Should kill opponent")
	}
}

// TestDefaultGameLogicCheckWin verifies DefaultGameLogic.CheckWin
func TestDefaultGameLogicCheckWin(t *testing.T) {
	logic := NewGameLogic()
	player := logic.NewPlayer(0)

	// Not a win initially
	if logic.CheckWin(player) {
		t.Error("Should not win initially")
	}

	// Set all tokens to finished
	for _, token := range player.Tokens {
		token.State = TokenFinished
	}

	if !logic.CheckWin(player) {
		t.Error("Should win with all tokens finished")
	}
}

// TestDefaultGameLogicRollFromShellStates verifies DefaultGameLogic.RollFromShellStates
func TestDefaultGameLogicRollFromShellStates(t *testing.T) {
	logic := NewGameLogic()

	shells := []bool{false, false, false, false}
	result := logic.RollFromShellStates(shells)
	if result != 8 {
		t.Errorf("All closed = 8, got %d", result)
	}
}

// MockGameLogic is a mock implementation for testing
type MockGameLogic struct {
	NewPlayerCalled          bool
	NewBoardCalled           bool
	CanMoveTokenCalled       bool
	GetValidTokensCalled     bool
	ApplyMoveCalled          bool
	GetCellCoordinatesCalled bool
	IsSafePositionCalled     bool
	CheckKillCalled          bool
	CheckWinCalled           bool
	RollFromShellStatesCalled bool
}

func (m *MockGameLogic) NewPlayer(idx int) *Player {
	m.NewPlayerCalled = true
	return &Player{Idx: idx}
}

func (m *MockGameLogic) NewBoard() *Board {
	m.NewBoardCalled = true
	return &Board{}
}

func (m *MockGameLogic) CanMoveToken(token *Token, roll int, player *Player) bool {
	m.CanMoveTokenCalled = true
	return true
}

func (m *MockGameLogic) GetValidTokens(player *Player, roll int) []int {
	m.GetValidTokensCalled = true
	return []int{0, 1, 2, 3}
}

func (m *MockGameLogic) ApplyMove(token *Token, roll int, player *Player, checkKillFunc func(*Token) bool, checkWinFunc func()) bool {
	m.ApplyMoveCalled = true
	return false
}

func (m *MockGameLogic) GetCellCoordinates(token *Token, playerIdx int) (int, int) {
	m.GetCellCoordinatesCalled = true
	return 0, 0
}

func (m *MockGameLogic) IsSafePosition(position int) bool {
	m.IsSafePositionCalled = true
	return false
}

func (m *MockGameLogic) CheckKill(token *Token, allPlayers []*Player) bool {
	m.CheckKillCalled = true
	return false
}

func (m *MockGameLogic) CheckWin(player *Player) bool {
	m.CheckWinCalled = true
	return false
}

func (m *MockGameLogic) RollFromShellStates(shellStates []bool) int {
	m.RollFromShellStatesCalled = true
	return 4
}

// TestMockGameLogic verifies mock implementation works
func TestMockGameLogic(t *testing.T) {
	var _ GameLogic = (*MockGameLogic)(nil)

	mock := &MockGameLogic{}

	// Call all methods
	mock.NewPlayer(0)
	mock.NewBoard()
	mock.CanMoveToken(nil, 0, nil)
	mock.GetValidTokens(nil, 0)
	mock.ApplyMove(nil, 0, nil, nil, nil)
	mock.GetCellCoordinates(nil, 0)
	mock.IsSafePosition(0)
	mock.CheckKill(nil, nil)
	mock.CheckWin(nil)
	mock.RollFromShellStates(nil)

	// Verify all methods were called
	if !mock.NewPlayerCalled {
		t.Error("NewPlayer was not called")
	}
	if !mock.NewBoardCalled {
		t.Error("NewBoard was not called")
	}
	if !mock.CanMoveTokenCalled {
		t.Error("CanMoveToken was not called")
	}
	if !mock.GetValidTokensCalled {
		t.Error("GetValidTokens was not called")
	}
	if !mock.ApplyMoveCalled {
		t.Error("ApplyMove was not called")
	}
	if !mock.GetCellCoordinatesCalled {
		t.Error("GetCellCoordinates was not called")
	}
	if !mock.IsSafePositionCalled {
		t.Error("IsSafePosition was not called")
	}
	if !mock.CheckKillCalled {
		t.Error("CheckKill was not called")
	}
	if !mock.CheckWinCalled {
		t.Error("CheckWin was not called")
	}
	if !mock.RollFromShellStatesCalled {
		t.Error("RollFromShellStates was not called")
	}
}
