package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// CreateRoomRequest represents a request to create a room
type CreateRoomRequest struct {
	RoomName   string `json:"roomName"`
	PlayerName string `json:"playerName"`
}

// CreateRoomResponse represents a response to create room request
type CreateRoomResponse struct {
	Room   *Room   `json:"room"`
	Player *Player `json:"player"`
}

// JoinRoomRequest represents a request to join a room
type JoinRoomRequest struct {
	RoomID     string `json:"roomId"`
	PlayerName string `json:"playerName"`
	PlayerID   string `json:"playerId,omitempty"`
}

// JoinRoomResponse represents a response to join room request
type JoinRoomResponse struct {
	Success bool    `json:"success"`
	Room    *Room   `json:"room,omitempty"`
	Player  *Player `json:"player,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// StartGameRequest represents a request to start the game
type StartGameRequest struct {
	RoomID string `json:"roomId"`
}

// GameActionRequest represents a game action from a player
type GameActionRequest struct {
	RoomID   string `json:"roomId"`
	Action   string `json:"action"` // "roll", "move"
	TokenIdx int    `json:"tokenIdx,omitempty"`
}

// GameStateUpdate represents a game state update
type GameStateUpdate struct {
	GameState *GameState `json:"gameState"`
}

// PlayerJoinedPayload represents payload for player joined event
type PlayerJoinedPayload struct {
	Player *Player `json:"player"`
	Room   *Room   `json:"room"`
}

// generatePlayerID generates a unique player ID
func generatePlayerID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

// handleMessage handles incoming WebSocket messages
func handleMessage(client *Client, data []byte, roomManager *RoomManager) {
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
func handleCreateRoom(client *Client, payload interface{}, roomManager *RoomManager) {
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
		Type: "roomCreated",
		Payload: CreateRoomResponse{
			Room:   room,
			Player: player,
		},
	}

	sendMessage(client, response)
}

// handleJoinRoom handles joining a room
func handleJoinRoom(client *Client, payload interface{}, roomManager *RoomManager) {
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
		Type: "roomJoined",
		Payload: JoinRoomResponse{
			Success: true,
			Room:    room,
			Player:  player,
		},
	}
	sendMessage(client, response)

	// Notify other players
	room.BroadcastToOthers(player.ID, Message{
		Type: "playerJoined",
		Payload: PlayerJoinedPayload{
			Player: player,
			Room:   room,
		},
	})
}

// handleStartGame handles starting the game
func handleStartGame(client *Client, payload interface{}, roomManager *RoomManager) {
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
			Type:    "gameStarted",
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
		Type:    "gameStarted",
		Payload: room.GameState,
	})
}

// handleGameAction handles game actions (roll, move)
func handleGameAction(client *Client, payload interface{}, roomManager *RoomManager) {
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
	case "roll":
		room.mu.Lock()
		if room.GameState.HasRolled {
			room.mu.Unlock()
			sendError(client, "Already rolled")
			return
		}

		// Roll shells server-side so all clients are in sync
		shellStates := make([]bool, 4)
		openCount := 0
		for i := 0; i < 4; i++ {
			open := rand.Intn(2) == 1
			shellStates[i] = open
			if open {
				openCount++
			}
		}

		rollResult := 0
		switch openCount {
		case 0:
			rollResult = 8
		case 1:
			rollResult = 1
		case 2:
			rollResult = 2
		case 3:
			rollResult = 3
		case 4:
			rollResult = 4
		}

		room.GameState.RollResult = rollResult
		room.GameState.HasRolled = true
		room.GameState.ShellStates = shellStates
		if rollResult == 4 || rollResult == 8 {
			room.GameState.ExtraTurn = true
		}

		// Check if current player has all tokens at home (start position)
		// If so and roll is not 1, 4, or 8, they cannot move - auto-advance turn
		currentPlayerIdx := room.GameState.CurrentPlayer
		allAtHome := true
		for _, pt := range room.GameState.PlayerTokens {
			if pt.PlayerIdx == currentPlayerIdx {
				for _, t := range pt.Tokens {
					if t.State != 0 { // not at start
						allAtHome = false
						break
					}
				}
				break
			}
		}

		if allAtHome && rollResult != 1 && rollResult != 4 && rollResult != 8 {
			// Cannot move out - auto-advance turn (no extra turn possible)
			// Keep roll result/shells visible for this update
			room.GameState.ExtraTurn = false
			room.GameState.HasRolled = false
			room.GameState.CurrentPlayer = (currentPlayerIdx + 1) % room.GameState.NumActivePlayers
		}
		room.mu.Unlock()

		// Broadcast full game state update to all players so they sync
		room.Broadcast(Message{
			Type:    "gameStateUpdate",
			Payload: room.GameState,
		})

	case "move":
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

					if token.State == 0 { // atStart
						if roll == 1 || roll == 4 || roll == 8 {
							token.State = 1 // onBoard
							// Move from start position on outer path
							token.Position = (token.Position + roll) % 16
						} else {
							room.mu.Unlock()
							sendError(client, "Invalid move")
							return
						}
					} else if token.State == 1 { // onBoard
						if token.Position >= 16 {
							// Inner path
							step := token.Position - 16
							if step+roll >= 8 {
								token.State = 2
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
									token.State = 2
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

					// Check for kill on outer path (not on safe houses)
					if token.State == 1 && token.Position < 16 {
						isSafe := token.Position == 2 || token.Position == 6 ||
							token.Position == 10 || token.Position == 14
						if !isSafe {
							for j, opt := range room.GameState.PlayerTokens {
								if opt.PlayerIdx != playerIdx {
									for k, ot := range opt.Tokens {
										if ot.State == 1 && ot.Position == token.Position {
											room.GameState.PlayerTokens[j].Tokens[k].State = 0
											room.GameState.PlayerTokens[j].Tokens[k].Position = opt.PlayerIdx*4 + 2
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
						if t.State != 2 {
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
			room.GameState.ShellStates = make([]bool, 4)
		}
		room.mu.Unlock()

		// Broadcast full game state update to all players
		room.Broadcast(Message{
			Type:    "gameStateUpdate",
			Payload: room.GameState,
		})
	}
}

// handleGetRooms handles getting list of rooms
func handleGetRooms(client *Client, roomManager *RoomManager) {
	rooms := roomManager.ListRooms()
	response := Message{
		Type:    "roomsList",
		Payload: rooms,
	}
	sendMessage(client, response)
}

// handleSetReady handles player ready status
func handleSetReady(client *Client, payload interface{}, roomManager *RoomManager) {
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
			Type: "playerReady",
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
		Type: "error",
		Payload: map[string]string{
			"message": errorMsg,
		},
	}
	sendMessage(client, msg)
}
