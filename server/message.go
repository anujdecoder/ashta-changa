package server

import (
	"encoding/json"
	"log"
	"time"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// CreateRoomRequest represents a request to create a room
type CreateRoomRequest struct {
	RoomName string `json:"roomName"`
	PlayerName string `json:"playerName"`
}

// CreateRoomResponse represents a response to create room request
type CreateRoomResponse struct {
	Room   *Room  `json:"room"`
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
	Success bool   `json:"success"`
	Room    *Room  `json:"room,omitempty"`
	Player  *Player `json:"player,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StartGameRequest represents a request to start the game
type StartGameRequest struct {
	RoomID string `json:"roomId"`
}

// GameActionRequest represents a game action from a player
type GameActionRequest struct {
	RoomID     string `json:"roomId"`
	Action     string `json:"action"` // "roll", "move"
	TokenIdx   int    `json:"tokenIdx,omitempty"`
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
		ID:       playerID,
		Name:     req.PlayerName,
		RoomID:   room.ID,
		IsHost:   true,
		IsReady:  false,
		Conn:     client,
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
		ID:       playerID,
		Name:     req.PlayerName,
		RoomID:   room.ID,
		IsHost:   playerID == room.HostID,
		IsReady:  false,
		Conn:     client,
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
		// Handle roll action - this will be processed by the game logic
		// For now, broadcast the action to all players
		room.Broadcast(Message{
			Type: "playerRolled",
			Payload: map[string]interface{}{
				"playerIdx": client.player.PlayerIdx,
			},
		})

	case "move":
		// Handle move action
		room.Broadcast(Message{
			Type: "playerMoved",
			Payload: map[string]interface{}{
				"playerIdx": client.player.PlayerIdx,
				"tokenIdx":  req.TokenIdx,
			},
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
