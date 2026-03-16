// Package gui contains the Ebiten-based GUI for Ashta Changa.
package gui

import (
	"github.com/anujdecoder/ashta-board/game"
	"github.com/anujdecoder/ashta-board/network"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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
	case StateMenu:
		return g.updateMenu()
	case StateEnterName:
		return g.updateEnterName()
	case StateCreateOrJoin:
		return g.updateCreateOrJoin()
	case StateWaitingRoom:
		return g.updateWaitingRoom()
	case StateConnecting:
		return g.updateConnecting()
	case StateSelectingPlayers:
		return g.updateSelectingPlayers()
	case StateGameOver:
		return g.updateGameOver()
	case StatePlaying:
		return g.updatePlaying()
	}

	return nil
}

// tickAnimations updates all active animations
func (g *Game) tickAnimations() {
	if g.animManager != nil {
		g.animManager.Tick()
	}
}

// startRollAnim starts the pre-roll shell animation
func (g *Game) startRollAnim() {
	if g.animManager != nil {
		g.animManager.StartRollAnim()
	}
}

// startMoveAnim starts token movement animation
func (g *Game) startMoveAnim(playerIdx, tokenIdx int, fromX, fromY, toX, toY float64) {
	if g.animManager != nil {
		g.animManager.StartMoveAnim(playerIdx, tokenIdx, fromX, fromY, toX, toY)
	}
}

// startKillAnim starts kill animation (token returning home)
func (g *Game) startKillAnim(playerIdx, tokenIdx int, homeX, homeY float64) {
	if g.animManager != nil {
		g.animManager.StartKillAnim(playerIdx, tokenIdx, homeX, homeY)
	}
}

// startExtraAnim starts extra turn flash
func (g *Game) startExtraAnim() {
	if g.animManager != nil {
		g.animManager.StartExtraAnim()
	}
}

// updateMenu handles the main menu
func (g *Game) updateMenu() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Local Game button
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 250, 200, 40) {
			g.gameModeSelected = false
			g.isMultiplayer = false
			g.state = StateSelectingPlayers
			g.clickCooldown = 20
			return nil
		}

		// Multiplayer button
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 320, 200, 40) {
			g.gameModeSelected = true
			g.isMultiplayer = true
			g.state = StateEnterName
			g.clickCooldown = 20
			return nil
		}
	}

	return nil
}

