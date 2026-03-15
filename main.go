package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth     = 800
	screenHeight    = 700
	boardSize       = 5
	cellSize        = 100
	boardOffsetX    = 50
	boardOffsetY    = 50
	numPlayers      = 4
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

// Multiplayer related types and constants
const (
	stateMenu GameState = iota
	stateEnterName
	stateCreateOrJoin
	stateWaitingRoom
	stateConnecting
)

// NetworkClient handles WebSocket connection to the server
type NetworkClient struct {
	ws         js.Value
	serverURL  string
	playerID   string
	playerName string
	roomID     string
	connected  bool
	mu         sync.RWMutex

	// Callbacks
	onRoomCreated     func(room *NetworkRoom, player *NetworkPlayer)
	onRoomJoined      func(room *NetworkRoom, player *NetworkPlayer)
	onPlayerJoined    func(player *NetworkPlayer)
	onPlayerLeft      func(player *NetworkPlayer)
	onGameStarted     func(gameState *ServerGameState)
	onGameStateUpdate func(gameState *ServerGameState)
	onError           func(msg string)
}

// NetworkRoom represents a game room from server
type NetworkRoom struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	HostID      string           `json:"hostId"`
	Players     []*NetworkPlayer `json:"players"`
	MaxPlayers  int              `json:"maxPlayers"`
	GameStarted bool             `json:"gameStarted"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// NetworkPlayer represents a player from server
type NetworkPlayer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RoomID    string `json:"roomId"`
	PlayerIdx int    `json:"playerIdx"`
	IsHost    bool   `json:"isHost"`
	IsReady   bool   `json:"isReady"`
}

// ServerGameState represents game state from server
type ServerGameState struct {
	CurrentPlayer    int            `json:"currentPlayer"`
	RollResult       int            `json:"rollResult"`
	HasRolled        bool           `json:"hasRolled"`
	ShellStates      []bool         `json:"shellStates"`
	ExtraTurn        bool           `json:"extraTurn"`
	Winner           int            `json:"winner"`
	PlayerTokens     []PlayerTokens `json:"playerTokens"`
	NumActivePlayers int            `json:"numActivePlayers"`
}

// ServerTokenState represents a single token's state from server
type ServerTokenState struct {
	ID       int `json:"id"`
	State    int `json:"state"`
	Position int `json:"position"`
}

// PlayerTokens represents a player's tokens state from server
type PlayerTokens struct {
	PlayerIdx int                `json:"playerIdx"`
	Tokens    []ServerTokenState `json:"tokens"`
}

// Message represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewNetworkClient creates a new network client
func NewNetworkClient(serverURL string) *NetworkClient {
	return &NetworkClient{
		serverURL: serverURL,
	}
}

// Connect connects to the WebSocket server
func (nc *NetworkClient) Connect() error {
	// Use browser's WebSocket API via syscall/js
	jsWS := js.Global().Get("WebSocket")
	if jsWS.IsUndefined() {
		return fmt.Errorf("WebSockets not supported in this browser")
	}

	nc.ws = jsWS.New(nc.serverURL)

	// Set up event handlers
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		nc.mu.Lock()
		nc.connected = true
		nc.mu.Unlock()
		log.Println("WebSocket connected")
		return nil
	})

	onMessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		data := args[0].Get("data").String()
		var msg WSMessage
		if err := json.Unmarshal([]byte(data), &msg); err == nil {
			nc.handleMessage(msg)
		}
		return nil
	})

	onClose := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		nc.mu.Lock()
		nc.connected = false
		nc.mu.Unlock()
		if nc.onError != nil {
			nc.onError("Disconnected from server")
		}
		return nil
	})

	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if nc.onError != nil {
			nc.onError("Connection error")
		}
		return nil
	})

	nc.ws.Set("onopen", onOpen)
	nc.ws.Set("onmessage", onMessage)
	nc.ws.Set("onclose", onClose)
	nc.ws.Set("onerror", onError)

	// Wait for connection to open (max 5 seconds)
	start := time.Now()
	for {
		nc.mu.RLock()
		connected := nc.connected
		nc.mu.RUnlock()
		if connected {
			break
		}
		if time.Since(start) > 5*time.Second {
			return fmt.Errorf("connection timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// Disconnect closes the connection
func (nc *NetworkClient) Disconnect() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if !nc.ws.IsUndefined() {
		nc.ws.Call("close")
		nc.connected = false
	}
}

// IsConnected returns connection status
func (nc *NetworkClient) IsConnected() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.connected
}

// handleMessage handles incoming messages
func (nc *NetworkClient) handleMessage(msg WSMessage) {
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return
	}
	log.Printf("Received message type: %s", msg.Type)

	switch msg.Type {
	case "roomCreated":
		var resp struct {
			Room   *NetworkRoom   `json:"room"`
			Player *NetworkPlayer `json:"player"`
		}
		if err := json.Unmarshal(payloadBytes, &resp); err == nil {
			nc.playerID = resp.Player.ID
			nc.roomID = resp.Room.ID
			if nc.onRoomCreated != nil {
				nc.onRoomCreated(resp.Room, resp.Player)
			}
		}

	case "roomJoined":
		var resp struct {
			Success bool           `json:"success"`
			Room    *NetworkRoom   `json:"room"`
			Player  *NetworkPlayer `json:"player"`
			Error   string         `json:"error"`
		}
		if err := json.Unmarshal(payloadBytes, &resp); err == nil {
			if resp.Success {
				nc.playerID = resp.Player.ID
				nc.roomID = resp.Room.ID
				if nc.onRoomJoined != nil {
					nc.onRoomJoined(resp.Room, resp.Player)
				}
			} else if nc.onError != nil {
				nc.onError(resp.Error)
			}
		}

	case "playerJoined":
		var payload struct {
			Player *NetworkPlayer `json:"player"`
			Room   *NetworkRoom   `json:"room"`
		}
		if err := json.Unmarshal(payloadBytes, &payload); err == nil && nc.onPlayerJoined != nil {
			nc.onPlayerJoined(payload.Player)
		}

	case "playerLeft":
		var player NetworkPlayer
		if err := json.Unmarshal(payloadBytes, &player); err == nil && nc.onPlayerLeft != nil {
			nc.onPlayerLeft(&player)
		}

	case "gameStarted":
		var gameState ServerGameState
		if err := json.Unmarshal(payloadBytes, &gameState); err != nil {
			log.Printf("Error unmarshaling gameStarted: %v", err)
		} else if nc.onGameStarted != nil {
			log.Printf("Calling onGameStarted with %d players", gameState.NumActivePlayers)
			nc.onGameStarted(&gameState)
		}

	case "gameStateUpdate":
		var gameState ServerGameState
		if err := json.Unmarshal(payloadBytes, &gameState); err != nil {
			log.Printf("Error unmarshaling gameStateUpdate: %v", err)
		} else if nc.onGameStateUpdate != nil {
			nc.onGameStateUpdate(&gameState)
		}

	case "error":
		var payload struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(payloadBytes, &payload); err == nil && nc.onError != nil {
			nc.onError(payload.Message)
		}
	}
}

// send sends a message to the server
func (nc *NetworkClient) send(msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	nc.ws.Call("send", string(data))
	return nil
}

// CreateRoom creates a new room
func (nc *NetworkClient) CreateRoom(roomName, playerName string) {
	nc.playerName = playerName
	nc.send(WSMessage{
		Type: "createRoom",
		Payload: map[string]string{
			"roomName":   roomName,
			"playerName": playerName,
		},
	})
}

// JoinRoom joins an existing room
func (nc *NetworkClient) JoinRoom(roomID, playerName string, playerID string) {
	nc.playerName = playerName
	nc.roomID = roomID
	nc.send(WSMessage{
		Type: "joinRoom",
		Payload: map[string]string{
			"roomId":     roomID,
			"playerName": playerName,
			"playerId":   playerID,
		},
	})
}

// StartGame starts the game (host only)
func (nc *NetworkClient) StartGame() {
	nc.send(WSMessage{
		Type: "startGame",
		Payload: map[string]string{
			"roomId": nc.roomID,
		},
	})
}

// SendGameAction sends a game action to the server
func (nc *NetworkClient) SendGameAction(action string, tokenIdx int) {
	nc.send(WSMessage{
		Type: "gameAction",
		Payload: map[string]interface{}{
			"roomId":   nc.roomID,
			"action":   action,
			"tokenIdx": tokenIdx,
		},
	})
}

// SendRollAction sends a roll action request to the server
func (nc *NetworkClient) SendRollAction() {
	nc.send(WSMessage{
		Type: "gameAction",
		Payload: map[string]interface{}{
			"roomId":   nc.roomID,
			"action":   "roll",
			"tokenIdx": -1,
		},
	})
}

// SetReady sets player ready status
func (nc *NetworkClient) SetReady(isReady bool) {
	nc.send(WSMessage{
		Type: "setReady",
		Payload: map[string]interface{}{
			"roomId":  nc.roomID,
			"isReady": isReady,
		},
	})
}

// getServerURL gets the WebSocket server URL based on current location
func getServerURL() string {
	// Check if running in browser
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		// Not in browser, return default
		return "ws://localhost:8080/ws"
	}

	// Get current location
	location := js.Global().Get("window").Get("location")
	protocol := location.Get("protocol").String()
	host := location.Get("host").String()

	// Convert http/https to ws/wss
	wsProtocol := "ws"
	if protocol == "https:" {
		wsProtocol = "wss"
	}

	// If we're running on localhost (common for development),
	// assume the backend is on port 8080 as specified in README.
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		hostName := strings.Split(host, ":")[0]
		return fmt.Sprintf("%s://%s:8080/ws", wsProtocol, hostName)
	}

	return fmt.Sprintf("%s://%s/ws", wsProtocol, host)
}

// getAPIURL gets the REST API URL based on current location
func getAPIURL() string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return "http://localhost:8080/api/rooms"
	}

	location := js.Global().Get("window").Get("location")
	protocol := location.Get("protocol").String()
	host := location.Get("host").String()

	// If we're running on localhost, assume the backend is on port 8080
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		hostName := strings.Split(host, ":")[0]
		return fmt.Sprintf("http://%s:8080/api/rooms", hostName)
	}

	return fmt.Sprintf("%s//%s/api/rooms", protocol, host)
}

// getRoomFromURL gets room ID from URL query parameters
func getRoomFromURL() string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return ""
	}

	location := js.Global().Get("window").Get("location")
	search := location.Get("search").String()

	// Parse query string
	if strings.HasPrefix(search, "?") {
		search = search[1:]
	}

	pairs := strings.Split(search, "&")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) == 2 && parts[0] == "room" {
			return parts[1]
		}
	}

	return ""
}

// getRoomURL returns the full shareable room URL
func getRoomURL(roomID string) string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return "?room=" + roomID
	}

	location := js.Global().Get("window").Get("location")
	origin := location.Get("origin").String()
	pathname := location.Get("pathname").String()
	return fmt.Sprintf("%s%s?room=%s", origin, pathname, roomID)
}

// copyToClipboard copies text to clipboard using browser API
func copyToClipboard(text string) bool {
	nav := js.Global().Get("navigator")
	if nav.IsUndefined() {
		return false
	}
	clipboard := nav.Get("clipboard")
	if clipboard.IsUndefined() {
		return false
	}
	clipboard.Call("writeText", text)
	return true
}

// TokenState represents the state of a token
type TokenState int

const (
	tokenAtStart  TokenState = iota // Token is at starting point (safe house)
	tokenOnBoard                    // Token is moving on the board
	tokenFinished                   // Token has reached the center
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
	idx      int
	tokens   [tokensPerPlayer]*Token
	startPos int // Starting position on the path for this player
	entryPos int // Position where token enters inner path
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
	stateSelectingPlayers GameState = iota + 5 // Continue from multiplayer states
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

	// Multiplayer fields
	isMultiplayer    bool
	networkClient    *NetworkClient
	currentRoom      *NetworkRoom
	localPlayer      *NetworkPlayer
	roomPlayers      []*NetworkPlayer
	playerName       string
	inputText        string
	inputActive      bool
	joinRoomID       string
	errorMsg         string
	showError        bool
	gameModeSelected bool // true = multiplayer, false = local

	// Animation fields
	animRoll     int              // Roll animation counter (0 = idle)
	animMove     *animMoveState   // Token movement animation
	animKill     *animKillState   // Kill animation
	animExtra    int              // Extra turn flash counter
	lastTokenPos [4][4][2]float64 // Last drawn token screen positions [player][token][x,y]
}

type animMoveState struct {
	playerIdx int
	tokenIdx  int
	fromX     float64
	fromY     float64
	toX       float64
	toY       float64
	progress  int // 0-30 frames
}

type animKillState struct {
	playerIdx int
	tokenIdx  int
	homeX     float64
	homeY     float64
	progress  int // 0-30 frames
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
	g := &Game{
		state:       stateMenu,
		board:       NewBoard(),
		conchShells: NewConchShells(),
		winner:      -1,
	}

	// Check if there's a room ID in the URL for auto-join
	if roomID := getRoomFromURL(); roomID != "" {
		g.joinRoomID = roomID
	}

	return g
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
			if step+roll <= 8 { // 8 means center
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

	// Tick animations
	g.tickAnimations()

	// Handle menu states
	switch g.state {
	case stateMenu:
		return g.updateMenu()
	case stateEnterName:
		return g.updateEnterName()
	case stateCreateOrJoin:
		return g.updateCreateOrJoin()
	case stateWaitingRoom:
		return g.updateWaitingRoom()
	case stateConnecting:
		return g.updateConnecting()
	case stateSelectingPlayers:
		return g.updateSelectingPlayers()
	case stateGameOver:
		return g.updateGameOver()
	case statePlaying:
		return g.updatePlaying()
	}

	return nil
}

// tickAnimations updates all active animations
func (g *Game) tickAnimations() {
	// Roll animation: spins for 30 frames then stops
	if g.animRoll > 0 {
		g.animRoll--
	}

	// Move animation: progresses from 0 to 30
	if g.animMove != nil {
		g.animMove.progress++
		if g.animMove.progress >= 30 {
			g.animMove = nil
		}
	}

	// Kill animation: progresses from 0 to 30
	if g.animKill != nil {
		g.animKill.progress++
		if g.animKill.progress >= 30 {
			g.animKill = nil
		}
	}

	// Extra turn flash: counts down from 60
	if g.animExtra > 0 {
		g.animExtra--
	}
}

// startRollAnim starts the pre-roll shell animation
func (g *Game) startRollAnim() {
	g.animRoll = 30 // ~0.5 seconds at 60fps
}

// startMoveAnim starts token movement animation
func (g *Game) startMoveAnim(playerIdx, tokenIdx int, fromX, fromY, toX, toY float64) {
	g.animMove = &animMoveState{
		playerIdx: playerIdx,
		tokenIdx:  tokenIdx,
		fromX:     fromX,
		fromY:     fromY,
		toX:       toX,
		toY:       toY,
		progress:  0,
	}
}

// startKillAnim starts kill animation (token returning home)
func (g *Game) startKillAnim(playerIdx, tokenIdx int, homeX, homeY float64) {
	g.animKill = &animKillState{
		playerIdx: playerIdx,
		tokenIdx:  tokenIdx,
		homeX:     homeX,
		homeY:     homeY,
		progress:  0,
	}
}

// startExtraAnim starts extra turn flash
func (g *Game) startExtraAnim() {
	g.animExtra = 60 // ~1 second flash
}

// updateMenu handles the main menu
func (g *Game) updateMenu() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Local Game button
		if g.isButtonClicked(mx, my, screenWidth/2-100, 250, 200, 40) {
			g.gameModeSelected = false
			g.isMultiplayer = false
			g.state = stateSelectingPlayers
			g.clickCooldown = 20
			return nil
		}

		// Multiplayer button
		if g.isButtonClicked(mx, my, screenWidth/2-100, 320, 200, 40) {
			g.gameModeSelected = true
			g.isMultiplayer = true
			g.state = stateEnterName
			g.clickCooldown = 20
			return nil
		}
	}

	return nil
}

// updateEnterName handles the name entry screen
func (g *Game) updateEnterName() error {
	// Handle text input
	g.handleTextInput()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Continue button
		if g.isButtonClicked(mx, my, screenWidth/2-100, 400, 200, 40) {
			if len(g.inputText) > 0 {
				g.playerName = g.inputText
				g.state = stateCreateOrJoin
				g.clickCooldown = 20
			}
			return nil
		}
	}

	return nil
}

// updateCreateOrJoin handles the create/join room screen
func (g *Game) updateCreateOrJoin() error {
	// Handle text input for room ID
	g.handleTextInput()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Create Room button
		if g.isButtonClicked(mx, my, screenWidth/2-100, 220, 200, 40) {
			g.initMultiplayer()
			return nil
		}

		// Join Room button
		if g.isButtonClicked(mx, my, screenWidth/2-100, 450, 200, 40) {
			if len(g.joinRoomID) > 0 {
				g.initMultiplayer()
			}
			return nil
		}

		// Room ID input area
		if g.isButtonClicked(mx, my, screenWidth/2-100, 380, 200, 40) {
			g.inputActive = true
			g.joinRoomID = ""
		}
	}

	return nil
}

// updateWaitingRoom handles the waiting room screen
func (g *Game) updateWaitingRoom() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Copy Link button
		if g.currentRoom != nil && g.isButtonClicked(mx, my, 50, 170, 120, 30) {
			if roomURL := getRoomURL(g.currentRoom.ID); roomURL != "" {
				copyToClipboard(roomURL)
			}
			g.clickCooldown = 20
			return nil
		}

		// Start Game button (host only)
		if g.localPlayer != nil && g.localPlayer.IsHost {
			if g.isButtonClicked(mx, my, screenWidth/2-100, 500, 200, 40) {
				if g.networkClient != nil {
					g.networkClient.StartGame()
				}
				g.clickCooldown = 20
				return nil
			}
		}

		// Back button
		if g.isButtonClicked(mx, my, 50, 600, 100, 40) {
			if g.networkClient != nil {
				g.networkClient.Disconnect()
			}
			g.state = stateMenu
			g.clickCooldown = 20
			return nil
		}
	}

	return nil
}

// updateConnecting handles the connecting screen
func (g *Game) updateConnecting() error {
	// Just wait for connection
	return nil
}

// updateSelectingPlayers handles local player selection
func (g *Game) updateSelectingPlayers() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Back button
		if g.isButtonClicked(mx, my, 50, 600, 100, 40) {
			g.state = stateMenu
			g.clickCooldown = 20
			return nil
		}

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

// updateGameOver handles game over state
func (g *Game) updateGameOver() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		if g.isMultiplayer {
			g.state = stateWaitingRoom
		} else {
			g.state = stateSelectingPlayers
		}
		g.clickCooldown = 20
	}
	return nil
}

// updatePlaying handles the playing state
func (g *Game) updatePlaying() error {
	// Handle mouse click during playing
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// In multiplayer, only allow actions on your turn
		if g.isMultiplayer && g.localPlayer != nil {
			if g.currentPlayer != g.localPlayer.PlayerIdx {
				return nil // Not your turn
			}
		}

		// Check roll button click
		shellAreaX := 600
		buttonY := 450
		buttonWidth := 140
		buttonHeight := 40

		if mx >= shellAreaX+20 && mx <= shellAreaX+20+buttonWidth &&
			my >= buttonY && my <= buttonY+buttonHeight {
			if !g.hasRolled {
				// Start pre-roll animation
				g.startRollAnim()

				// Multiplayer: request roll from server (result comes via gameStateUpdate)
				if g.isMultiplayer && g.networkClient != nil {
					g.clickCooldown = 10
					g.networkClient.SendRollAction()
					return nil
				}

				// Local game roll (result shown after anim completes)
				g.rollResult = g.conchShells.Roll()
				g.hasRolled = true
				g.clickCooldown = 10

				if g.rollResult == 4 || g.rollResult == 8 {
					g.extraTurn = true
					g.startExtraAnim()
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

					// Send action to server in multiplayer
					if g.isMultiplayer && g.networkClient != nil {
						g.networkClient.SendGameAction("move", idx)
						return nil
					}

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

// isButtonClicked checks if a button was clicked
func (g *Game) isButtonClicked(mx, my, bx, by, bw, bh int) bool {
	return mx >= bx && mx <= bx+bw && my >= by && my <= by+bh
}

// handleTextInput handles keyboard text input
func (g *Game) handleTextInput() {
	// Handle backspace using inpututil for single press detection
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if g.state == stateEnterName && len(g.inputText) > 0 {
			g.inputText = g.inputText[:len(g.inputText)-1]
		} else if g.state == stateCreateOrJoin && g.inputActive && len(g.joinRoomID) > 0 {
			g.joinRoomID = g.joinRoomID[:len(g.joinRoomID)-1]
		}
		return
	}

	// Handle character input
	chars := ebiten.AppendInputChars(nil)
	for _, ch := range chars {
		if ch >= 32 && ch <= 126 { // Printable ASCII
			if g.state == stateEnterName && len(g.inputText) < 15 {
				g.inputText += string(ch)
			} else if g.state == stateCreateOrJoin && g.inputActive && len(g.joinRoomID) < 10 {
				g.joinRoomID += string(ch)
			}
		}
	}
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{60, 40, 20, 255})

	switch g.state {
	case stateMenu:
		g.drawMenu(screen)
	case stateEnterName:
		g.drawEnterName(screen)
	case stateCreateOrJoin:
		g.drawCreateOrJoin(screen)
	case stateWaitingRoom:
		g.drawWaitingRoom(screen)
	case stateConnecting:
		g.drawConnecting(screen)
	case stateSelectingPlayers:
		g.drawSelectionScreen(screen)
	case statePlaying, stateGameOver:
		g.drawGame(screen)
	}

	// Draw error message if any
	if g.showError {
		g.drawError(screen)
	}
}

// drawMenu draws the main menu
func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ASHTA CHANGA", screenWidth/2-50, 100)
	ebitenutil.DebugPrintAt(screen, "Traditional Indian Board Game", screenWidth/2-100, 130)

	// Local Game button
	g.drawButton(screen, screenWidth/2-100, 250, 200, 40, "Local Game", color.RGBA{139, 90, 43, 255})

	// Multiplayer button
	g.drawButton(screen, screenWidth/2-100, 320, 200, 40, "Multiplayer", color.RGBA{139, 90, 43, 255})

	// Instructions
	ebitenutil.DebugPrintAt(screen, "2-4 Players | Roll shells to move tokens", screenWidth/2-120, 450)
	ebitenutil.DebugPrintAt(screen, "Reach the center with all 4 tokens to win!", screenWidth/2-130, 470)
}

// drawEnterName draws the name entry screen
func (g *Game) drawEnterName(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ENTER YOUR NAME", screenWidth/2-60, 150)

	// Input box
	vector.DrawFilledRect(screen, float32(screenWidth/2-150), 250, 300, 50, color.RGBA{200, 200, 200, 255}, false)
	vector.StrokeRect(screen, float32(screenWidth/2-150), 250, 300, 50, 2, color.RGBA{255, 255, 255, 255}, false)

	// Show input text
	if len(g.inputText) > 0 {
		ebitenutil.DebugPrintAt(screen, g.inputText, screenWidth/2-140, 265)
	} else {
		ebitenutil.DebugPrintAt(screen, "Type your name...", screenWidth/2-140, 265)
	}

	// Continue button
	btnColor := color.RGBA{100, 100, 100, 255}
	if len(g.inputText) > 0 {
		btnColor = color.RGBA{139, 90, 43, 255}
	}
	g.drawButton(screen, screenWidth/2-100, 400, 200, 40, "Continue", btnColor)
}

// drawCreateOrJoin draws the create/join room screen
func (g *Game) drawCreateOrJoin(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "MULTIPLAYER", screenWidth/2-40, 80)

	// Create Room section
	ebitenutil.DebugPrintAt(screen, "Create a new room:", screenWidth/2-70, 180)
	g.drawButton(screen, screenWidth/2-100, 220, 200, 40, "Create Room", color.RGBA{139, 90, 43, 255})

	// Divider
	vector.StrokeLine(screen, float32(screenWidth/2-150), 300, float32(screenWidth/2+150), 300, 2, color.RGBA{150, 150, 150, 255}, false)
	ebitenutil.DebugPrintAt(screen, "OR", screenWidth/2-10, 310)

	// Join Room section
	ebitenutil.DebugPrintAt(screen, "Join with Room ID:", screenWidth/2-70, 340)

	// Room ID input box
	inputColor := color.RGBA{200, 200, 200, 255}
	if g.inputActive {
		inputColor = color.RGBA{230, 230, 230, 255}
	}
	vector.DrawFilledRect(screen, float32(screenWidth/2-100), 380, 200, 40, inputColor, false)
	vector.StrokeRect(screen, float32(screenWidth/2-100), 380, 200, 40, 2, color.RGBA{255, 255, 255, 255}, false)

	// Show room ID
	displayText := g.joinRoomID
	if len(displayText) == 0 {
		displayText = "Enter Room ID"
	}
	ebitenutil.DebugPrintAt(screen, displayText, screenWidth/2-90, 395)

	// Join button
	btnColor := color.RGBA{100, 100, 100, 255}
	if len(g.joinRoomID) > 0 {
		btnColor = color.RGBA{139, 90, 43, 255}
	}
	g.drawButton(screen, screenWidth/2-100, 450, 200, 40, "Join Room", btnColor)
}

// drawWaitingRoom draws the waiting room screen
func (g *Game) drawWaitingRoom(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "WAITING ROOM", screenWidth/2-50, 50)

	if g.currentRoom != nil {
		// Room info
		ebitenutil.DebugPrintAt(screen, "Room: "+g.currentRoom.Name, 50, 100)
		ebitenutil.DebugPrintAt(screen, "Room ID: "+g.currentRoom.ID, 50, 120)

		// Share link with copy button
		shareURL := "Link: ?room=" + g.currentRoom.ID
		ebitenutil.DebugPrintAt(screen, shareURL, 50, 150)
		g.drawButton(screen, 50, 170, 120, 30, "Copy Link", color.RGBA{70, 130, 180, 255})

		// Players list
		ebitenutil.DebugPrintAt(screen, "Players:", 50, 220)
		for i, player := range g.roomPlayers {
			y := 250 + i*30
			status := ""
			if player.IsHost {
				status = " (Host)"
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d. %s%s", i+1, player.Name, status), 70, y)
		}

		// Start Game button (host only)
		if g.localPlayer != nil && g.localPlayer.IsHost {
			g.drawButton(screen, screenWidth/2-100, 500, 200, 40, "Start Game", color.RGBA{0, 150, 0, 255})
			ebitenutil.DebugPrintAt(screen, "Waiting for players...", screenWidth/2-80, 560)
		} else {
			ebitenutil.DebugPrintAt(screen, "Waiting for host to start...", screenWidth/2-100, 500)
		}
	}

	// Back button
	g.drawButton(screen, 50, 600, 100, 40, "Back", color.RGBA{139, 90, 43, 255})
}

// initMultiplayer initializes multiplayer connection
func (g *Game) initMultiplayer() {
	g.state = stateConnecting

	go func() {
		var roomID string
		var playerID string

		// If creating a room, use HTTP API
		if g.joinRoomID == "" {
			apiURL := getAPIURL()
			reqBody, _ := json.Marshal(map[string]string{
				"roomName":   g.playerName + "'s Room",
				"playerName": g.playerName,
			})

			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				g.handleError("HTTP Error: " + err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				g.handleError(fmt.Sprintf("Server Error (%d): %s", resp.StatusCode, string(body)))
				return
			}

			var result struct {
				Room     *NetworkRoom `json:"room"`
				PlayerID string       `json:"playerId"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				g.handleError("Decode Error: " + err.Error())
				return
			}

			roomID = result.Room.ID
			playerID = result.PlayerID
		} else {
			roomID = g.joinRoomID
		}

		// Connect via WebSocket
		serverURL := getServerURL()
		g.networkClient = NewNetworkClient(serverURL)

		// Set up callbacks
		g.networkClient.onRoomCreated = func(r *NetworkRoom, p *NetworkPlayer) {
			g.currentRoom = r
			g.localPlayer = p
			g.roomPlayers = r.Players
			g.state = stateWaitingRoom
		}

		g.networkClient.onRoomJoined = func(r *NetworkRoom, p *NetworkPlayer) {
			g.currentRoom = r
			g.localPlayer = p
			g.roomPlayers = r.Players
			g.state = stateWaitingRoom
		}

		g.networkClient.onPlayerJoined = func(p *NetworkPlayer) {
			g.roomPlayers = append(g.roomPlayers, p)
		}

		g.networkClient.onPlayerLeft = func(p *NetworkPlayer) {
			for i, player := range g.roomPlayers {
				if player.ID == p.ID {
					g.roomPlayers = append(g.roomPlayers[:i], g.roomPlayers[i+1:]...)
					break
				}
			}
		}

		g.networkClient.onGameStarted = func(gameState *ServerGameState) {
			g.startMultiplayerGame(gameState)
		}

		g.networkClient.onGameStateUpdate = func(gameState *ServerGameState) {
			g.updateGameStateFromServer(gameState)
		}

		g.networkClient.onError = func(msg string) {
			g.handleError(msg)
		}

		if err := g.networkClient.Connect(); err != nil {
			g.handleError("Connection Error: " + err.Error())
			return
		}

		// Join the room (either the one we just created or the one we want to join)
		g.networkClient.JoinRoom(roomID, g.playerName, playerID)
	}()
}

