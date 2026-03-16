package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/anujdecoder/ashta-board/game"
)

// generatePlayerID generates a unique player ID
func generatePlayerID() string {
	return time.Now().Format(PlayerIDTimestampFormat) + randomString(RandomStringLength)
}

// randomString generates a random string of given length
func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = PlayerIDCharset[time.Now().UnixNano()%int64(len(PlayerIDCharset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

// handleMessage handles incoming WebSocket messages
func handleMessage(client *Client, data []byte, roomManager RoomManagerInterface) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		sendError(client, "Invalid message format")
		return
	}

	switch msg.Type {
	case "createRoom":
		handleCreateRoom(client, msg.Payload, roomManager)
	case "joinRoom":
		handleJoinRoom(client, msg.Payload, roomManager)
	case "startGame":
		handleStartGame(client, msg.Payload, roomManager)
	case "gameAction":
		handleGameAction(client, msg.Payload, roomManager)
	case "getRooms":
		handleGetRooms(client, roomManager)
	case "setReady":
		handleSetReady(client, msg.Payload, roomManager)
	default:
		sendError(client, "Unknown message type: "+msg.Type)
	}
}

// handleCreateRoom handles room creation
func handleCreateRoom(client *Client, payload interface{}, roomManager RoomManagerInterface) {
	payloadBytes, _ := json.Marshal(payload)
	var req CreateRoomRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		sendError(client, "Invalid create room request")
		return
	}

	playerID := generatePlayerID()
	room := roomManager.CreateRoom(req.RoomName, playerID)

	player := &Player{
		ID:      playerID,
		Name:    req.PlayerName,
		RoomID:  room.ID,
		IsHost:  true,
		IsReady: false,
		Conn:    client,
	}

	room.AddPlayer(player)
	client.player = player

	response := Message{
		Type: MessageTypeRoomCreated,
		Payload: CreateRoomResponse{
			Room:   room,
			Player: player,
		},
	}

	sendMessage(client, response)
}

// handleJoinRoom handles joining a room
func handleJoinRoom(client *Client, payload interface{}, roomManager RoomManagerInterface) {
	payloadBytes, _ := json.Marshal(payload)
	var req JoinRoomRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		sendError(client, "Invalid join room request")
		return
	}

	room, exists := roomManager.GetRoom(req.RoomID)
	if !exists {
		sendError(client, "Room not found")
		return
	}

	if room.GameStarted {
		sendError(client, "Game already started")
		return
	}

	if len(room.Players) >= room.MaxPlayers {
		sendError(client, "Room is full")
		return
	}

	playerID := req.PlayerID
	if playerID == "" {
		playerID = generatePlayerID()
	}

	player := &Player{
		ID:      playerID,
		Name:    req.PlayerName,
		RoomID:  room.ID,
		IsHost:  playerID == room.HostID,
		IsReady: false,
		Conn:    client,
	}

	if !room.AddPlayer(player) {
		sendError(client, "Failed to join room")
		return
	}

	client.player = player

	// Send success response to joining player
	response := Message{
		Type: MessageTypeRoomJoined,
		Payload: JoinRoomResponse{
			Success: true,
			Room:    room,
			Player:  player,
		},
	}
	sendMessage(client, response)

	// Notify other players
	room.BroadcastToOthers(player.ID, Message{
		Type: MessageTypePlayerJoined,
		Payload: PlayerJoinedPayload{
			Player: player,
			Room:   room,
		},
	})
}

// handleStartGame handles starting the game
func handleStartGame(client *Client, payload interface{}, roomManager RoomManagerInterface) {
	payloadBytes, _ := json.Marshal(payload)
	var req StartGameRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		sendError(client, "Invalid start game request")
		return
	}

	room, exists := roomManager.GetRoom(req.RoomID)
	if !exists {
		sendError(client, "Room not found")
		return
	}

	// Only host can start the game
	if client.player == nil || !client.player.IsHost {
		sendError(client, "Only host can start the game")
		return
	}

	// Check if game is already started
	if room.GameStarted {
		// Just send the current game state to this player
		sendMessage(client, Message{
			Type:    MessageTypeGameStarted,
			Payload: room.GameState,
		})
		return
	}

	if !room.StartGame() {
		sendError(client, "Failed to start game")
		return
	}

	// Broadcast game started to all players
	room.Broadcast(Message{
		Type:    MessageTypeGameStarted,
		Payload: room.GameState,
	})
}

