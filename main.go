package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 700
	boardSize    = 5
	cellSize     = 100
	boardOffsetX = 50
	boardOffsetY = 50
	numPlayers   = 4
	tokensPerPlayer = 4
)

// Player colors for the 4 players
var playerColors = []color.RGBA{
	{255, 0, 0, 255},   // Red (Top)
	{0, 128, 0, 255},   // Green (Right)
	{0, 0, 255, 255},   // Blue (Bottom)
	{255, 165, 0, 255}, // Orange (Left)
}

var playerNames = []string{"Red", "Green", "Blue", "Orange"}

// TokenState represents the state of a token
type TokenState int

const (
	tokenAtStart TokenState = iota // Token is at starting point (safe house)
	tokenOnBoard                   // Token is moving on the board
	tokenFinished                  // Token has reached the center
)

// Token represents a player's token
type Token struct {
	id        int
	playerIdx int
	state     TokenState
	position  int // Position on the path (0-23 for outer path, 24 for inner path, 25 for center)
}

// Player represents a player in the game
type Player struct {
	idx        int
	tokens     [tokensPerPlayer]*Token
	startPos   int // Starting position on the path for this player
	entryPos   int // Position where token enters inner path
}

// ConchShell represents a single conch shell
type ConchShell struct {
	openSideUp bool
}

// ConchShells represents the 4 conch shells used for rolling
type ConchShells struct {
	shells [4]ConchShell
	result int
}

// Board represents the Ashta Changa game board
type Board struct {
	cells [boardSize][boardSize]*Cell
}

// Cell represents a single cell on the board
type Cell struct {
	x, y      int
	isStart   bool
	isCenter  bool
	isPath    bool
	playerIdx int
}

// Game represents the game state
type Game struct {
	board         *Board
	conchShells   *ConchShells
	players       [numPlayers]*Player
	currentPlayer int
	rollResult    int
	hasRolled     bool
	extraTurn     bool
	gameOver      bool
	winner        int
	clickCooldown int // Prevent multiple clicks
}

// Path coordinates for the outer path (anti-clockwise)
// Total 24 positions on outer path + inner paths
var outerPath = [][2]int{
	// Top row (left to right)
	{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4},
	// Right column (top to bottom, skip top corner)
	{1, 4}, {2, 4}, {3, 4},
	// Bottom row (right to left)
	{4, 4}, {4, 3}, {4, 2}, {4, 1}, {4, 0},
	// Left column (bottom to top, skip bottom corner)
	{3, 0}, {2, 0}, {1, 0},
}

// Inner path coordinates for each player (path to center)
var innerPaths = [][2]int{
	{1, 2}, // Player 0 (Red) - from top
	{2, 3}, // Player 1 (Green) - from right
	{3, 2}, // Player 2 (Blue) - from bottom
	{2, 1}, // Player 3 (Orange) - from left
}

// Starting positions on the path for each player (index into outerPath)
var playerStartPositions = []int{
	2,  // Player 0 (Red) - position 2 on outer path (0, 2)
	6,  // Player 1 (Green) - position 6 on outer path (2, 4)
	10, // Player 2 (Blue) - position 10 on outer path (4, 2)
	14, // Player 3 (Orange) - position 14 on outer path (2, 0)
}

// Entry positions to inner path for each player (index into outerPath)
var playerEntryPositions = []int{
	1,  // Player 0 (Red) - enters inner after position 1 (0, 1)
	5,  // Player 1 (Green) - enters inner after position 5 (1, 4)
	9,  // Player 2 (Blue) - enters inner after position 9 (4, 3)
	13, // Player 3 (Orange) - enters inner after position 13 (3, 0)
}

// NewConchShells creates new conch shells
func NewConchShells() *ConchShells {
	return &ConchShells{}
}

// Roll rolls all 4 conch shells and calculates the result
func (cs *ConchShells) Roll() int {
	openCount := 0
	for i := range cs.shells {
		cs.shells[i].openSideUp = rand.Intn(2) == 1
		if cs.shells[i].openSideUp {
			openCount++
		}
	}

	switch openCount {
	case 0:
		cs.result = 8
	case 1:
		cs.result = 1
	case 2:
		cs.result = 2
	case 3:
		cs.result = 3
	case 4:
		cs.result = 4
	}

	return cs.result
}

