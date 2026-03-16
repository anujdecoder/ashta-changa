// Package network contains the network client for multiplayer gameplay.
package network

import (
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"
	"time"
)

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
		if nc.OnError != nil {
			nc.OnError("Disconnected from server")
		}
		return nil
	})

	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if nc.OnError != nil {
			nc.OnError("Connection error")
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
			if nc.OnRoomCreated != nil {
				nc.OnRoomCreated(resp.Room, resp.Player)
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
				if nc.OnRoomJoined != nil {
					nc.OnRoomJoined(resp.Room, resp.Player)
				}
			} else if nc.OnError != nil {
				nc.OnError(resp.Error)
			}
		}

	case "playerJoined":
		var payload struct {
			Player *NetworkPlayer `json:"player"`
			Room   *NetworkRoom   `json:"room"`
		}
		if err := json.Unmarshal(payloadBytes, &payload); err == nil && nc.OnPlayerJoined != nil {
			nc.OnPlayerJoined(payload.Player)
		}

	case "playerLeft":
		var player NetworkPlayer
		if err := json.Unmarshal(payloadBytes, &player); err == nil && nc.OnPlayerLeft != nil {
			nc.OnPlayerLeft(&player)
		}

	case "gameStarted":
		var gameState ServerGameState
		if err := json.Unmarshal(payloadBytes, &gameState); err != nil {
			log.Printf("Error unmarshaling gameStarted: %v", err)
		} else if nc.OnGameStarted != nil {
			log.Printf("Calling onGameStarted with %d players", gameState.NumActivePlayers)
			nc.OnGameStarted(&gameState)
		}

	case "gameStateUpdate":
		var gameState ServerGameState
		if err := json.Unmarshal(payloadBytes, &gameState); err != nil {
			log.Printf("Error unmarshaling gameStateUpdate: %v", err)
		} else if nc.OnGameStateUpdate != nil {
			nc.OnGameStateUpdate(&gameState)
		}

	case "error":
		var payload struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(payloadBytes, &payload); err == nil && nc.OnError != nil {
			nc.OnError(payload.Message)
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