// handleGameAction handles game actions (roll, move)
func handleGameAction(client *Client, payload interface{}, roomManager RoomManagerInterface) {
	payloadBytes, _ := json.Marshal(payload)
	var req GameActionRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		sendError(client, "Invalid game action request")
		return
	}

	room, exists := roomManager.GetRoom(req.RoomID)
	if !exists {
		sendError(client, "Room not found")
		return
	}

	if !room.GameStarted {
		sendError(client, "Game not started")
		return
	}

	// Verify it's the player's turn
	if client.player == nil || client.player.PlayerIdx != room.GameState.CurrentPlayer {
		sendError(client, "Not your turn")
		return
	}

	switch req.Action {
	case ActionRoll:
		room.mu.Lock()
		if room.GameState.HasRolled {
			room.mu.Unlock()
			sendError(client, "Already rolled")
			return
		}

		// Roll shells server-side so all clients are in sync
		shellStates := make([]bool, DefaultMaxPlayers)
		for i := 0; i < DefaultMaxPlayers; i++ {
			shellStates[i] = rand.Intn(2) == 1
		}

		// Use game package for roll result calculation
		rollResult := game.RollFromShellStates(shellStates)

		room.GameState.RollResult = rollResult
		room.GameState.HasRolled = true
		room.GameState.ShellStates = shellStates
		if rollResult == 4 || rollResult == 8 {
			room.GameState.ExtraTurn = true
		}

		// Check if current player has any valid token moves
		// If no valid moves possible, auto-advance turn
		currentPlayerIdx := room.GameState.CurrentPlayer
		hasValidMove := false

		// Find current player's tokens
		for _, pt := range room.GameState.PlayerTokens {
			if pt.PlayerIdx == currentPlayerIdx {
				// Create a temporary player for CanMoveToken check
				tempPlayer := &game.Player{
					Idx:      currentPlayerIdx,
					StartPos: game.PlayerStartPositions[currentPlayerIdx],
					EntryPos: game.PlayerEntryPositions[currentPlayerIdx],
				}
				for _, t := range pt.Tokens {
					// Create a temporary token for CanMoveToken check
					tempToken := &game.Token{
						ID:        t.ID,
						PlayerIdx: currentPlayerIdx,
						State:     game.TokenState(t.State),
						Position:  t.Position,
					}
					if game.CanMoveToken(tempToken, rollResult, tempPlayer) {
						hasValidMove = true
						break
					}
				}
				break
			}
		}

		if !hasValidMove {
			// No valid moves - auto-advance turn (no extra turn possible)
			// Keep roll result/shells visible for this update
			room.GameState.ExtraTurn = false
			room.GameState.HasRolled = false
			room.GameState.CurrentPlayer = (currentPlayerIdx + 1) % room.GameState.NumActivePlayers
		}
		room.mu.Unlock()

		// Broadcast full game state update to all players so they sync
		room.Broadcast(Message{
			Type:    MessageTypeGameStateUpdate,
			Payload: room.GameState,
		})

	case ActionMove:
		// Handle move action - update token state on server
		room.mu.Lock()
		if !room.GameState.HasRolled {
			room.mu.Unlock()
			sendError(client, "Must roll before moving")
			return
		}
		tokenIdx := req.TokenIdx
		playerIdx := client.player.PlayerIdx

		// Find and update the token
		for i, pt := range room.GameState.PlayerTokens {
			if pt.PlayerIdx == playerIdx {
				if tokenIdx >= 0 && tokenIdx < len(pt.Tokens) {
					token := &room.GameState.PlayerTokens[i].Tokens[tokenIdx]

					// Apply move logic (same rules as client)
					roll := room.GameState.RollResult

					if token.State == TokenStateAtStart { // atStart
						if roll == 1 || roll == 4 || roll == 8 {
							token.State = TokenStateOnBoard // onBoard
							// Move from start position on outer path
							token.Position = (token.Position + roll) % 16
						} else {
							room.mu.Unlock()
							sendError(client, "Invalid move")
							return
						}
					} else if token.State == TokenStateOnBoard { // onBoard
						if token.Position >= 16 {
							// Inner path
							step := token.Position - 16
							if step+roll >= 8 {
								token.State = TokenStateFinished
								token.Position = 24
								room.GameState.ExtraTurn = true
							} else {
								token.Position = 16 + step + roll
							}
						} else {
							// Outer path
							entryPos := playerIdx*4 + 1 // entry is one before start
							distToEntry := (entryPos - token.Position + 16) % 16
							if roll > distToEntry {
								remaining := roll - distToEntry - 1
								if remaining >= 8 {
									token.State = TokenStateFinished
									token.Position = 24
									room.GameState.ExtraTurn = true
								} else {
									token.Position = 16 + remaining
								}
							} else {
								token.Position = (token.Position + roll) % 16
							}
						}
					}

					// Check for kill on outer path and inner edge (not on safe houses)
					if token.State == TokenStateOnBoard {
						isSafe := false
						// Safe houses are only on outer path at positions 2, 6, 10, 14
						if token.Position < 16 {
							isSafe = token.Position == 2 || token.Position == 6 ||
								token.Position == 10 || token.Position == 14
						}
						// Inner edge (position >= 16) has no safe houses
						if !isSafe {
							for j, opt := range room.GameState.PlayerTokens {
								if opt.PlayerIdx != playerIdx {
									for k, ot := range opt.Tokens {
										if ot.State == TokenStateOnBoard && ot.Position == token.Position {
											room.GameState.PlayerTokens[j].Tokens[k].State = TokenStateAtStart
											room.GameState.PlayerTokens[j].Tokens[k].Position = opt.PlayerIdx*DefaultMaxPlayers + 2
											room.GameState.ExtraTurn = true
										}
									}
								}
							}
						}
					}

					// Check win condition
					allFinished := true
					for _, t := range room.GameState.PlayerTokens[i].Tokens {
						if t.State != TokenStateFinished {
							allFinished = false
							break
						}
					}
					if allFinished {
						room.GameState.Winner = playerIdx
					}
				}
				break
			}
		}

		if room.GameState.Winner < 0 {
			if room.GameState.ExtraTurn {
				room.GameState.ExtraTurn = false
			} else {
				room.GameState.CurrentPlayer = (room.GameState.CurrentPlayer + 1) % room.GameState.NumActivePlayers
			}
			room.GameState.HasRolled = false
			room.GameState.RollResult = 0
			room.GameState.ShellStates = make([]bool, DefaultMaxPlayers)
		}
		room.mu.Unlock()

		// Broadcast full game state update to all players
		room.Broadcast(Message{
			Type:    MessageTypeGameStateUpdate,
			Payload: room.GameState,
		})
	}
}

