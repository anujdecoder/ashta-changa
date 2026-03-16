package server

import (
	"net/http"
)

// RoomManagerInterface defines the interface for room management
type RoomManagerInterface interface {
	// CreateRoom creates a new game room with the given name and host ID
	CreateRoom(name, hostID string) *Room

	// GetRoom retrieves a room by its ID
	GetRoom(roomID string) (*Room, bool)

	// RemoveRoom removes a room by its ID
	RemoveRoom(roomID string)

	// ListRooms returns all active rooms
	ListRooms() []*Room
}

// RoomInterface defines the interface for room operations
type RoomInterface interface {
	// AddPlayer adds a player to the room
	AddPlayer(player *Player) bool

	// RemovePlayer removes a player from the room by ID
	RemovePlayer(playerID string)

	// GetPlayer retrieves a player by ID
	GetPlayer(playerID string) (*Player, bool)

	// StartGame starts the game if conditions are met
	StartGame() bool

	// Broadcast sends a message to all players in the room
	Broadcast(msg interface{})

	// BroadcastToOthers sends a message to all players except the sender
	BroadcastToOthers(senderID string, msg interface{})
}

// MessageHandlerInterface defines the interface for handling WebSocket messages
type MessageHandlerInterface interface {
	// HandleMessage processes an incoming WebSocket message
	HandleMessage(client *Client, data []byte, roomManager RoomManagerInterface)

	// HandleCreateRoom processes a create room request
	HandleCreateRoom(client *Client, payload interface{}, roomManager RoomManagerInterface)

	// HandleJoinRoom processes a join room request
	HandleJoinRoom(client *Client, payload interface{}, roomManager RoomManagerInterface)

	// HandleStartGame processes a start game request
	HandleStartGame(client *Client, payload interface{}, roomManager RoomManagerInterface)

	// HandleGameAction processes a game action (roll, move)
	HandleGameAction(client *Client, payload interface{}, roomManager RoomManagerInterface)

	// HandleGetRooms processes a get rooms request
	HandleGetRooms(client *Client, roomManager RoomManagerInterface)

	// HandleSetReady processes a player ready status update
	HandleSetReady(client *Client, payload interface{}, roomManager RoomManagerInterface)
}

// ClientInterface defines the interface for WebSocket client operations
type ClientInterface interface {
	// ReadPump pumps messages from the WebSocket connection
	ReadPump(roomManager RoomManagerInterface)

	// WritePump pumps messages to the WebSocket connection
	WritePump()

	// Send sends a message to the client's send channel
	Send(data []byte)

	// Close closes the client connection
	Close()
}

// HubInterface defines the interface for the WebSocket hub
type HubInterface interface {
	// Run starts the hub's event loop
	Run()

	// Register registers a new client
	Register(client *Client)

	// Unregister removes a client
	Unregister(client *Client)
}

// ServerInterface defines the interface for the HTTP server
type ServerInterface interface {
	// Start starts the server
	Start() error

	// HandleRooms handles room list requests
	HandleRooms(w http.ResponseWriter, r *http.Request)

	// HandleRoom handles individual room requests
	HandleRoom(w http.ResponseWriter, r *http.Request)
}

// IDGeneratorInterface defines the interface for ID generation
type IDGeneratorInterface interface {
	// GenerateRoomID generates a unique room ID
	GenerateRoomID() string

	// GeneratePlayerID generates a unique player ID
	GeneratePlayerID() string

	// RandomString generates a random string of given length
	RandomString(length int) string
}

// BroadcasterInterface defines the interface for broadcasting messages
type BroadcasterInterface interface {
	// Broadcast sends a message to all players
	Broadcast(msg interface{})

	// BroadcastToOthers sends a message to all players except the sender
	BroadcastToOthers(senderID string, msg interface{})
}
