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

// GameState represents the current state of the game
type GameState int

const (
	stateSelectingPlayers GameState = iota
	statePlaying
	stateGameOver
)

// Game represents the game state
type Game struct {
	state            GameState
	board            *Board
	conchShells      *ConchShells
	players          [4]*Player
	numActivePlayers int
	currentPlayer    int
	rollResult       int
	hasRolled        bool
	extraTurn        bool
	winner           int
	clickCooldown    int // Prevent multiple clicks
}

// Path coordinates for the outer path (anti-clockwise)
// Total 16 positions on outer path
var outerPath = [][2]int{
	{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4},
	{1, 4}, {2, 4}, {3, 4},
	{4, 4}, {4, 3}, {4, 2}, {4, 1}, {4, 0},
	{3, 0}, {2, 0}, {1, 0},
}

// Inner path blocks in clockwise order
var innerCircle = [][2]int{
	{1, 1}, {1, 2}, {1, 3}, {2, 3}, {3, 3}, {3, 2}, {3, 1}, {2, 1},
}

// Starting indices in innerCircle for each player
var playerInnerStartIndices = []int{0, 2, 4, 6}

// Starting positions on the path for each player (index into outerPath)
var playerStartPositions = []int{
	2,  // Player 0 (Red) - (0, 2)
	6,  // Player 1 (Green) - (2, 4)
	10, // Player 2 (Blue) - (4, 2)
	14, // Player 3 (Orange) - (2, 0)
}

// Entry positions to inner path for each player (index into outerPath)
// This is the block JUST BEFORE the starting position
var playerEntryPositions = []int{
	1,  // Player 0 (Red) - (0, 1)
	5,  // Player 1 (Green) - (1, 4)
	9,  // Player 2 (Blue) - (4, 3)
	13, // Player 3 (Orange) - (3, 0)
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

	// Mark inner circle as path
	for _, coords := range innerCircle {
		b.cells[coords[0]][coords[1]].isPath = true
	}

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
	return &Game{
		state:       stateSelectingPlayers,
		board:       NewBoard(),
		conchShells: NewConchShells(),
		winner:      -1,
	}
}

// initGame initializes the game with selected number of players
func (g *Game) initGame(num int) {
	g.numActivePlayers = num
	for i := 0; i < 4; i++ {
		if i < num {
			g.players[i] = NewPlayer(i)
		} else {
			g.players[i] = nil
		}
	}
	g.state = statePlaying
	g.currentPlayer = 0
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
		// position 16-23: steps 0-7 on inner circle
		if token.position >= 16 {
			step := token.position - 16
			if step + roll <= 8 { // 8 means center
				return true
			}
			return false
		}

		// On outer path
		// Check if would pass entry point
		distToEntry := (player.entryPos - token.position + 16) % 16
		
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
	if player == nil {
		return
	}
	token := player.tokens[tokenIdx]
	roll := g.rollResult

	if token.state == tokenAtStart {
		token.state = tokenOnBoard
		g.applyMove(token, roll)
	} else if token.state == tokenOnBoard {
		g.applyMove(token, roll)
	}
}

// applyMove applies the roll to a token, handling transitions
func (g *Game) applyMove(token *Token, roll int) {
	player := g.players[token.playerIdx]

	if token.position >= 16 {
		// On inner path
		step := token.position - 16
		newStep := step + roll
		if newStep >= 8 {
			token.state = tokenFinished
			token.position = 24
			g.extraTurn = true // Rule: Reaching center grants a bonus roll
			g.checkWin()
		} else {
			token.position = 16 + newStep
			g.checkKill(token) // Rule: Killing possible in inner circle too
		}
		return
	}

	// On outer path
	distToEntry := (player.entryPos - token.position + 16) % 16
	if roll > distToEntry {
		// Transition to inner path
		remaining := roll - distToEntry - 1
		if remaining >= 8 {
			token.state = tokenFinished
			token.position = 24
			g.extraTurn = true // Rule: Reaching center grants a bonus roll
			g.checkWin()
		} else {
			token.position = 16 + remaining
			g.checkKill(token) // Rule: Killing possible in inner circle too
		}
	} else {
		// Stay on outer path
		token.position = (token.position + roll) % 16
		g.checkKill(token)
	}
}

