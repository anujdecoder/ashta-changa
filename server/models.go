package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Room represents a game room where players can join and play together
type Room struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	HostID      string     `json:"hostId"`
	Players     []*Player  `json:"players"`
	MaxPlayers  int        `json:"maxPlayers"`
	GameStarted bool       `json:"gameStarted"`
	GameState   *GameState `json:"gameState"`
	CreatedAt   time.Time  `json:"createdAt"`

	mu sync.RWMutex
}

// Player represents a player in a room
type Player struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	RoomID    string  `json:"roomId"`
	PlayerIdx int     `json:"playerIdx"` // 0-3 for the 4 players in game
	IsHost    bool    `json:"isHost"`
	IsReady   bool    `json:"isReady"`
	Conn      *Client `json:"-"` // WebSocket connection
}

// GameState represents the current state of the game
type GameState struct {
	CurrentPlayer    int            `json:"currentPlayer"`
	RollResult       int            `json:"rollResult"`
	HasRolled        bool           `json:"hasRolled"`
	ShellStates      []bool         `json:"shellStates"`
	ExtraTurn        bool           `json:"extraTurn"`
	Winner           int            `json:"winner"`
	PlayerTokens     []PlayerTokens `json:"playerTokens"`
	NumActivePlayers int            `json:"numActivePlayers"`
}

// PlayerTokens represents a player's tokens state
type PlayerTokens struct {
	PlayerIdx int          `json:"playerIdx"`
	Tokens    []TokenState `json:"tokens"`
}

// TokenState represents the state of a single token
type TokenState struct {
	ID       int `json:"id"`
	State    int `json:"state"` // 0=atStart, 1=onBoard, 2=finished
	Position int `json:"position"`
}

// Client is a middleman between the WebSocket connection and the hub
type Client struct {
	hub *Hub

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Player info
	player *Player
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Room manager
	roomManager RoomManagerInterface
}

// RoomManager manages all game rooms
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

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

// ReadyRequest represents a player ready status request
type ReadyRequest struct {
	RoomID  string `json:"roomId"`
	IsReady bool   `json:"isReady"`
}

// ErrorPayload represents an error message payload
type ErrorPayload struct {
	Message string `json:"message"`
}