// NewPlayer creates a new player
func NewPlayer(idx int) *Player {
	p := &Player{
		idx:      idx,
		startPos: playerStartPositions[idx],
		entryPos: playerEntryPositions[idx],
	}

	for i := 0; i < tokensPerPlayer; i++ {
		p.tokens[i] = &Token{
			id:        i,
			playerIdx: idx,
			state:     tokenAtStart,
			position:  playerStartPositions[idx],
		}
	}

	return p
}

// NewBoard creates a new Ashta Changa board
func NewBoard() *Board {
	b := &Board{}

	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			b.cells[y][x] = &Cell{
				x:         x,
				y:         y,
				playerIdx: -1,
			}
		}
	}

	b.cells[2][2].isCenter = true

	b.cells[0][2].isStart = true
	b.cells[0][2].playerIdx = 0

	b.cells[2][4].isStart = true
	b.cells[2][4].playerIdx = 1

	b.cells[4][2].isStart = true
	b.cells[4][2].playerIdx = 2

	b.cells[2][0].isStart = true
	b.cells[2][0].playerIdx = 3

	for x := 0; x < boardSize; x++ {
		b.cells[0][x].isPath = true
	}
	for y := 1; y < boardSize; y++ {
		b.cells[y][4].isPath = true
	}
	for x := 0; x < boardSize; x++ {
		b.cells[4][x].isPath = true
	}
	for y := 1; y < boardSize-1; y++ {
		b.cells[y][0].isPath = true
	}

	b.cells[1][2].isPath = true
	b.cells[2][3].isPath = true
	b.cells[3][2].isPath = true
	b.cells[2][1].isPath = true

	return b
}

// Draw draws the board
func (b *Board) Draw(screen *ebiten.Image) {
	boardBg := color.RGBA{245, 222, 179, 255}
	vector.DrawFilledRect(screen, float32(boardOffsetX), float32(boardOffsetY), float32(boardSize*cellSize), float32(boardSize*cellSize), boardBg, false)

	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			cell := b.cells[y][x]
			b.drawCell(screen, cell)
		}
	}

	b.drawGrid(screen)
	b.drawStartMarkers(screen)
}

// drawCell draws a single cell
func (b *Board) drawCell(screen *ebiten.Image, cell *Cell) {
	x := boardOffsetX + cell.x*cellSize
	y := boardOffsetY + cell.y*cellSize

	var fillColor color.RGBA

	switch {
	case cell.isCenter:
		fillColor = color.RGBA{255, 215, 0, 255}
	case cell.isStart:
		fillColor = playerColors[cell.playerIdx]
		fillColor.A = 180
	case cell.isPath:
		fillColor = color.RGBA{210, 180, 140, 255}
	default:
		fillColor = color.RGBA{139, 90, 43, 255}
	}

	vector.DrawFilledRect(screen, float32(x+2), float32(y+2), float32(cellSize-4), float32(cellSize-4), fillColor, false)
}

// drawGrid draws the grid lines
func (b *Board) drawGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{101, 67, 33, 255}

	for i := 0; i <= boardSize; i++ {
		x := boardOffsetX + i*cellSize
		vector.StrokeLine(screen, float32(x), float32(boardOffsetY), float32(x), float32(boardOffsetY+boardSize*cellSize), 2, gridColor, false)
	}

	for i := 0; i <= boardSize; i++ {
		y := boardOffsetY + i*cellSize
		vector.StrokeLine(screen, float32(boardOffsetX), float32(y), float32(boardOffsetX+boardSize*cellSize), float32(y), 2, gridColor, false)
	}
}

// drawStartMarkers draws X markers on starting points
func (b *Board) drawStartMarkers(screen *ebiten.Image) {
	markerColor := color.RGBA{0, 0, 0, 255}

	for y := 0; y < boardSize; y++ {
		for x := 0; x < boardSize; x++ {
			cell := b.cells[y][x]
			if cell.isStart {
				cx := boardOffsetX + x*cellSize + cellSize/2
				cy := boardOffsetY + y*cellSize + cellSize/2
				offset := 25

				vector.StrokeLine(screen, float32(cx-offset), float32(cy-offset), float32(cx+offset), float32(cy+offset), 3, markerColor, false)
				vector.StrokeLine(screen, float32(cx+offset), float32(cy-offset), float32(cx-offset), float32(cy+offset), 3, markerColor, false)
			}
		}
	}
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		board:       NewBoard(),
		conchShells: NewConchShells(),
		rollResult:  0,
		hasRolled:   false,
		extraTurn:   false,
		gameOver:    false,
		winner:      -1,
	}

	for i := 0; i < numPlayers; i++ {
		g.players[i] = NewPlayer(i)
	}

	return g
}

