// Package game contains the core game logic for Ashta Changa.
package game

import "math/rand"

// NewConchShells creates new conch shells
func NewConchShells() *ConchShells {
	return &ConchShells{}
}

// Roll rolls all 4 conch shells and calculates the result
func (cs *ConchShells) Roll() int {
	openCount := 0
	for i := range cs.Shells {
		cs.Shells[i].OpenSideUp = rand.Intn(2) == 1
		if cs.Shells[i].OpenSideUp {
			openCount++
		}
	}

	switch openCount {
	case 0:
		cs.Result = 8
	case 1:
		cs.Result = 1
	case 2:
		cs.Result = 2
	case 3:
		cs.Result = 3
	case 4:
		cs.Result = 4
	}

	return cs.Result
}

// RollFromShellStates calculates result from given shell states
func RollFromShellStates(shellStates []bool) int {
	openCount := 0
	for _, open := range shellStates {
		if open {
			openCount++
		}
	}

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

// NewPlayer creates a new player
func NewPlayer(idx int) *Player {
	p := &Player{
		Idx:      idx,
		StartPos: PlayerStartPositions[idx],
		EntryPos: PlayerEntryPositions[idx],
	}

	for i := 0; i < TokensPerPlayer; i++ {
		p.Tokens[i] = &Token{
			ID:        i,
			PlayerIdx: idx,
			State:     TokenAtStart,
			Position:  PlayerStartPositions[idx],
		}
	}

	return p
}

// NewBoard creates a new Ashta Changa board
func NewBoard() *Board {
	b := &Board{}

	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			b.Cells[y][x] = &Cell{
				X:         x,
				Y:         y,
				PlayerIdx: -1,
			}
		}
	}

	b.Cells[2][2].IsCenter = true

	b.Cells[0][2].IsStart = true
	b.Cells[0][2].PlayerIdx = 0

	b.Cells[2][4].IsStart = true
	b.Cells[2][4].PlayerIdx = 1

	b.Cells[4][2].IsStart = true
	b.Cells[4][2].PlayerIdx = 2

	b.Cells[2][0].IsStart = true
	b.Cells[2][0].PlayerIdx = 3

	for x := 0; x < BoardSize; x++ {
		b.Cells[0][x].IsPath = true
	}
	for y := 1; y < BoardSize; y++ {
		b.Cells[y][4].IsPath = true
	}
	for x := 0; x < BoardSize; x++ {
		b.Cells[4][x].IsPath = true
	}
	for y := 1; y < BoardSize-1; y++ {
		b.Cells[y][0].IsPath = true
	}

	// Mark inner circle as path
	for _, coords := range InnerCircle {
		b.Cells[coords[0]][coords[1]].IsPath = true
	}

	return b
}

// CanMoveToken checks if a token can be moved with the given roll
func CanMoveToken(token *Token, roll int, player *Player) bool {
	if token.State == TokenFinished {
		return false
	}

	// Token at start - needs 1, 4, or 8 to move out
	if token.State == TokenAtStart {
		return roll == 1 || roll == 4 || roll == 8
	}

	// Token on board - check if move is valid
	if token.State == TokenOnBoard {
		// Check if token is on inner path
		// position 16-23: steps 0-7 on inner circle
		if token.Position >= 16 {
			step := token.Position - 16
			if step+roll <= 8 { // 8 means center
				return true
			}
			return false
		}

		// On outer path
		// Check if would pass entry point
		distToEntry := (player.EntryPos - token.Position + 16) % 16

		if roll > distToEntry {
			// Steps remaining after reaching entry point
			remaining := roll - distToEntry - 1 // -1 to enter inner circle

			// After entry point, it enters the inner circle
			// It then has to complete 8 steps in the circle before center
			if remaining <= 8 {
				return true
			}
			return false
		}

		return true
	}

	return false
}

// GetValidTokens returns indices of tokens that can be moved
func GetValidTokens(player *Player, roll int) []int {
	var valid []int

	for i, token := range player.Tokens {
		if CanMoveToken(token, roll, player) {
			valid = append(valid, i)
		}
	}

	return valid
}

