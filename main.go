package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	boardSize    = 5
	cellSize     = 100
	boardOffsetX = 50
	boardOffsetY = 50
)

// Player colors for the 4 players
var playerColors = []color.RGBA{
	{255, 0, 0, 255},   // Red (Top)
	{0, 128, 0, 255},   // Green (Right)
	{0, 0, 255, 255},   // Blue (Bottom)
	{255, 165, 0, 255}, // Orange (Left)
}

// ConchShell represents a single conch shell
type ConchShell struct {
	openSideUp bool // true = open side up (counts as 1), false = closed side up (counts as 0)
}

// ConchShells represents the 4 conch shells used for rolling
type ConchShells struct {
	shells [4]ConchShell
	result int // Current roll result (0-4 open shells, mapped to game values)
}

// Board represents the Ashta Changa game board
type Board struct {
	cells [boardSize][boardSize]*Cell
}

// Cell represents a single cell on the board
type Cell struct {
	x, y      int
	isStart   bool // Starting point for a player (also the only safe zones)
	isCenter  bool // Center block (goal)
	isPath    bool // Part of the movement path
	playerIdx int  // Player index (0-3) for start cells, -1 otherwise
}

// Game represents the game state
type Game struct {
	board       *Board
	conchShells *ConchShells
	rollResult  int
	hasRolled   bool
}

// NewConchShells creates new conch shells
func NewConchShells() *ConchShells {
	return &ConchShells{}
}

// Roll rolls all 4 conch shells and calculates the result
func (cs *ConchShells) Roll() int {
	openCount := 0
	for i := range cs.shells {
		// Randomly determine if shell is open (1) or closed (0)
		cs.shells[i].openSideUp = rand.Intn(2) == 1
		if cs.shells[i].openSideUp {
			openCount++
		}
	}

	// Map the open count to game values
	// 0 open = 8 (Ashta)
	// 1 open = 1
	// 2 open = 2
	// 3 open = 3
	// 4 open = 4 (Changa)
	switch openCount {
	case 0:
		cs.result = 8 // Ashta (all closed)
	case 1:
		cs.result = 1
	case 2:
		cs.result = 2
	case 3:
		cs.result = 3
	case 4:
		cs.result = 4 // Changa (all open)
	}

	return cs.result
}

// GetOpenCount returns the number of shells with open side up
func (cs *ConchShells) GetOpenCount() int {
	count := 0
	for _, shell := range cs.shells {
		if shell.openSideUp {
			count++
		}
	}
	return count
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

	// Mark starting points (center of each border edge) - these are the ONLY safe zones
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

	// Mark inner path cells (path to center from starting point)
	// Player 0 (Red) - column 2, row 1 (path to center from top)
	b.cells[1][2].isPath = true

	// Player 1 (Green) - row 2, column 3 (path to center from right)
	b.cells[2][3].isPath = true

	// Player 2 (Blue) - column 2, row 3 (path to center from bottom)
	b.cells[3][2].isPath = true

	// Player 3 (Orange) - row 2, column 1 (path to center from left)
	b.cells[2][1].isPath = true

	return b
}

// Draw draws the board
func (b *Board) Draw(screen *ebiten.Image) {
	// Draw board background
	boardBg := color.RGBA{245, 222, 179, 255} // Wheat color for traditional look
	vector.DrawFilledRect(screen, float32(boardOffsetX), float32(boardOffsetY), float32(boardSize*cellSize), float32(boardSize*cellSize), boardBg, false)

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
	x := boardOffsetX + cell.x*cellSize
	y := boardOffsetY + cell.y*cellSize

	var fillColor color.RGBA

	switch {
	case cell.isCenter:
		// Center block - golden color
		fillColor = color.RGBA{255, 215, 0, 255}
	case cell.isStart:
		// Starting points (also safe zones) - use player color with transparency
		fillColor = playerColors[cell.playerIdx]
		fillColor.A = 180
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
		x := boardOffsetX + i*cellSize
		vector.StrokeLine(screen, float32(x), float32(boardOffsetY), float32(x), float32(boardOffsetY+boardSize*cellSize), 2, gridColor, false)
	}

	// Horizontal lines
	for i := 0; i <= boardSize; i++ {
		y := boardOffsetY + i*cellSize
		vector.StrokeLine(screen, float32(boardOffsetX), float32(y), float32(boardOffsetX+boardSize*cellSize), float32(y), 2, gridColor, false)
	}
}

// drawStartMarkers draws X markers on starting points
func (b *Board) drawStartMarkers(screen *ebiten.Image) {
	markerColor := color.RGBA{0, 0, 0, 255} // Black for X marker

	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			cell := b.cells[y][x]
			if cell.isStart {
				cx := boardOffsetX + x*cellSize + cellSize/2
				cy := boardOffsetY + y*cellSize + cellSize/2
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
		board:       NewBoard(),
		conchShells: NewConchShells(),
		rollResult:  0,
		hasRolled:   false,
	}
}