// canMoveToken checks if a token can be moved with the current roll
func (g *Game) canMoveToken(token *Token, roll int) bool {
	if token.state == tokenFinished {
		return false
	}

	// Token at start - needs 1, 4, or 8 to move out
	if token.state == tokenAtStart {
		return roll == 1 || roll == 4 || roll == 8
	}

	// Token on board - check if move is valid
	if token.state == tokenOnBoard {
		player := g.players[token.playerIdx]
		
		// Check if token is on inner path
		if token.position >= 24 {
			// On inner path, can only move if roll gets to center (position 25)
			if token.position == 24 {
				return roll == 1 // Need exactly 1 to reach center
			}
			return false
		}

		// On outer path - check if would pass entry point
		// If we've gone around and are approaching entry
		stepsToEntry := (player.entryPos - token.position + len(outerPath)) % len(outerPath)
		
		if stepsToEntry <= roll && stepsToEntry > 0 {
			// We can enter inner path
			remaining := roll - stepsToEntry
			if remaining <= 1 { // Can reach center with remaining steps
				return true
			}
			return false
		}
		
		return true
	}

	return false
}

// getValidTokens returns indices of tokens that can be moved
func (g *Game) getValidTokens() []int {
	var valid []int
	player := g.players[g.currentPlayer]

	for i, token := range player.tokens {
		if g.canMoveToken(token, g.rollResult) {
			valid = append(valid, i)
		}
	}

	return valid
}

// moveToken moves a token based on the roll result
func (g *Game) moveToken(tokenIdx int) {
	player := g.players[g.currentPlayer]
	token := player.tokens[tokenIdx]
	roll := g.rollResult

	if token.state == tokenAtStart {
		// Move token out of starting point
		token.state = tokenOnBoard
		token.position = (player.startPos + roll) % len(outerPath)
		// Check for killing opponent token
		g.checkKill(token)
	} else if token.state == tokenOnBoard {
		// Check if entering inner path
		stepsToEntry := (player.entryPos - token.position + len(outerPath)) % len(outerPath)
		
		if stepsToEntry <= roll && stepsToEntry > 0 {
			// Enter inner path
			remaining := roll - stepsToEntry
			if remaining == 0 {
				token.position = 24 // Inner path position
			} else if remaining == 1 {
				token.state = tokenFinished
				token.position = 25 // Center
				g.checkWin()
			}
		} else {
			// Normal movement on outer path
			token.position = (token.position + roll) % len(outerPath)
			
			// Check for killing opponent token
			g.checkKill(token)
		}
	}
}

// checkKill checks if the moved token lands on an opponent and kills them
func (g *Game) checkKill(token *Token) {
	if token.state != tokenOnBoard || token.position >= len(outerPath) {
		return
	}

	// Rule: Starting blocks are safe houses. No killing can happen on a safe house.
	for _, startPos := range playerStartPositions {
		if token.position == startPos {
			return // Safe house, no kill
		}
	}

	// Rule: If landing on an opponent's token in a non-safe block, kill it.
	killedAny := false
	for _, player := range g.players {
		if player.idx == token.playerIdx {
			continue // Cannot kill own tokens
		}
		for _, opponentToken := range player.tokens {
			if opponentToken.state == tokenOnBoard && opponentToken.position == token.position {
				// Kill opponent token and send it back to starting point
				opponentToken.state = tokenAtStart
				opponentToken.position = playerStartPositions[opponentToken.playerIdx]
				killedAny = true
			}
		}
	}

	if killedAny {
		g.extraTurn = true // Rule: Killing an opponent's token grants an extra turn
	}
}

// checkWin checks if current player has won
func (g *Game) checkWin() {
	player := g.players[g.currentPlayer]
	finishedCount := 0
	for _, token := range player.tokens {
		if token.state == tokenFinished {
			finishedCount++
		}
	}
	if finishedCount == tokensPerPlayer {
		g.gameOver = true
		g.winner = g.currentPlayer
	}
}

// nextTurn advances to the next player
func (g *Game) nextTurn() {
	if g.extraTurn {
		g.extraTurn = false
	} else {
		g.currentPlayer = (g.currentPlayer + 1) % numPlayers
	}
	g.hasRolled = false
	g.rollResult = 0
}

