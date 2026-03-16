// Package game contains the core game logic for Ashta Changa.
package game

// GameLogic defines the interface for game logic operations.
// This interface allows for mocking in tests and alternative implementations.
type GameLogic interface {
	// RollFromShellStates calculates result from given shell states
	RollFromShellStates(shellStates []bool) int

	// NewPlayer creates a new player with the given index
	NewPlayer(idx int) *Player

	// NewBoard creates a new game board
	NewBoard() *Board

	// CanMoveToken checks if a token can be moved with the given roll
	CanMoveToken(token *Token, roll int, player *Player) bool

	// GetValidTokens returns indices of tokens that can be moved
	GetValidTokens(player *Player, roll int) []int

	// ApplyMove applies the roll to a token, handling transitions
	// Returns true if extra turn should be granted
	ApplyMove(token *Token, roll int, player *Player, checkKillFunc func(*Token) bool, checkWinFunc func()) bool

	// GetCellCoordinates returns the board grid coordinates for a token
	GetCellCoordinates(token *Token, playerIdx int) (int, int)

	// IsSafePosition checks if a position is a safe zone (start positions)
	IsSafePosition(position int) bool

	// CheckKill checks if the moved token lands on an opponent and kills them
	// Returns true if any opponent was killed (grants extra turn)
	CheckKill(token *Token, allPlayers []*Player) bool

	// CheckWin checks if all tokens of a player are finished
	CheckWin(player *Player) bool
}

// DefaultGameLogic implements GameLogic with the standard game rules
type DefaultGameLogic struct{}

// Ensure DefaultGameLogic implements GameLogic
var _ GameLogic = (*DefaultGameLogic)(nil)

// NewGameLogic creates a new instance of the default game logic
func NewGameLogic() GameLogic {
	return &DefaultGameLogic{}
}

// RollFromShellStates calculates result from given shell states
func (d *DefaultGameLogic) RollFromShellStates(shellStates []bool) int {
	return RollFromShellStates(shellStates)
}

// NewPlayer creates a new player with the given index
func (d *DefaultGameLogic) NewPlayer(idx int) *Player {
	return NewPlayer(idx)
}

// NewBoard creates a new game board
func (d *DefaultGameLogic) NewBoard() *Board {
	return NewBoard()
}

// CanMoveToken checks if a token can be moved with the given roll
func (d *DefaultGameLogic) CanMoveToken(token *Token, roll int, player *Player) bool {
	return CanMoveToken(token, roll, player)
}

// GetValidTokens returns indices of tokens that can be moved
func (d *DefaultGameLogic) GetValidTokens(player *Player, roll int) []int {
	return GetValidTokens(player, roll)
}

// ApplyMove applies the roll to a token, handling transitions
func (d *DefaultGameLogic) ApplyMove(token *Token, roll int, player *Player, checkKillFunc func(*Token) bool, checkWinFunc func()) bool {
	return ApplyMove(token, roll, player, checkKillFunc, checkWinFunc)
}

// GetCellCoordinates returns the board grid coordinates for a token
func (d *DefaultGameLogic) GetCellCoordinates(token *Token, playerIdx int) (int, int) {
	return GetCellCoordinates(token, playerIdx)
}

// IsSafePosition checks if a position is a safe zone
func (d *DefaultGameLogic) IsSafePosition(position int) bool {
	return IsSafePosition(position)
}

// CheckKill checks if the moved token lands on an opponent and kills them
func (d *DefaultGameLogic) CheckKill(token *Token, allPlayers []*Player) bool {
	return CheckKill(token, allPlayers)
}

// CheckWin checks if all tokens of a player are finished
func (d *DefaultGameLogic) CheckWin(player *Player) bool {
	return CheckWin(player)
}
