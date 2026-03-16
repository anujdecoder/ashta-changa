// Package gui contains the Ebiten-based GUI for Ashta Changa.
package gui

import (
	"image/color"

	"github.com/anujdecoder/ashta-board/game"
	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer defines the interface for rendering game elements.
// This interface allows for mocking in tests and alternative rendering implementations.
type Renderer interface {
	// Draw renders the complete game frame
	Draw(screen *ebiten.Image)

	// DrawBoard draws the game board grid
	DrawBoard(screen *ebiten.Image, board *game.Board)

	// DrawTokens draws all tokens on the board
	DrawTokens(screen *ebiten.Image, players [4]*game.Player, currentPlayer int, hasRolled bool, rollResult int)

	// DrawConchShellArea draws the conch shell rolling area
	DrawConchShellArea(screen *ebiten.Image, conchShells *game.ConchShells, hasRolled bool, currentPlayer int, animRoll int)

	// DrawPlayerInfo draws current player and turn information
	DrawPlayerInfo(screen *ebiten.Image, currentPlayer int, players [4]*game.Player, extraTurn bool, hasRolled bool, rollResult int, winner int)

	// DrawButton draws a button with the given parameters
	DrawButton(screen *ebiten.Image, x, y, w, h int, label string, bgColor color.RGBA)

	// DrawMenu draws the main menu
	DrawMenu(screen *ebiten.Image)

	// DrawEnterName draws the name entry screen
	DrawEnterName(screen *ebiten.Image, inputText string)

	// DrawCreateOrJoin draws the create/join room screen
	DrawCreateOrJoin(screen *ebiten.Image, joinRoomID string, inputActive bool)

	// DrawWaitingRoom draws the waiting room screen
	DrawWaitingRoom(screen *ebiten.Image, roomPlayers []*NetworkPlayerDisplay, localPlayer *NetworkPlayerDisplay, currentRoom *NetworkRoomDisplay)

	// DrawConnecting draws the connecting screen
	DrawConnecting(screen *ebiten.Image)

	// DrawSelectionScreen draws the player selection screen
	DrawSelectionScreen(screen *ebiten.Image)

	// DrawGame draws the game board during play
	DrawGame(screen *ebiten.Image, board *game.Board, players [4]*game.Player, currentPlayer int, hasRolled bool, rollResult int, extraTurn bool, winner int, conchShells *game.ConchShells, animRoll int)

	// DrawError draws an error message
	DrawError(screen *ebiten.Image, errorMsg string)
}

// NetworkPlayerDisplay is a simplified player struct for rendering
type NetworkPlayerDisplay struct {
	ID        string
	Name      string
	PlayerIdx int
	IsHost    bool
	IsReady   bool
}

// NetworkRoomDisplay is a simplified room struct for rendering
type NetworkRoomDisplay struct {
	ID     string
	Name   string
	HostID string
}

// DefaultRenderer implements Renderer with the standard Ebiten rendering
type DefaultRenderer struct {
	game *Game
}

// Ensure DefaultRenderer implements Renderer
var _ Renderer = (*DefaultRenderer)(nil)

// NewRenderer creates a new instance of the default renderer
func NewRenderer(g *Game) Renderer {
	return &DefaultRenderer{game: g}
}

// Draw renders the complete game frame
func (r *DefaultRenderer) Draw(screen *ebiten.Image) {
	r.game.Draw(screen)
}

// DrawBoard draws the game board grid
func (r *DefaultRenderer) DrawBoard(screen *ebiten.Image, board *game.Board) {
	// Implementation would draw the board
	r.game.drawBoard(screen)
}

// DrawTokens draws all tokens on the board
func (r *DefaultRenderer) DrawTokens(screen *ebiten.Image, players [4]*game.Player, currentPlayer int, hasRolled bool, rollResult int) {
	// Store original state
	origPlayers := r.game.players
	origCurrentPlayer := r.game.currentPlayer
	origHasRolled := r.game.hasRolled
	origRollResult := r.game.rollResult

	// Set temporary state
	r.game.players = players
	r.game.currentPlayer = currentPlayer
	r.game.hasRolled = hasRolled
	r.game.rollResult = rollResult

	// Draw
	r.game.drawTokens(screen)

	// Restore original state
	r.game.players = origPlayers
	r.game.currentPlayer = origCurrentPlayer
	r.game.hasRolled = origHasRolled
	r.game.rollResult = origRollResult
}

// DrawConchShellArea draws the conch shell rolling area
func (r *DefaultRenderer) DrawConchShellArea(screen *ebiten.Image, conchShells *game.ConchShells, hasRolled bool, currentPlayer int, animRoll int) {
	origConchShells := r.game.conchShells
	origHasRolled := r.game.hasRolled
	origCurrentPlayer := r.game.currentPlayer

	r.game.conchShells = conchShells
	r.game.hasRolled = hasRolled
	r.game.currentPlayer = currentPlayer

	r.game.drawConchShellArea(screen)

	r.game.conchShells = origConchShells
	r.game.hasRolled = origHasRolled
	r.game.currentPlayer = origCurrentPlayer
}

// DrawPlayerInfo draws current player and turn information
func (r *DefaultRenderer) DrawPlayerInfo(screen *ebiten.Image, currentPlayer int, players [4]*game.Player, extraTurn bool, hasRolled bool, rollResult int, winner int) {
	origCurrentPlayer := r.game.currentPlayer
	origPlayers := r.game.players
	origExtraTurn := r.game.extraTurn
	origHasRolled := r.game.hasRolled
	origRollResult := r.game.rollResult
	origWinner := r.game.winner

	r.game.currentPlayer = currentPlayer
	r.game.players = players
	r.game.extraTurn = extraTurn
	r.game.hasRolled = hasRolled
	r.game.rollResult = rollResult
	r.game.winner = origWinner

	r.game.drawPlayerInfo(screen)

	r.game.currentPlayer = origCurrentPlayer
	r.game.players = origPlayers
	r.game.extraTurn = origExtraTurn
	r.game.hasRolled = origHasRolled
	r.game.rollResult = origRollResult
	r.game.winner = origWinner
}

// DrawButton draws a button with the given parameters
func (r *DefaultRenderer) DrawButton(screen *ebiten.Image, x, y, w, h int, label string, bgColor color.RGBA) {
	r.game.drawButton(screen, x, y, w, h, label, bgColor)
}

// DrawMenu draws the main menu
func (r *DefaultRenderer) DrawMenu(screen *ebiten.Image) {
	r.game.drawMenu(screen)
}

// DrawEnterName draws the name entry screen
func (r *DefaultRenderer) DrawEnterName(screen *ebiten.Image, inputText string) {
	origInputText := r.game.inputText
	r.game.inputText = inputText
	r.game.drawEnterName(screen)
	r.game.inputText = origInputText
}

// DrawCreateOrJoin draws the create/join room screen
func (r *DefaultRenderer) DrawCreateOrJoin(screen *ebiten.Image, joinRoomID string, inputActive bool) {
	origJoinRoomID := r.game.joinRoomID
	origInputActive := r.game.inputActive
	r.game.joinRoomID = joinRoomID
	r.game.inputActive = inputActive
	r.game.drawCreateOrJoin(screen)
	r.game.joinRoomID = origJoinRoomID
	r.game.inputActive = origInputActive
}

// DrawWaitingRoom draws the waiting room screen
func (r *DefaultRenderer) DrawWaitingRoom(screen *ebiten.Image, roomPlayers []*NetworkPlayerDisplay, localPlayer *NetworkPlayerDisplay, currentRoom *NetworkRoomDisplay) {
	r.game.drawWaitingRoom(screen)
}

// DrawConnecting draws the connecting screen
func (r *DefaultRenderer) DrawConnecting(screen *ebiten.Image) {
	r.game.drawConnecting(screen)
}

// DrawSelectionScreen draws the player selection screen
func (r *DefaultRenderer) DrawSelectionScreen(screen *ebiten.Image) {
	r.game.drawSelectionScreen(screen)
}

// DrawGame draws the game board during play
func (r *DefaultRenderer) DrawGame(screen *ebiten.Image, board *game.Board, players [4]*game.Player, currentPlayer int, hasRolled bool, rollResult int, extraTurn bool, winner int, conchShells *game.ConchShells, animRoll int) {
	// Store original state
	origBoard := r.game.board
	origPlayers := r.game.players
	origCurrentPlayer := r.game.currentPlayer
	origHasRolled := r.game.hasRolled
	origRollResult := r.game.rollResult
	origExtraTurn := r.game.extraTurn
	origWinner := r.game.winner
	origConchShells := r.game.conchShells

	// Set temporary state
	r.game.board = board
	r.game.players = players
	r.game.currentPlayer = currentPlayer
	r.game.hasRolled = hasRolled
	r.game.rollResult = rollResult
	r.game.extraTurn = extraTurn
	r.game.winner = winner
	r.game.conchShells = conchShells

	// Draw
	r.game.drawGame(screen)

	// Restore original state
	r.game.board = origBoard
	r.game.players = origPlayers
	r.game.currentPlayer = origCurrentPlayer
	r.game.hasRolled = origHasRolled
	r.game.rollResult = origRollResult
	r.game.extraTurn = origExtraTurn
	r.game.winner = origWinner
	r.game.conchShells = origConchShells
}

// DrawError draws an error message
func (r *DefaultRenderer) DrawError(screen *ebiten.Image, errorMsg string) {
	origErrorMsg := r.game.errorMsg
	origShowError := r.game.showError
	r.game.errorMsg = errorMsg
	r.game.showError = true
	r.game.drawError(screen)
	r.game.errorMsg = origErrorMsg
	r.game.showError = origShowError
}