// handleError handles errors in multiplayer setup
func (g *Game) handleError(msg string) {
	g.errorMsg = msg
	g.showError = true
	g.state = stateMenu
}

// startMultiplayerGame starts the game with multiplayer state
func (g *Game) startMultiplayerGame(serverState *ServerGameState) {
	log.Printf("Starting multiplayer game with %d players", serverState.NumActivePlayers)

	g.numActivePlayers = serverState.NumActivePlayers
	g.isMultiplayer = true

	// Initialize players from server state
	for i := 0; i < g.numActivePlayers; i++ {
		g.players[i] = NewPlayer(i)
		// Update player name from room players if available
		for _, roomPlayer := range g.roomPlayers {
			if roomPlayer.PlayerIdx == i {
				playerNames[i] = roomPlayer.Name
				break
			}
		}
	}

	// Initialize token positions from server state
	for _, playerTokens := range serverState.PlayerTokens {
		playerIdx := playerTokens.PlayerIdx
		if playerIdx < 0 || playerIdx >= g.numActivePlayers || g.players[playerIdx] == nil {
			continue
		}
		for _, tokenState := range playerTokens.Tokens {
			if tokenState.ID < 0 || tokenState.ID >= tokensPerPlayer {
				continue
			}
			token := g.players[playerIdx].tokens[tokenState.ID]
			token.state = TokenState(tokenState.State)
			token.position = tokenState.Position
		}
	}

	// Set game state from server
	g.currentPlayer = serverState.CurrentPlayer
	g.rollResult = serverState.RollResult
	g.hasRolled = serverState.HasRolled
	g.extraTurn = serverState.ExtraTurn
	g.winner = serverState.Winner

	// Sync shell states for multiplayer
	if len(serverState.ShellStates) == len(g.conchShells.shells) {
		for i := range g.conchShells.shells {
			g.conchShells.shells[i].openSideUp = serverState.ShellStates[i]
		}
	}

	g.state = statePlaying

	log.Printf("Multiplayer game started. Current player: %d, State: %d", g.currentPlayer, g.state)
}

