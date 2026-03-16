// Package game contains the core game logic for Ashta Changa.
// This package is shared between the GUI client and the backend server.
package game

import "image/color"

// Constants for game configuration
const (
	BoardSize       = 5
	CellSize        = 100
	NumPlayers      = 4
	TokensPerPlayer = 4
)

// TokenState represents the state of a token
type TokenState int

const (
	TokenAtStart  TokenState = iota // Token is at starting point (safe house)
	TokenOnBoard                    // Token is moving on the board
	TokenFinished                   // Token has reached the center
)

// Token represents a player's token
type Token struct {
	ID        int
	PlayerIdx int
	State     TokenState
	Position  int // Position on the path (0-23 for outer path, 24 for inner path, 25 for center)
}

// Player represents a player in the game
type Player struct {
	Idx      int
	Tokens   [TokensPerPlayer]*Token
	StartPos int // Starting position on the path for this player
	EntryPos int // Position where token enters inner path
}

// ConchShell represents a single conch shell
type ConchShell struct {
	OpenSideUp bool
}

// ConchShells represents the 4 conch shells used for rolling
type ConchShells struct {
	Shells [4]ConchShell
	Result int
}

// Board represents the Ashta Changa game board
type Board struct {
	Cells [BoardSize][BoardSize]*Cell
}

// Cell represents a single cell on the board
type Cell struct {
	X, Y      int
	IsStart   bool
	IsCenter  bool
	IsPath    bool
	PlayerIdx int
}

// PlayerColors for the 4 players
var PlayerColors = []color.RGBA{
	{255, 0, 0, 255},   // Red (Top)
	{0, 128, 0, 255},   // Green (Right)
	{0, 0, 255, 255},   // Blue (Bottom)
	{255, 165, 0, 255}, // Orange (Left)
}

// PlayerNames for the 4 players
var PlayerNames = []string{"Red", "Green", "Blue", "Orange"}

// OuterPath coordinates for the outer path (anti-clockwise)
// Total 16 positions on outer path
var OuterPath = [][2]int{
	{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4},
	{1, 4}, {2, 4}, {3, 4},
	{4, 4}, {4, 3}, {4, 2}, {4, 1}, {4, 0},
	{3, 0}, {2, 0}, {1, 0},
}

// InnerCircle path blocks in clockwise order
var InnerCircle = [][2]int{
	{1, 1}, {1, 2}, {1, 3}, {2, 3}, {3, 3}, {3, 2}, {3, 1}, {2, 1},
}

// PlayerInnerStartIndices starting indices in innerCircle for each player
var PlayerInnerStartIndices = []int{0, 2, 4, 6}

// PlayerStartPositions on the path for each player (index into outerPath)
var PlayerStartPositions = []int{
	2,  // Player 0 (Red) - (0, 2)
	6,  // Player 1 (Green) - (2, 4)
	10, // Player 2 (Blue) - (4, 2)
	14, // Player 3 (Orange) - (2, 0)
}

// PlayerEntryPositions to inner path for each player (index into outerPath)
// This is the block JUST BEFORE the starting position
var PlayerEntryPositions = []int{
	1,  // Player 0 (Red) - (0, 1)
	5,  // Player 1 (Green) - (1, 4)
	9,  // Player 2 (Blue) - (4, 3)
	13, // Player 3 (Orange) - (3, 0)
}