// checkKill checks if the moved token lands on an opponent and kills them
func (g *Game) checkKill(token *Token) {
	if token.state != tokenOnBoard {
		return
	}

	// Get current grid coordinates
	row, col := g.getCellCoordinates(token)

	// Rule: Starting blocks are safe houses. No killing can happen on a safe house.
	for i := 0; i < 4; i++ {
		startCoords := outerPath[playerStartPositions[i]]
		if row == startCoords[0] && col == startCoords[1] {
			return // Safe house, no kill
		}
	}

	// Rule: If landing on an opponent's token in a non-safe block, kill it.
	killedAny := false
	for _, player := range g.players {
		if player == nil || player.idx == token.playerIdx {
			continue // Cannot kill own tokens or from non-existent players
		}
		for _, opponentToken := range player.tokens {
			if opponentToken.state == tokenOnBoard {
				or, oc := g.getCellCoordinates(opponentToken)
				if or == row && oc == col {
					// Kill opponent token and send it back to starting point
					opponentToken.state = tokenAtStart
					opponentToken.position = playerStartPositions[opponentToken.playerIdx]
					killedAny = true
				}
			}
		}
	}

	if killedAny {
		g.extraTurn = true 
	}
}

// checkWin checks if current player has won
func (g *Game) checkWin() {
	player := g.players[g.currentPlayer]
	if player == nil {
		return
	}
	finishedCount := 0
	for _, token := range player.tokens {
		if token.state == tokenFinished {
			finishedCount++
		}
	}
	if finishedCount == tokensPerPlayer {
		g.state = stateGameOver
		g.winner = g.currentPlayer
	}
}

// nextTurn advances to the next player
func (g *Game) nextTurn() {
	if g.extraTurn {
		g.extraTurn = false
	} else {
		g.currentPlayer = (g.currentPlayer + 1) % g.numActivePlayers
	}
	g.hasRolled = false
	g.rollResult = 0
}

// getCellCoordinates returns the board grid coordinates for a token
func (g *Game) getCellCoordinates(token *Token) (int, int) {
	if token.state == tokenFinished {
		return 2, 2 // Center block
	}

	if token.position >= 16 {
		// Inner path
		player := g.players[token.playerIdx]
		step := token.position - 16
		innerIdx := (playerInnerStartIndices[player.idx] + step) % 8
		coords := innerCircle[innerIdx]
		return coords[0], coords[1]
	}

	// Outer path
	coords := outerPath[token.position]
	return coords[0], coords[1]
}