// updateGameStateFromServer updates the game state from server during gameplay
func (g *Game) updateGameStateFromServer(serverState *ServerGameState) {
	g.currentPlayer = serverState.CurrentPlayer
	g.rollResult = serverState.RollResult
	g.hasRolled = serverState.HasRolled
	g.extraTurn = serverState.ExtraTurn
	g.winner = serverState.Winner

	// Sync shell states for multiplayer
	if len(serverState.ShellStates) == len(g.conchShells.shells) {
		for i := range g.conchShells.shells {
			g.conchShells.shells[i].openSideUp = serverState.ShellStates[i]
		}
	}

	// Update token positions from server
	for _, playerTokens := range serverState.PlayerTokens {
		playerIdx := playerTokens.PlayerIdx
		if playerIdx < 0 || playerIdx >= g.numActivePlayers || g.players[playerIdx] == nil {
			continue
		}
		for _, tokenState := range playerTokens.Tokens {
			if tokenState.ID < 0 || tokenState.ID >= tokensPerPlayer {
				continue
			}
			token := g.players[playerIdx].tokens[tokenState.ID]
			token.state = TokenState(tokenState.State)
			token.position = tokenState.Position
		}
	}

	// Check for game over
	if serverState.Winner >= 0 {
		g.state = stateGameOver
	}
}