// updateEnterName handles the name entry screen
func (g *Game) updateEnterName() error {
	// Handle text input
	g.HandleTextInput()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Continue button
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 400, 200, 40) {
			if len(g.inputText) > 0 {
				g.playerName = g.inputText
				g.state = StateCreateOrJoin
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
	g.HandleTextInput()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.clickCooldown == 0 {
		mx, my := ebiten.CursorPosition()

		// Create Room button
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 220, 200, 40) {
			g.InitMultiplayer()
			return nil
		}

		// Join Room button
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 450, 200, 40) {
			if len(g.joinRoomID) > 0 {
				g.InitMultiplayer()
			}
			return nil
		}

		// Room ID input area
		if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 380, 200, 40) {
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
		if g.currentRoom != nil && g.IsButtonClicked(mx, my, 50, 170, 120, 30) {
			if roomURL := network.GetRoomURL(g.currentRoom.ID); roomURL != "" {
				network.CopyToClipboard(roomURL)
			}
			g.clickCooldown = 20
			return nil
		}

		// Start Game button (host only)
		if g.localPlayer != nil && g.localPlayer.IsHost {
			if g.IsButtonClicked(mx, my, ScreenWidth/2-100, 500, 200, 40) {
				if g.networkClient != nil {
					g.networkClient.StartGame()
				}
				g.clickCooldown = 20
				return nil
			}
		}

		// Back button
		if g.IsButtonClicked(mx, my, 50, 600, 100, 40) {
			if g.networkClient != nil {
				g.networkClient.Disconnect()
			}
			g.state = StateMenu
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
		if g.IsButtonClicked(mx, my, 50, 600, 100, 40) {
			g.state = StateMenu
			g.clickCooldown = 20
			return nil
		}

		// Selection buttons
		for i := 1; i <= 4; i++ {
			bx := ScreenWidth/2 - 100
			by := 200 + i*60
			if mx >= bx && mx <= bx+200 && my >= by && my <= by+40 {
				g.InitGame(i)
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
			g.state = StateWaitingRoom
		} else {
			g.state = StateSelectingPlayers
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

				validTokens := game.GetValidTokens(g.players[g.currentPlayer], g.rollResult)
				if len(validTokens) == 0 {
					g.nextTurn()
				}
			}
		}

		if g.hasRolled {
			validTokens := game.GetValidTokens(g.players[g.currentPlayer], g.rollResult)
			player := g.players[g.currentPlayer]

			for _, idx := range validTokens {
				token := player.Tokens[idx]
				row, col := game.GetCellCoordinates(token, token.PlayerIdx)
				tx := BoardOffsetX + col*game.CellSize
				ty := BoardOffsetY + row*game.CellSize

				if mx >= tx && mx <= tx+game.CellSize && my >= ty && my <= ty+game.CellSize {
					g.moveToken(idx)
					g.clickCooldown = 10

					// Send action to server in multiplayer
					if g.isMultiplayer && g.networkClient != nil {
						g.networkClient.SendGameAction("move", idx)
						return nil
					}

					if g.state == StatePlaying {
						g.nextTurn()
					}
					break
				}
			}
		}
	}

	return nil
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

// moveToken moves a token based on the roll result
func (g *Game) moveToken(tokenIdx int) {
	player := g.players[g.currentPlayer]
	if player == nil {
		return
	}
	token := player.Tokens[tokenIdx]
	roll := g.rollResult

	if token.State == game.TokenAtStart {
		token.State = game.TokenOnBoard
		g.applyMove(token, roll)
	} else if token.State == game.TokenOnBoard {
		g.applyMove(token, roll)
	}
}

// applyMove applies the roll to a token, handling transitions
func (g *Game) applyMove(token *game.Token, roll int) {
	player := g.players[token.PlayerIdx]

	if token.Position >= 16 {
		// On inner path
		step := token.Position - 16
		newStep := step + roll
		if newStep >= 8 {
			token.State = game.TokenFinished
			token.Position = 24
			g.extraTurn = true // Rule: Reaching center grants a bonus roll
			g.checkWin()
		} else {
			token.Position = 16 + newStep
			g.checkKill(token) // Rule: Killing possible in inner circle too
		}
		return
	}

	// On outer path
	distToEntry := (player.EntryPos - token.Position + 16) % 16
	if roll > distToEntry {
		// Transition to inner path
		remaining := roll - distToEntry - 1
		if remaining >= 8 {
			token.State = game.TokenFinished
			token.Position = 24
			g.extraTurn = true // Rule: Reaching center grants a bonus roll
			g.checkWin()
		} else {
			token.Position = 16 + remaining
			g.checkKill(token) // Rule: Killing possible in inner circle too
		}
	} else {
		// Stay on outer path
		token.Position = (token.Position + roll) % 16
		g.checkKill(token)
	}
}

// checkKill checks if the moved token lands on an opponent and kills them
func (g *Game) checkKill(token *game.Token) {
	if token.State != game.TokenOnBoard {
		return
	}

	// Get current grid coordinates
	row, col := game.GetCellCoordinates(token, token.PlayerIdx)

	// Rule: Starting blocks are safe houses. No killing can happen on a safe house.
	for i := 0; i < 4; i++ {
		startCoords := game.OuterPath[game.PlayerStartPositions[i]]
		if row == startCoords[0] && col == startCoords[1] {
			return // Safe house, no kill
		}
	}

	// Rule: If landing on an opponent's token in a non-safe block, kill it.
	killedAny := false
	for _, player := range g.players {
		if player == nil || player.Idx == token.PlayerIdx {
			continue // Cannot kill own tokens or from non-existent players
		}
		for _, opponentToken := range player.Tokens {
			if opponentToken.State == game.TokenOnBoard {
				or, oc := game.GetCellCoordinates(opponentToken, opponentToken.PlayerIdx)
				if or == row && oc == col {
					// Kill opponent token and send it back to starting point
					opponentToken.State = game.TokenAtStart
					opponentToken.Position = game.PlayerStartPositions[opponentToken.PlayerIdx]
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
	for _, token := range player.Tokens {
		if token.State == game.TokenFinished {
			finishedCount++
		}
	}
	if finishedCount == game.TokensPerPlayer {
		g.state = StateGameOver
		g.winner = g.currentPlayer
	}
}