// getCellCoordinates returns the board grid coordinates for a token
func (g *Game) getCellCoordinates(token *Token) (int, int) {
	if token.state == tokenFinished {
		return 2, 2 // Center block
	}

	if token.position >= 24 {
		// Inner path
		player := g.players[token.playerIdx]
		coords := innerPaths[player.idx]
		return coords[0], coords[1]
	}

	// Outer path (includes starting points)
	coords := outerPath[token.position]
	return coords[0], coords[1]
}

// Update updates the game state
func (g *Game) Update() error {
	if g.gameOver {
		return nil
	}

	// Decrease click cooldown
	if g.clickCooldown > 0 {
		g.clickCooldown--
		return nil
	}

	// Handle mouse click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		// Check roll button click
		shellAreaX := 600
		buttonY := 450
		buttonWidth := 140
		buttonHeight := 40

		if mx >= shellAreaX+20 && mx <= shellAreaX+20+buttonWidth &&
			my >= buttonY && my <= buttonY+buttonHeight {
			if !g.hasRolled {
				g.rollResult = g.conchShells.Roll()
				g.hasRolled = true
				g.clickCooldown = 10

				// Rule: Rolling a 4 or 8 grants an extra turn
				if g.rollResult == 4 || g.rollResult == 8 {
					g.extraTurn = true
				}

				// Check if player has any valid moves
				validTokens := g.getValidTokens()
				if len(validTokens) == 0 {
					// No valid moves, auto skip turn
					g.nextTurn()
				}
			}
		}

		// Check token click (only after rolling)
		if g.hasRolled {
			validTokens := g.getValidTokens()
			player := g.players[g.currentPlayer]

			for _, idx := range validTokens {
				token := player.tokens[idx]
				
				// Calculate hit area for token (this is tricky with side-by-side)
				// For simplicity, we'll check if the click is in the cell where the token is
				row, col := g.getCellCoordinates(token)
				tx := boardOffsetX + col*cellSize
				ty := boardOffsetY + row*cellSize
				
				if mx >= tx && mx <= tx+cellSize && my >= ty && my <= ty+cellSize {
					// Check if this token is actually one of the valid ones
					g.moveToken(idx)
					g.clickCooldown = 10

					if !g.gameOver {
						g.nextTurn()
					}
					break
				}
			}
		}
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{60, 40, 20, 255})

	g.board.Draw(screen)

	// Draw tokens
	g.drawTokens(screen)

	// Draw conch shell area
	g.drawConchShellArea(screen)

	// Draw current player indicator
	g.drawPlayerInfo(screen)

	// Draw title
	ebitenutil.DebugPrintAt(screen, "Ashta Changa - Traditional Indian Board Game", 50, 10)

	// Draw game over message
	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER!", 250, 280)
		ebitenutil.DebugPrintAt(screen, playerNames[g.winner]+" Wins!", 250, 300)
	}
}