// handleGetRooms handles getting list of rooms
func handleGetRooms(client *Client, roomManager RoomManagerInterface) {
	rooms := roomManager.ListRooms()
	response := Message{
		Type:    MessageTypeRoomsList,
		Payload: rooms,
	}
	sendMessage(client, response)
}

// handleSetReady handles player ready status
func handleSetReady(client *Client, payload interface{}, roomManager RoomManagerInterface) {
	payloadBytes, _ := json.Marshal(payload)
	var req struct {
		RoomID  string `json:"roomId"`
		IsReady bool   `json:"isReady"`
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		sendError(client, "Invalid ready request")
		return
	}

	room, exists := roomManager.GetRoom(req.RoomID)
	if !exists {
		sendError(client, "Room not found")
		return
	}

	if client.player != nil {
		client.player.IsReady = req.IsReady

		// Notify all players about ready status change
		room.Broadcast(Message{
			Type: MessageTypePlayerReady,
			Payload: map[string]interface{}{
				"playerId": client.player.ID,
				"isReady":  req.IsReady,
			},
		})
	}
}

// sendMessage sends a message to a client
func sendMessage(client *Client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	client.send <- data
}

// sendError sends an error message to a client
func sendError(client *Client, errorMsg string) {
	msg := Message{
		Type: MessageTypeError,
		Payload: map[string]string{
			"message": errorMsg,
		},
	}
	sendMessage(client, msg)
}