// Update updates the game state
func (g *Game) Update() error {
	// Decrease click cooldown
	if g.clickCooldown > 0 {
		g.clickCooldown--
	}

	if g.state == stateSelectingPlayers {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
			mx, my := ebiten.CursorPosition()
			// Selection buttons
			for i := 1; i <= 4; i++ {
				bx := screenWidth/2 - 100
				by := 200 + i*60
				if mx >= bx && mx <= bx+200 && my >= by && my <= by+40 {
					g.initGame(i)
					g.clickCooldown = 20
					break
				}
			}
		}
		return nil
	}

	if g.state == stateGameOver {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
			g.state = stateSelectingPlayers
			g.clickCooldown = 20
		}
		return nil
	}

	// Handle mouse click during playing
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
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

				if g.rollResult == 4 || g.rollResult == 8 {
					g.extraTurn = true
				}

				validTokens := g.getValidTokens()
				if len(validTokens) == 0 {
					g.nextTurn()
				}
			}
		}

		if g.hasRolled {
			validTokens := g.getValidTokens()
			player := g.players[g.currentPlayer]

			for _, idx := range validTokens {
				token := player.tokens[idx]
				row, col := g.getCellCoordinates(token)
				tx := boardOffsetX + col*cellSize
				ty := boardOffsetY + row*cellSize
				
				if mx >= tx && mx <= tx+cellSize && my >= ty && my <= ty+cellSize {
					g.moveToken(idx)
					g.clickCooldown = 10
					if g.state == statePlaying {
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

	if g.state == stateSelectingPlayers {
		g.drawSelectionScreen(screen)
		return
	}

	g.board.Draw(screen)
	g.drawTokens(screen)
	g.drawConchShellArea(screen)
	g.drawPlayerInfo(screen)

	ebitenutil.DebugPrintAt(screen, "Ashta Changa - Traditional Indian Board Game", 50, 10)

	if g.state == stateGameOver {
		vector.DrawFilledRect(screen, 200, 250, 400, 100, color.RGBA{0, 0, 0, 200}, false)
		ebitenutil.DebugPrintAt(screen, "GAME OVER!", 350, 280)
		ebitenutil.DebugPrintAt(screen, playerNames[g.winner]+" Wins!", 350, 300)
		ebitenutil.DebugPrintAt(screen, "Click anywhere to restart", 320, 330)
	}
}

// drawSelectionScreen draws the player selection UI
func (g *Game) drawSelectionScreen(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ASHTA CHANGA", screenWidth/2-50, 100)
	ebitenutil.DebugPrintAt(screen, "Select Number of Players", screenWidth/2-80, 150)

	for i := 1; i <= 4; i++ {
		bx := screenWidth/2 - 100
		by := 200 + i*60
		vector.DrawFilledRect(screen, float32(bx), float32(by), 200, 40, color.RGBA{139, 90, 43, 255}, false)
		vector.StrokeRect(screen, float32(bx), float32(by), 200, 40, 2, color.RGBA{210, 180, 140, 255}, false)
		ebitenutil.DebugPrintAt(screen, string(rune('0'+i))+" Players", bx+60, by+12)
	}
}

// drawTokens draws all tokens on the board
func (g *Game) drawTokens(screen *ebiten.Image) {
	type pos struct{ r, c int }
	tokensByCell := make(map[pos][]*Token)

	for _, player := range g.players {
		if player == nil {
			continue
		}
		for _, token := range player.tokens {
			r, c := g.getCellCoordinates(token)
			p := pos{r, c}
			tokensByCell[p] = append(tokensByCell[p], token)
		}
	}

	for cellPos, tokens := range tokensByCell {
		cellX := boardOffsetX + cellPos.c*cellSize
		cellY := boardOffsetY + cellPos.r*cellSize

		for i, token := range tokens {
			numInRow := 2
			if len(tokens) > 4 { numInRow = 3 }
			if len(tokens) > 9 { numInRow = 4 }

			offsetX := (i % numInRow) * (cellSize / (numInRow + 1)) + (cellSize / (numInRow + 1))
			offsetY := (i / numInRow) * (cellSize / (numInRow + 1)) + (cellSize / (numInRow + 1))

			tx := cellX + offsetX
			ty := cellY + offsetY

			tokenColor := playerColors[token.playerIdx]
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
			if numInRow > 2 { radius = 8 }

			if isValid {
				vector.StrokeCircle(screen, float32(tx), float32(ty), radius+4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
			}

			vector.DrawFilledCircle(screen, float32(tx), float32(ty), radius, tokenColor, false)
			vector.StrokeCircle(screen, float32(tx), float32(ty), radius, 2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)
			
			if radius >= 10 {
				numStr := string(rune('1' + token.id))
				ebitenutil.DebugPrintAt(screen, numStr, tx-3, ty-6)
			}
		}
	}
}

// drawPlayerInfo draws current player and turn information
func (g *Game) drawPlayerInfo(screen *ebiten.Image) {
	if g.players[g.currentPlayer] == nil {
		return
	}
	
	infoX := 600
	infoY := 550

	ebitenutil.DebugPrintAt(screen, "Current Player:", infoX, infoY)
	playerColor := playerColors[g.currentPlayer]
	vector.DrawFilledRect(screen, float32(infoX+110), float32(infoY), 60, 20, playerColor, false)
	ebitenutil.DebugPrintAt(screen, playerNames[g.currentPlayer], infoX+115, infoY+2)

	if g.extraTurn {
		ebitenutil.DebugPrintAt(screen, "EXTRA TURN!", infoX, infoY+25)
	}

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