// drawTokens draws all tokens on the board
func (g *Game) drawTokens(screen *ebiten.Image) {
	// Group tokens by cell position
	type pos struct{ r, c int }
	tokensByCell := make(map[pos][]*Token)

	for _, player := range g.players {
		for _, token := range player.tokens {
			if token.state == tokenFinished && token.playerIdx != g.winner {
				// Only show winner tokens in center or show all? 
				// Let's show all finished tokens in center
			}
			r, c := g.getCellCoordinates(token)
			p := pos{r, c}
			tokensByCell[p] = append(tokensByCell[p], token)
		}
	}

	// Draw tokens in each cell
	for cellPos, tokens := range tokensByCell {
		cellX := boardOffsetX + cellPos.c*cellSize
		cellY := boardOffsetY + cellPos.r*cellSize

		// Calculate positions for tokens within the cell (up to 16 tokens possible in safe house!)
		// We'll use a 4x4 grid if needed, or just offset them
		for i, token := range tokens {
			// Offset within cell
			// For 5x5 board and 100px cells, we can fit tokens nicely
			numInRow := 2
			if len(tokens) > 4 {
				numInRow = 3
			}
			if len(tokens) > 9 {
				numInRow = 4
			}

			offsetX := (i % numInRow) * (cellSize / (numInRow + 1)) + (cellSize / (numInRow + 1))
			offsetY := (i / numInRow) * (cellSize / (numInRow + 1)) + (cellSize / (numInRow + 1))

			tx := cellX + offsetX
			ty := cellY + offsetY

			tokenColor := playerColors[token.playerIdx]
			
			// Highlight valid tokens
			isValid := false
			if g.hasRolled && token.playerIdx == g.currentPlayer {
				for _, idx := range g.getValidTokens() {
					if g.players[g.currentPlayer].tokens[idx] == token {
						isValid = true
						break
					}
				}
			}

			radius := float32(12)
			if numInRow > 2 {
				radius = 8
			}

			if isValid {
				vector.StrokeCircle(screen, float32(tx), float32(ty), radius+4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
			}

			vector.DrawFilledCircle(screen, float32(tx), float32(ty), radius, tokenColor, false)
			vector.StrokeCircle(screen, float32(tx), float32(ty), radius, 2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)
			
			// Draw token number if not too small
			if radius >= 10 {
				numStr := string(rune('1' + token.id))
				ebitenutil.DebugPrintAt(screen, numStr, tx-3, ty-6)
			}
		}
	}
}

// drawPlayerInfo draws current player and turn information
func (g *Game) drawPlayerInfo(screen *ebiten.Image) {
	infoX := 600
	infoY := 550

	// Current player
	ebitenutil.DebugPrintAt(screen, "Current Player:", infoX, infoY)
	playerColor := playerColors[g.currentPlayer]
	vector.DrawFilledRect(screen, float32(infoX+110), float32(infoY), 60, 20, playerColor, false)
	ebitenutil.DebugPrintAt(screen, playerNames[g.currentPlayer], infoX+115, infoY+2)

	// Extra turn indicator
	if g.extraTurn {
		ebitenutil.DebugPrintAt(screen, "EXTRA TURN!", infoX, infoY+25)
	}

	// Instructions
	if g.hasRolled {
		validTokens := g.getValidTokens()
		if len(validTokens) > 0 {
			ebitenutil.DebugPrintAt(screen, "Click a cell with your", infoX, infoY+50)
			ebitenutil.DebugPrintAt(screen, "highlighted token", infoX, infoY+65)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "Click ROLL to roll shells", infoX, infoY+50)
	}
}

// drawConchShellArea draws the conch shell rolling area
func (g *Game) drawConchShellArea(screen *ebiten.Image) {
	shellAreaX := 600
	shellAreaY := 50
	shellAreaWidth := 180
	shellAreaHeight := 380

	vector.DrawFilledRect(screen, float32(shellAreaX), float32(shellAreaY), float32(shellAreaWidth), float32(shellAreaHeight), color.RGBA{139, 90, 43, 255}, false)
	vector.StrokeRect(screen, float32(shellAreaX), float32(shellAreaY), float32(shellAreaWidth), float32(shellAreaHeight), 3, color.RGBA{101, 67, 33, 255}, false)

	ebitenutil.DebugPrintAt(screen, "Conch Shells", shellAreaX+45, shellAreaY+10)

	shellStartY := shellAreaY + 50
	shellSpacing := 60

	for i := 0; i < 4; i++ {
		shellX := shellAreaX + 40
		shellY := shellStartY + i*shellSpacing
		g.drawConchShell(screen, shellX, shellY, g.conchShells.shells[i].openSideUp)
	}

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

		if g.rollResult == 4 || g.rollResult == 8 {
			ebitenutil.DebugPrintAt(screen, "Bonus Turn!", shellAreaX+55, resultY+40)
		}
	}

	buttonY := resultY + 80
	buttonWidth := 140
	buttonHeight := 40
	buttonColor := color.RGBA{210, 180, 140, 255}
	if !g.hasRolled {
		buttonColor = color.RGBA{255, 215, 0, 255}
	}
	vector.DrawFilledRect(screen, float32(shellAreaX+20), float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(shellAreaX+20), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 2, color.RGBA{101, 67, 33, 255}, false)
	ebitenutil.DebugPrintAt(screen, "ROLL", shellAreaX+70, buttonY+12)
}

// drawConchShell draws a single conch shell
func (g *Game) drawConchShell(screen *ebiten.Image, x, y int, openSideUp bool) {
	shellSize := 50

	shellColor := color.RGBA{255, 248, 220, 255}
	outlineColor := color.RGBA{139, 90, 43, 255}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(shellSize), float32(shellSize-10), shellColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(shellSize), float32(shellSize-10), 2, outlineColor, false)

	if openSideUp {
		centerX := x + shellSize/2
		centerY := y + (shellSize-10)/2
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), 8, color.RGBA{255, 0, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "OPEN", x+8, y+shellSize-8)
	} else {
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
