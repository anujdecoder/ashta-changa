// Package gui contains the Ebiten-based GUI for Ashta Changa.
package gui

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/anujdecoder/ashta-board/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{60, 40, 20, 255})

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StateEnterName:
		g.drawEnterName(screen)
	case StateCreateOrJoin:
		g.drawCreateOrJoin(screen)
	case StateWaitingRoom:
		g.drawWaitingRoom(screen)
	case StateConnecting:
		g.drawConnecting(screen)
	case StateSelectingPlayers:
		g.drawSelectionScreen(screen)
	case StatePlaying, StateGameOver:
		g.drawGame(screen)
	}

	// Draw error message if any
	if g.showError {
		g.drawError(screen)
	}
}

// drawMenu draws the main menu
func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ASHTA CHANGA", ScreenWidth/2-50, 100)
	ebitenutil.DebugPrintAt(screen, "Traditional Indian Board Game", ScreenWidth/2-100, 130)

	// Local Game button
	g.drawButton(screen, ScreenWidth/2-100, 250, 200, 40, "Local Game", color.RGBA{139, 90, 43, 255})

	// Multiplayer button
	g.drawButton(screen, ScreenWidth/2-100, 320, 200, 40, "Multiplayer", color.RGBA{139, 90, 43, 255})

	// Instructions
	ebitenutil.DebugPrintAt(screen, "2-4 Players | Roll shells to move tokens", ScreenWidth/2-120, 450)
	ebitenutil.DebugPrintAt(screen, "Reach the center with all 4 tokens to win!", ScreenWidth/2-130, 470)
}

// drawEnterName draws the name entry screen
func (g *Game) drawEnterName(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ENTER YOUR NAME", ScreenWidth/2-60, 150)

	// Input box
	vector.DrawFilledRect(screen, float32(ScreenWidth/2-150), 250, 300, 50, color.RGBA{200, 200, 200, 255}, false)
	vector.StrokeRect(screen, float32(ScreenWidth/2-150), 250, 300, 50, 2, color.RGBA{255, 255, 255, 255}, false)

	// Show input text
	if len(g.inputText) > 0 {
		ebitenutil.DebugPrintAt(screen, g.inputText, ScreenWidth/2-140, 265)
	} else {
		ebitenutil.DebugPrintAt(screen, "Type your name...", ScreenWidth/2-140, 265)
	}

	// Continue button
	btnColor := color.RGBA{100, 100, 100, 255}
	if len(g.inputText) > 0 {
		btnColor = color.RGBA{139, 90, 43, 255}
	}
	g.drawButton(screen, ScreenWidth/2-100, 400, 200, 40, "Continue", btnColor)
}

// drawCreateOrJoin draws the create/join room screen
func (g *Game) drawCreateOrJoin(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "MULTIPLAYER", ScreenWidth/2-40, 80)

	// Create Room section
	ebitenutil.DebugPrintAt(screen, "Create a new room:", ScreenWidth/2-70, 180)
	g.drawButton(screen, ScreenWidth/2-100, 220, 200, 40, "Create Room", color.RGBA{139, 90, 43, 255})

	// Divider
	vector.StrokeLine(screen, float32(ScreenWidth/2-150), 300, float32(ScreenWidth/2+150), 300, 2, color.RGBA{150, 150, 150, 255}, false)
	ebitenutil.DebugPrintAt(screen, "OR", ScreenWidth/2-10, 310)

	// Join Room section
	ebitenutil.DebugPrintAt(screen, "Join with Room ID:", ScreenWidth/2-70, 340)

	// Room ID input box
	inputColor := color.RGBA{200, 200, 200, 255}
	if g.inputActive {
		inputColor = color.RGBA{230, 230, 230, 255}
	}
	vector.DrawFilledRect(screen, float32(ScreenWidth/2-100), 380, 200, 40, inputColor, false)
	vector.StrokeRect(screen, float32(ScreenWidth/2-100), 380, 200, 40, 2, color.RGBA{255, 255, 255, 255}, false)

	// Show room ID
	displayText := g.joinRoomID
	if len(displayText) == 0 {
		displayText = "Enter Room ID"
	}
	ebitenutil.DebugPrintAt(screen, displayText, ScreenWidth/2-90, 395)

	// Join button
	btnColor := color.RGBA{100, 100, 100, 255}
	if len(g.joinRoomID) > 0 {
		btnColor = color.RGBA{139, 90, 43, 255}
	}
	g.drawButton(screen, ScreenWidth/2-100, 450, 200, 40, "Join Room", btnColor)
}