// Update updates the game state
func (g *Game) Update() error {
	// Check for click on roll button
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		// Check if click is in the conch shell area (right side of screen)
		shellAreaX := 600
		shellAreaY := 200
		shellAreaWidth := 150
		shellAreaHeight := 200

		if mx >= shellAreaX && mx <= shellAreaX+shellAreaWidth &&
			my >= shellAreaY && my <= shellAreaY+shellAreaHeight {
			if !g.hasRolled {
				g.rollResult = g.conchShells.Roll()
				g.hasRolled = true
			}
		}
	}

	// Reset roll state when space is pressed (for next turn)
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.hasRolled = false
		g.rollResult = 0
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.RGBA{60, 40, 20, 255})

	// Draw board
	g.board.Draw(screen)

	// Draw conch shell area
	g.drawConchShellArea(screen)

	// Draw title
	ebitenutil.DebugPrintAt(screen, "Ashta Changa - Traditional Indian Board Game", 50, 10)

	// Draw instructions
	ebitenutil.DebugPrintAt(screen, "Click shells to roll", 600, 420)
	ebitenutil.DebugPrintAt(screen, "Press SPACE to reset roll", 600, 440)
}

// drawConchShellArea draws the conch shell rolling area
func (g *Game) drawConchShellArea(screen *ebiten.Image) {
	// Draw shell area background
	shellAreaX := 600
	shellAreaY := 50
	shellAreaWidth := 180
	shellAreaHeight := 400

	// Background for shell area
	vector.DrawFilledRect(screen, float32(shellAreaX), float32(shellAreaY), float32(shellAreaWidth), float32(shellAreaHeight), color.RGBA{139, 90, 43, 255}, false)

	// Border for shell area
	vector.StrokeRect(screen, float32(shellAreaX), float32(shellAreaY), float32(shellAreaWidth), float32(shellAreaHeight), 3, color.RGBA{101, 67, 33, 255}, false)

	// Title
	ebitenutil.DebugPrintAt(screen, "Conch Shells", shellAreaX+45, shellAreaY+10)

	// Draw the 4 conch shells
	shellStartY := shellAreaY + 50
	shellSpacing := 60

	for i := 0; i < 4; i++ {
		shellX := shellAreaX + 40
		shellY := shellStartY + i*shellSpacing

		// Draw shell
		g.drawConchShell(screen, shellX, shellY, g.conchShells.shells[i].openSideUp)
	}

	// Draw roll result
	resultY := shellStartY + 4*shellSpacing + 20
	ebitenutil.DebugPrintAt(screen, "Result:", shellAreaX+60, resultY)

	if g.hasRolled {
		resultText := ""
		switch g.rollResult {
		case 8:
			resultText = "8 (Ashta!)"
		case 4:
			resultText = "4 (Changa!)"
		default:
			resultText = string(rune('0' + g.rollResult))
		}
		ebitenutil.DebugPrintAt(screen, resultText, shellAreaX+70, resultY+20)

		// Show if extra turn
		if g.rollResult == 4 || g.rollResult == 8 {
			ebitenutil.DebugPrintAt(screen, "Extra Turn!", shellAreaX+55, resultY+40)
		}
	}

	// Draw roll button
	buttonY := resultY + 80
	buttonWidth := 140
	buttonHeight := 40
	buttonColor := color.RGBA{210, 180, 140, 255}
	if !g.hasRolled {
		buttonColor = color.RGBA{255, 215, 0, 255} // Gold when clickable
	}
	vector.DrawFilledRect(screen, float32(shellAreaX+20), float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(shellAreaX+20), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 2, color.RGBA{101, 67, 33, 255}, false)
	ebitenutil.DebugPrintAt(screen, "ROLL", shellAreaX+70, buttonY+12)
}

// drawConchShell draws a single conch shell
func (g *Game) drawConchShell(screen *ebiten.Image, x, y int, openSideUp bool) {
	shellSize := 50

	// Draw shell body (oval shape)
	shellColor := color.RGBA{255, 248, 220, 255} // Cornsilk color
	outlineColor := color.RGBA{139, 90, 43, 255} // Brown outline

	// Draw oval shape for shell
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(shellSize), float32(shellSize-10), shellColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(shellSize), float32(shellSize-10), 2, outlineColor, false)

	// Draw indicator of open/closed side
	if openSideUp {
		// Open side up - draw a dot
		centerX := x + shellSize/2
		centerY := y + (shellSize-10)/2
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), 8, color.RGBA{255, 0, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "OPEN", x+8, y+shellSize-8)
	} else {
		// Closed side up - draw an X
		centerX := x + shellSize/2
		centerY := y + (shellSize-10)/2
		vector.StrokeLine(screen, float32(centerX-8), float32(centerY-8), float32(centerX+8), float32(centerY+8), 2, color.RGBA{0, 0, 0, 255}, false)
		vector.StrokeLine(screen, float32(centerX+8), float32(centerY-8), float32(centerX-8), float32(centerY+8), 2, color.RGBA{0, 0, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "CLOSED", x, y+shellSize-8)
	}
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
