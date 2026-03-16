// Package gui contains the Ebiten-based GUI for Ashta Changa.
// This package handles rendering, input, and game state management.
package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/anujdecoder/ashta-board/game"
	"github.com/anujdecoder/ashta-board/network"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 700
	BoardOffsetX = 50
	BoardOffsetY = 50
)

// GameState represents the current state of the game UI
type GameState int

const (
	StateMenu GameState = iota
	StateEnterName
	StateCreateOrJoin
	StateWaitingRoom
	StateConnecting
	StateSelectingPlayers
	StatePlaying
	StateGameOver
)

// AnimationManager is now defined in animation.go

// Game represents the game state and implements ebiten.Game interface
type Game struct {
	state            GameState
	board            *game.Board
	conchShells      *game.ConchShells
	players          [4]*game.Player
	numActivePlayers int
	currentPlayer    int
	rollResult       int
	hasRolled        bool
	extraTurn        bool
	winner           int
	clickCooldown    int // Prevent multiple clicks

	// Multiplayer fields
	isMultiplayer    bool
	networkClient    *network.NetworkClient
	currentRoom      *network.NetworkRoom
	localPlayer      *network.NetworkPlayer
	roomPlayers      []*network.NetworkPlayer
	playerName       string
	inputText        string
	inputActive      bool
	joinRoomID       string
	errorMsg         string
	showError        bool
	gameModeSelected bool // true = multiplayer, false = local

	// Animation manager
	animManager *AnimationManager
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		state:       StateMenu,
		board:       game.NewBoard(),
		conchShells: game.NewConchShells(),
		winner:      -1,
		animManager: NewAnimationManager(),
	}

	// Check if there's a room ID in the URL for auto-join
	if roomID := network.GetRoomFromURL(); roomID != "" {
		g.joinRoomID = roomID
	}

	return g
}

// InitGame initializes the game with selected number of players
func (g *Game) InitGame(num int) {
	g.numActivePlayers = num
	for i := 0; i < 4; i++ {
		if i < num {
			g.players[i] = game.NewPlayer(i)
		} else {
			g.players[i] = nil
		}
	}
	g.state = StatePlaying
	g.currentPlayer = 0
}