// drawWaitingRoom draws the waiting room screen
func (g *Game) drawWaitingRoom(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "WAITING ROOM", ScreenWidth/2-50, 50)

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
			g.drawButton(screen, ScreenWidth/2-100, 500, 200, 40, "Start Game", color.RGBA{0, 150, 0, 255})
			ebitenutil.DebugPrintAt(screen, "Waiting for players...", ScreenWidth/2-80, 560)
		} else {
			ebitenutil.DebugPrintAt(screen, "Waiting for host to start...", ScreenWidth/2-100, 500)
		}
	}

	// Back button
	g.drawButton(screen, 50, 600, 100, 40, "Back", color.RGBA{139, 90, 43, 255})
}

// drawConnecting draws the connecting screen
func (g *Game) drawConnecting(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "CONNECTING...", ScreenWidth/2-50, ScreenHeight/2)
	ebitenutil.DebugPrintAt(screen, "Please wait", ScreenWidth/2-40, ScreenHeight/2+30)
}

// drawSelectionScreen draws the player selection UI
func (g *Game) drawSelectionScreen(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ASHTA CHANGA", ScreenWidth/2-50, 100)
	ebitenutil.DebugPrintAt(screen, "Select Number of Players", ScreenWidth/2-80, 150)

	for i := 1; i <= 4; i++ {
		bx := ScreenWidth/2 - 100
		by := 200 + i*60
		g.drawButton(screen, bx, by, 200, 40, strconv.Itoa(i)+" Players", color.RGBA{139, 90, 43, 255})
	}

	// Back button
	g.drawButton(screen, 50, 600, 100, 40, "Back", color.RGBA{139, 90, 43, 255})
}

// drawBoard draws the game board grid
func (g *Game) drawBoard(screen *ebiten.Image) {
	for r := 0; r < game.BoardSize; r++ {
		for c := 0; c < game.BoardSize; c++ {
			cell := g.board.Cells[r][c]
			if cell == nil {
				continue
			}

			x := BoardOffsetX + c*game.CellSize
			y := BoardOffsetY + r*game.CellSize

			// Determine cell color
			cellColor := color.RGBA{100, 70, 50, 255} // Default path color

			if cell.IsCenter {
				cellColor = color.RGBA{139, 90, 43, 255} // Center color
			} else if cell.IsStart {
				// Use player color for start positions
				if cell.PlayerIdx >= 0 && cell.PlayerIdx < 4 {
					cellColor = game.PlayerColors[cell.PlayerIdx]
				}
			} else if !cell.IsPath {
				cellColor = color.RGBA{60, 40, 20, 255} // Background color
			}

			// Draw cell
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(game.CellSize), float32(game.CellSize), cellColor, false)
			vector.StrokeRect(screen, float32(x), float32(y), float32(game.CellSize), float32(game.CellSize), 1, color.RGBA{40, 30, 20, 255}, false)
		}
	}
}

