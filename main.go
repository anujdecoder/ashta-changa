package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 600
	screenHeight = 600
	boardSize    = 5
	cellSize     = 100
)

// Player colors for the 4 players
var playerColors = []color.RGBA{
	{255, 0, 0, 255},   // Red (Top)
	{0, 128, 0, 255},   // Green (Right)
	{0, 0, 255, 255},   // Blue (Bottom)
	{255, 165, 0, 255}, // Orange (Left)
}

// Board represents the Ashta Changa game board
type Board struct {
	cells [boardSize][boardSize]*Cell
}

// Cell represents a single cell on the board
type Cell struct {
	x, y      int
	isStart   bool   // Starting point for a player
	isCenter  bool   // Center block (goal)
	isPath    bool   // Part of the movement path
	isSafe    bool   // Safe zone (home column)
	playerIdx int    // Player index (0-3) for start/safe cells, -1 otherwise
}

// Game represents the game state
type Game struct {
	board *Board
}

// NewBoard creates a new Ashta Changa board
func NewBoard() *Board {
	b := &Board{}

	// Initialize all cells
	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			b.cells[y][x] = &Cell{
				x:         x,
				y:         y,
				playerIdx: -1,
			}
		}
	}

	// Mark the center block (goal)
	b.cells[2][2].isCenter = true

	// Mark starting points (center of each border edge)
	// Player 0 (Red) - Top center
	b.cells[0][2].isStart = true
	b.cells[0][2].playerIdx = 0

	// Player 1 (Green) - Right center
	b.cells[2][4].isStart = true
	b.cells[2][4].playerIdx = 1

	// Player 2 (Blue) - Bottom center
	b.cells[4][2].isStart = true
	b.cells[4][2].playerIdx = 2

	// Player 3 (Orange) - Left center
	b.cells[2][0].isStart = true
	b.cells[2][0].playerIdx = 3

	// Mark the outer path cells (anti-clockwise movement path)
	// Top row (left to right)
	for x := 0; x < boardSize; x++ {
		b.cells[0][x].isPath = true
	}
	// Right column (top to bottom)
	for y := 1; y < boardSize; y++ {
		b.cells[y][4].isPath = true
	}
	// Bottom row (right to left)
	for x := 0; x < boardSize; x++ {
		b.cells[4][x].isPath = true
	}
	// Left column (bottom to top)
	for y := 1; y < boardSize-1; y++ {
		b.cells[y][0].isPath = true
	}

	// Mark safe zones (home columns leading to center)
	// Player 0 (Red) - column 2, row 1 (path to center from top)
	b.cells[1][2].isSafe = true
	b.cells[1][2].isPath = false
	b.cells[1][2].playerIdx = 0

	// Player 1 (Green) - row 2, column 3 (path to center from right)
	b.cells[2][3].isSafe = true
	b.cells[2][3].isPath = false
	b.cells[2][3].playerIdx = 1

	// Player 2 (Blue) - column 2, row 3 (path to center from bottom)
	b.cells[3][2].isSafe = true
	b.cells[3][2].isPath = false
	b.cells[3][2].playerIdx = 2

	// Player 3 (Orange) - row 2, column 1 (path to center from left)
	b.cells[2][1].isSafe = true
	b.cells[2][1].isPath = false
	b.cells[2][1].playerIdx = 3

	return b
}

// Draw draws the board
func (b *Board) Draw(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.RGBA{245, 222, 179, 255}) // Wheat color for traditional look

	// Draw cells
	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			cell := b.cells[y][x]
			b.drawCell(screen, cell)
		}
	}

	// Draw grid lines
	b.drawGrid(screen)

	// Draw starting point markers (X)
	b.drawStartMarkers(screen)
}

// drawCell draws a single cell
func (b *Board) drawCell(screen *ebiten.Image, cell *Cell) {
	x := cell.x * cellSize
	y := cell.y * cellSize

	var fillColor color.RGBA

	switch {
	case cell.isCenter:
		// Center block - golden color
		fillColor = color.RGBA{255, 215, 0, 255}
	case cell.isStart:
		// Starting points - use player color with transparency
		fillColor = playerColors[cell.playerIdx]
		fillColor.A = 180
	case cell.isSafe:
		// Safe zones - lighter version of player color
		fillColor = playerColors[cell.playerIdx]
		fillColor.R = fillColor.R/2 + 127
		fillColor.G = fillColor.G/2 + 127
		fillColor.B = fillColor.B/2 + 127
		fillColor.A = 200
	case cell.isPath:
		// Path cells - light brown
		fillColor = color.RGBA{210, 180, 140, 255}
	default:
		// Empty cells (corners) - darker background
		fillColor = color.RGBA{139, 90, 43, 255}
	}

	vector.DrawFilledRect(screen, float32(x+2), float32(y+2), float32(cellSize-4), float32(cellSize-4), fillColor, false)
}

// drawGrid draws the grid lines
func (b *Board) drawGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{101, 67, 33, 255} // Dark brown

	// Vertical lines
	for i := 0; i <= boardSize; i++ {
		x := i * cellSize
		vector.StrokeLine(screen, float32(x), 0, float32(x), screenHeight, 2, gridColor, false)
	}

	// Horizontal lines
	for i := 0; i <= boardSize; i++ {
		y := i * cellSize
		vector.StrokeLine(screen, 0, float32(y), screenWidth, float32(y), 2, gridColor, false)
	}
}

// drawStartMarkers draws X markers on starting points
func (b *Board) drawStartMarkers(screen *ebiten.Image) {
	markerColor := color.RGBA{0, 0, 0, 255} // Black for X marker

	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			cell := b.cells[y][x]
			if cell.isStart {
				cx := x*cellSize + cellSize/2
				cy := y*cellSize + cellSize/2
				offset := 25

				// Draw X using lines
				vector.StrokeLine(screen, float32(cx-offset), float32(cy-offset), float32(cx+offset), float32(cy+offset), 3, markerColor, false)
				vector.StrokeLine(screen, float32(cx+offset), float32(cy-offset), float32(cx-offset), float32(cy+offset), 3, markerColor, false)
			}
		}
	}
}

// NewGame creates a new game instance
func NewGame() *Game {
	return &Game{
		board: NewBoard(),
	}
}

// Update updates the game state
func (g *Game) Update() error {
	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	g.board.Draw(screen)

	// Draw title
	ebitenutil.DebugPrint(screen, "Ashta Changa - Traditional Indian Board Game")
}

// Layout returns the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ashta Changa")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