// InitMultiplayer initializes multiplayer connection
func (g *Game) InitMultiplayer() {
	g.state = StateConnecting

	go func() {
		var roomID string
		var playerID string

		// If creating a room, use HTTP API
		if g.joinRoomID == "" {
			apiURL := network.GetAPIURL()
			reqBody, _ := json.Marshal(map[string]string{
				"roomName":   g.playerName + "'s Room",
				"playerName": g.playerName,
			})

			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				g.HandleError("HTTP Error: " + err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				g.HandleError(fmt.Sprintf("Server Error (%d): %s", resp.StatusCode, string(body)))
				return
			}

			var result struct {
				Room     *network.NetworkRoom `json:"room"`
				PlayerID string               `json:"playerId"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				g.HandleError("Decode Error: " + err.Error())
				return
			}

			roomID = result.Room.ID
			playerID = result.PlayerID
		} else {
			roomID = g.joinRoomID
		}

		// Connect via WebSocket
		serverURL := network.GetServerURL()
		g.networkClient = network.NewNetworkClient(serverURL)

		// Set up callbacks
		g.networkClient.OnRoomCreated = func(r *network.NetworkRoom, p *network.NetworkPlayer) {
			g.currentRoom = r
			g.localPlayer = p
			g.roomPlayers = r.Players
			g.state = StateWaitingRoom
		}

		g.networkClient.OnRoomJoined = func(r *network.NetworkRoom, p *network.NetworkPlayer) {
			g.currentRoom = r
			g.localPlayer = p
			g.roomPlayers = r.Players
			g.state = StateWaitingRoom
		}

		g.networkClient.OnPlayerJoined = func(p *network.NetworkPlayer) {
			g.roomPlayers = append(g.roomPlayers, p)
		}

		g.networkClient.OnPlayerLeft = func(p *network.NetworkPlayer) {
			for i, player := range g.roomPlayers {
				if player.ID == p.ID {
					g.roomPlayers = append(g.roomPlayers[:i], g.roomPlayers[i+1:]...)
					break
				}
			}
		}

		g.networkClient.OnGameStarted = func(gameState *network.ServerGameState) {
			g.StartMultiplayerGame(gameState)
		}

		g.networkClient.OnGameStateUpdate = func(gameState *network.ServerGameState) {
			g.UpdateGameStateFromServer(gameState)
		}

		g.networkClient.OnError = func(msg string) {
			g.HandleError(msg)
		}

		if err := g.networkClient.Connect(); err != nil {
			g.HandleError("Connection Error: " + err.Error())
			return
		}

		// Join the room (either the one we just created or the one we want to join)
		g.networkClient.JoinRoom(roomID, g.playerName, playerID)
	}()
}

// HandleError handles errors in multiplayer setup
func (g *Game) HandleError(msg string) {
	g.errorMsg = msg
	g.showError = true
	g.state = StateMenu
}

// StartMultiplayerGame starts the game with multiplayer state
func (g *Game) StartMultiplayerGame(serverState *network.ServerGameState) {
	log.Printf("Starting multiplayer game with %d players", serverState.NumActivePlayers)

	g.numActivePlayers = serverState.NumActivePlayers
	g.isMultiplayer = true

	// Initialize players from server state
	for i := 0; i < g.numActivePlayers; i++ {
		g.players[i] = game.NewPlayer(i)
		// Update player name from room players if available
		for _, roomPlayer := range g.roomPlayers {
			if roomPlayer.PlayerIdx == i {
				game.PlayerNames[i] = roomPlayer.Name
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
			if tokenState.ID < 0 || tokenState.ID >= game.TokensPerPlayer {
				continue
			}
			token := g.players[playerIdx].Tokens[tokenState.ID]
			token.State = game.TokenState(tokenState.State)
			token.Position = tokenState.Position
		}
	}

	// Set game state from server
	g.currentPlayer = serverState.CurrentPlayer
	g.rollResult = serverState.RollResult
	g.hasRolled = serverState.HasRolled
	g.extraTurn = serverState.ExtraTurn
	g.winner = serverState.Winner

	// Sync shell states for multiplayer
	if len(serverState.ShellStates) == len(g.conchShells.Shells) {
		for i := range g.conchShells.Shells {
			g.conchShells.Shells[i].OpenSideUp = serverState.ShellStates[i]
		}
	}

	g.state = StatePlaying

	log.Printf("Multiplayer game started. Current player: %d, State: %d", g.currentPlayer, g.state)
}

// UpdateGameStateFromServer updates the game state from server during gameplay
func (g *Game) UpdateGameStateFromServer(serverState *network.ServerGameState) {
	g.currentPlayer = serverState.CurrentPlayer
	g.rollResult = serverState.RollResult
	g.hasRolled = serverState.HasRolled
	g.extraTurn = serverState.ExtraTurn
	g.winner = serverState.Winner

	// Sync shell states for multiplayer
	if len(serverState.ShellStates) == len(g.conchShells.Shells) {
		for i := range g.conchShells.Shells {
			g.conchShells.Shells[i].OpenSideUp = serverState.ShellStates[i]
		}
	}

	// Update token positions from server
	for _, playerTokens := range serverState.PlayerTokens {
		playerIdx := playerTokens.PlayerIdx
		if playerIdx < 0 || playerIdx >= g.numActivePlayers || g.players[playerIdx] == nil {
			continue
		}
		for _, tokenState := range playerTokens.Tokens {
			if tokenState.ID < 0 || tokenState.ID >= game.TokensPerPlayer {
				continue
			}
			token := g.players[playerIdx].Tokens[tokenState.ID]
			token.State = game.TokenState(tokenState.State)
			token.Position = tokenState.Position
		}
	}

	// Check for game over
	if serverState.Winner >= 0 {
		g.state = StateGameOver
	}
}

// IsButtonClicked checks if a button was clicked
func (g *Game) IsButtonClicked(mx, my, bx, by, bw, bh int) bool {
	return mx >= bx && mx <= bx+bw && my >= by && my <= by+bh
}

// HandleTextInput handles keyboard text input
func (g *Game) HandleTextInput() {
	// Handle backspace using inpututil for single press detection
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if g.state == StateEnterName && len(g.inputText) > 0 {
			g.inputText = g.inputText[:len(g.inputText)-1]
		} else if g.state == StateCreateOrJoin && g.inputActive && len(g.joinRoomID) > 0 {
			g.joinRoomID = g.joinRoomID[:len(g.joinRoomID)-1]
		}
		return
	}

	// Handle character input
	chars := ebiten.AppendInputChars(nil)
	for _, ch := range chars {
		if ch >= 32 && ch <= 126 { // Printable ASCII
			if g.state == StateEnterName && len(g.inputText) < 15 {
				g.inputText += string(ch)
			} else if g.state == StateCreateOrJoin && g.inputActive && len(g.joinRoomID) < 10 {
				g.joinRoomID += string(ch)
			}
		}
	}
}

// Layout returns the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