// drawGame draws the game board
func (g *Game) drawGame(screen *ebiten.Image) {
	g.drawBoard(screen)
	g.drawTokens(screen)
	g.drawConchShellArea(screen)
	g.drawPlayerInfo(screen)

	title := "Ashta Changa - Traditional Indian Board Game"
	if g.isMultiplayer {
		title = "Ashta Changa - Multiplayer"
	}
	ebitenutil.DebugPrintAt(screen, title, 50, 10)

	if g.state == StateGameOver {
		vector.DrawFilledRect(screen, 200, 250, 400, 100, color.RGBA{0, 0, 0, 200}, false)
		ebitenutil.DebugPrintAt(screen, "GAME OVER!", 350, 280)
		ebitenutil.DebugPrintAt(screen, game.PlayerNames[g.winner]+" Wins!", 350, 300)
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

// drawTokens draws all tokens on the board
func (g *Game) drawTokens(screen *ebiten.Image) {
	type pos struct{ r, c int }
	tokensByCell := make(map[pos][]*game.Token)

	for _, player := range g.players {
		if player == nil {
			continue
		}
		for _, token := range player.Tokens {
			r, c := game.GetCellCoordinates(token, token.PlayerIdx)
			p := pos{r, c}
			tokensByCell[p] = append(tokensByCell[p], token)
		}
	}

	for cellPos, tokens := range tokensByCell {
		cellX := BoardOffsetX + cellPos.c*game.CellSize
		cellY := BoardOffsetY + cellPos.r*game.CellSize

		for i, token := range tokens {
			numInRow := 2
			if len(tokens) > 4 {
				numInRow = 3
			}
			if len(tokens) > 9 {
				numInRow = 4
			}

			offsetX := (i%numInRow)*(game.CellSize/(numInRow+1)) + (game.CellSize / (numInRow + 1))
			offsetY := (i/numInRow)*(game.CellSize/(numInRow+1)) + (game.CellSize / (numInRow + 1))

			tx := cellX + offsetX
			ty := cellY + offsetY

			tokenColor := game.PlayerColors[token.PlayerIdx]
			isValid := false
			if g.hasRolled && token.PlayerIdx == g.currentPlayer {
				if !g.isMultiplayer || (g.localPlayer != nil && g.currentPlayer == g.localPlayer.PlayerIdx) {
					for _, idx := range game.GetValidTokens(g.players[g.currentPlayer], g.rollResult) {
						if g.players[g.currentPlayer].Tokens[idx] == token {
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
				numStr := string(rune('1' + token.ID))
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

	playerColor := game.PlayerColors[g.currentPlayer]
	vector.DrawFilledRect(screen, float32(infoX+110), float32(infoY), 60, 20, playerColor, false)
	ebitenutil.DebugPrintAt(screen, game.PlayerNames[g.currentPlayer], infoX+115, infoY+2)

	if g.extraTurn {
		// Pulsing animation for extra turn
		alpha := uint8(255)
		animExtra := g.animManager.GetExtraAnimCount()
		if animExtra > 0 && animExtra%10 < 5 {
			alpha = 100
		}
		extraColor := color.RGBA{255, 255, 0, alpha}
		vector.DrawFilledRect(screen, float32(infoX), float32(infoY+25), 100, 16, extraColor, false)
		ebitenutil.DebugPrintAt(screen, "EXTRA TURN!", infoX, infoY+25)
	}

	if g.hasRolled {
		validTokens := game.GetValidTokens(g.players[g.currentPlayer], g.rollResult)
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
			ebitenutil.DebugPrintAt(screen, game.PlayerNames[g.currentPlayer]+" is rolling...", infoX, infoY+50)
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

	animRoll := g.animManager.GetRollAnimCount()
	for i := 0; i < 4; i++ {
		shellX := shellAreaX + 40
		shellY := shellStartY + i*shellSpacing
		// Pre-roll animation: shake shells
		if animRoll > 0 {
			shake := float64(animRoll%6 - 3) // -3 to +3
			shellX += int(shake)
			shellY += int(shake * 0.5)
		}
		g.drawConchShell(screen, shellX, shellY, g.conchShells.Shells[i].OpenSideUp)
	}

	resultY := shellStartY + 4*shellSpacing + 20
	ebitenutil.DebugPrintAt(screen, "Result:", shellAreaX+60, resultY)

	// Show result only when not animating and have result
	showResult := g.rollResult > 0 && animRoll == 0
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
	} else if animRoll > 0 {
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