// ApplyMove applies the roll to a token, handling transitions
// Returns true if extra turn should be granted
func ApplyMove(token *Token, roll int, player *Player, checkKillFunc func(*Token) bool, checkWinFunc func()) bool {
	extraTurn := false

	if token.Position >= 16 {
		// On inner path
		step := token.Position - 16
		newStep := step + roll
		if newStep >= 8 {
			token.State = TokenFinished
			token.Position = 24
			extraTurn = true // Rule: Reaching center grants a bonus roll
			if checkWinFunc != nil {
				checkWinFunc()
			}
		} else {
			token.Position = 16 + newStep
			if checkKillFunc != nil {
				checkKillFunc(token) // Rule: Killing possible in inner circle too
			}
		}
		return extraTurn
	}

	// On outer path
	distToEntry := (player.EntryPos - token.Position + 16) % 16
	if roll > distToEntry {
		// Transition to inner path
		remaining := roll - distToEntry - 1
		if remaining >= 8 {
			token.State = TokenFinished
			token.Position = 24
			extraTurn = true // Rule: Reaching center grants a bonus roll
			if checkWinFunc != nil {
				checkWinFunc()
			}
		} else {
			token.Position = 16 + remaining
			if checkKillFunc != nil {
				checkKillFunc(token) // Rule: Killing possible in inner circle too
			}
		}
	} else {
		// Stay on outer path
		token.Position = (token.Position + roll) % 16
		if checkKillFunc != nil {
			checkKillFunc(token)
		}
	}

	return extraTurn
}

// GetCellCoordinates returns the board grid coordinates for a token
func GetCellCoordinates(token *Token, playerIdx int) (int, int) {
	if token.State == TokenFinished {
		return 2, 2 // Center block
	}

	if token.Position >= 16 {
		// Inner path
		step := token.Position - 16
		innerIdx := (PlayerInnerStartIndices[playerIdx] + step) % 8
		coords := InnerCircle[innerIdx]
		return coords[0], coords[1]
	}

	// Outer path
	coords := OuterPath[token.Position]
	return coords[0], coords[1]
}

// IsSafePosition checks if a position is a safe zone (start positions)
func IsSafePosition(position int) bool {
	if position >= 16 {
		return false // Inner circle is not safe
	}
	for _, startPos := range PlayerStartPositions {
		if position == startPos {
			return true
		}
	}
	return false
}

// CheckKill checks if the moved token lands on an opponent and kills them
// Returns true if any opponent was killed (grants extra turn)
func CheckKill(token *Token, allPlayers []*Player) bool {
	if token.State != TokenOnBoard {
		return false
	}

	// Get current grid coordinates
	row, col := GetCellCoordinates(token, token.PlayerIdx)

	// Rule: Starting blocks are safe houses. No killing can happen on a safe house.
	for i := 0; i < NumPlayers; i++ {
		startCoords := OuterPath[PlayerStartPositions[i]]
		if row == startCoords[0] && col == startCoords[1] {
			return false // Safe house, no kill
		}
	}

	// Rule: If landing on an opponent's token in a non-safe block, kill it.
	killedAny := false
	for _, player := range allPlayers {
		if player == nil || player.Idx == token.PlayerIdx {
			continue // Cannot kill own tokens or from non-existent players
		}
		for _, opponentToken := range player.Tokens {
			if opponentToken.State == TokenOnBoard {
				or, oc := GetCellCoordinates(opponentToken, opponentToken.PlayerIdx)
				if or == row && oc == col {
					// Kill opponent token and send it back to starting point
					opponentToken.State = TokenAtStart
					opponentToken.Position = PlayerStartPositions[opponentToken.PlayerIdx]
					killedAny = true
				}
			}
		}
	}

	return killedAny
}

// CheckWin checks if all tokens of a player are finished
func CheckWin(player *Player) bool {
	if player == nil {
		return false
	}
	finishedCount := 0
	for _, token := range player.Tokens {
		if token.State == TokenFinished {
			finishedCount++
		}
	}
	return finishedCount == TokensPerPlayer
}