// drawConnecting draws the connecting screen
func (g *Game) drawConnecting(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "CONNECTING...", screenWidth/2-50, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "Please wait", screenWidth/2-40, screenHeight/2+30)
}

// drawGame draws the game board
func (g *Game) drawGame(screen *ebiten.Image) {
	g.board.Draw(screen)
	g.drawTokens(screen)
	g.drawConchShellArea(screen)
	g.drawPlayerInfo(screen)

	title := "Ashta Changa - Traditional Indian Board Game"
	if g.isMultiplayer {
		title = "Ashta Changa - Multiplayer"
	}
	ebitenutil.DebugPrintAt(screen, title, 50, 10)

	if g.state == stateGameOver {
		vector.DrawFilledRect(screen, 200, 250, 400, 100, color.RGBA{0, 0, 0, 200}, false)
		ebitenutil.DebugPrintAt(screen, "GAME OVER!", 350, 280)
		ebitenutil.DebugPrintAt(screen, playerNames[g.winner]+" Wins!", 350, 300)
		ebitenutil.DebugPrintAt(screen, "Click anywhere to restart", 320, 330)
	}
}

// drawButton draws a button
func (g *Game) drawButton(screen *ebiten.Image, x, y, w, h int, label string, bgColor color.RGBA) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bgColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, color.RGBA{210, 180, 140, 255}, false)

	// Center text
	textX := x + (w-len(label)*8)/2
	textY := y + (h-16)/2 + 5
	ebitenutil.DebugPrintAt(screen, label, textX, textY)
}

// drawError draws an error message
func (g *Game) drawError(screen *ebiten.Image) {
	// Error background
	vector.DrawFilledRect(screen, 150, 300, 500, 100, color.RGBA{150, 0, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "ERROR", 370, 320)
	ebitenutil.DebugPrintAt(screen, g.errorMsg, 200, 350)
	ebitenutil.DebugPrintAt(screen, "Click to dismiss", 330, 380)
}

// drawSelectionScreen draws the player selection UI
func (g *Game) drawSelectionScreen(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ASHTA CHANGA", screenWidth/2-50, 100)
	ebitenutil.DebugPrintAt(screen, "Select Number of Players", screenWidth/2-80, 150)

	for i := 1; i <= 4; i++ {
		bx := screenWidth/2 - 100
		by := 200 + i*60
		g.drawButton(screen, bx, by, 200, 40, strconv.Itoa(i)+" Players", color.RGBA{139, 90, 43, 255})
	}

	// Back button
	g.drawButton(screen, 50, 600, 100, 40, "Back", color.RGBA{139, 90, 43, 255})
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
			if len(tokens) > 4 {
				numInRow = 3
			}
			if len(tokens) > 9 {
				numInRow = 4
			}

			offsetX := (i%numInRow)*(cellSize/(numInRow+1)) + (cellSize / (numInRow + 1))
			offsetY := (i/numInRow)*(cellSize/(numInRow+1)) + (cellSize / (numInRow + 1))

			tx := cellX + offsetX
			ty := cellY + offsetY

			tokenColor := playerColors[token.playerIdx]
			isValid := false
			if g.hasRolled && token.playerIdx == g.currentPlayer {
				if !g.isMultiplayer || (g.localPlayer != nil && g.currentPlayer == g.localPlayer.PlayerIdx) {
					for _, idx := range g.getValidTokens() {
						if g.players[g.currentPlayer].tokens[idx] == token {
							isValid = true
							break
						}
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

	// Check if it's local player's turn in multiplayer
	isMyTurn := !g.isMultiplayer || (g.localPlayer != nil && g.currentPlayer == g.localPlayer.PlayerIdx)

	if isMyTurn {
		ebitenutil.DebugPrintAt(screen, "YOUR TURN!", infoX, infoY)
	} else {
		ebitenutil.DebugPrintAt(screen, "Waiting for:", infoX, infoY)
	}

	playerColor := playerColors[g.currentPlayer]
	vector.DrawFilledRect(screen, float32(infoX+110), float32(infoY), 60, 20, playerColor, false)
	ebitenutil.DebugPrintAt(screen, playerNames[g.currentPlayer], infoX+115, infoY+2)

	if g.extraTurn {
		// Pulsing animation for extra turn
		alpha := uint8(255)
		if g.animExtra > 0 && g.animExtra%10 < 5 {
			alpha = 100
		}
		extraColor := color.RGBA{255, 255, 0, alpha}
		vector.DrawFilledRect(screen, float32(infoX), float32(infoY+25), 100, 16, extraColor, false)
		ebitenutil.DebugPrintAt(screen, "EXTRA TURN!", infoX, infoY+25)
	}

	if g.hasRolled {
		validTokens := g.getValidTokens()
		if len(validTokens) > 0 {
			if isMyTurn {
				ebitenutil.DebugPrintAt(screen, "Click a cell with your", infoX, infoY+50)
				ebitenutil.DebugPrintAt(screen, "highlighted token", infoX, infoY+65)
			}
		}
	} else {
		if isMyTurn {
			ebitenutil.DebugPrintAt(screen, "Click ROLL to roll shells", infoX, infoY+50)
		} else {
			ebitenutil.DebugPrintAt(screen, playerNames[g.currentPlayer]+" is rolling...", infoX, infoY+50)
		}
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
		// Pre-roll animation: shake shells
		if g.animRoll > 0 {
			shake := float64(g.animRoll%6 - 3) // -3 to +3
			shellX += int(shake)
			shellY += int(shake * 0.5)
		}
		g.drawConchShell(screen, shellX, shellY, g.conchShells.shells[i].openSideUp)
	}

	resultY := shellStartY + 4*shellSpacing + 20
	ebitenutil.DebugPrintAt(screen, "Result:", shellAreaX+60, resultY)

	// Show result only when not animating and have result
	showResult := g.rollResult > 0 && g.animRoll == 0
	if showResult {
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
	} else if g.animRoll > 0 {
		ebitenutil.DebugPrintAt(screen, "Rolling...", shellAreaX+55, resultY+20)
	}

	buttonY := resultY + 80
	buttonWidth := 140
	buttonHeight := 40
	buttonColor := color.RGBA{150, 150, 150, 255} // Grey default

	// Check if it's my turn in multiplayer
	isMyTurn := !g.isMultiplayer || (g.localPlayer != nil && g.currentPlayer == g.localPlayer.PlayerIdx)

	if !g.hasRolled && isMyTurn {
		buttonColor = color.RGBA{255, 215, 0, 255} // Gold if can roll
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
