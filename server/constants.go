package server

import (
	"time"
)

// WebSocket timing constants
const (
	// WriteWait is the time allowed to write a message to the peer
	WriteWait = 10 * time.Second

	// PongWait is the time allowed to read the next pong message from the peer
	PongWait = 60 * time.Second

	// PingPeriod is the period for sending pings to peer (must be less than PongWait)
	PingPeriod = (PongWait * 9) / 10

	// MaxMessageSize is the maximum message size allowed from peer
	MaxMessageSize = 512
)

// Buffer sizes
const (
	// SendChannelSize is the buffer size for client send channels
	SendChannelSize = 256

	// ReadBufferSize is the WebSocket read buffer size
	ReadBufferSize = 1024

	// WriteBufferSize is the WebSocket write buffer size
	WriteBufferSize = 1024
)

// Game constants
const (
	// DefaultMaxPlayers is the default maximum number of players per room
	DefaultMaxPlayers = 4

	// RoomIDLength is the length of generated room IDs
	RoomIDLength = 6

	// PlayerIDTimestampFormat is the time format for player ID generation
	PlayerIDTimestampFormat = "20060102150405"

	// RandomStringLength is the length of the random suffix for player IDs
	RandomStringLength = 6
)

// Token states
const (
	// TokenStateAtStart represents a token at the starting position
	TokenStateAtStart = 0

	// TokenStateOnBoard represents a token on the board
	TokenStateOnBoard = 1

	// TokenStateFinished represents a token that has reached the center
	TokenStateFinished = 2
)

// Game actions
const (
	// ActionRoll represents a roll action
	ActionRoll = "roll"

	// ActionMove represents a move action
	ActionMove = "move"
)

// Message types
const (
	// MessageTypeRoomCreated is sent when a room is created
	MessageTypeRoomCreated = "roomCreated"

	// MessageTypeRoomJoined is sent when a player joins a room
	MessageTypeRoomJoined = "roomJoined"

	// MessageTypePlayerJoined is sent when a new player joins
	MessageTypePlayerJoined = "playerJoined"

	// MessageTypePlayerLeft is sent when a player leaves
	MessageTypePlayerLeft = "playerLeft"

	// MessageTypeGameStarted is sent when the game starts
	MessageTypeGameStarted = "gameStarted"

	// MessageTypeGameStateUpdate is sent when game state updates
	MessageTypeGameStateUpdate = "gameStateUpdate"

	// MessageTypePlayerReady is sent when a player's ready status changes
	MessageTypePlayerReady = "playerReady"

	// MessageTypeRoomsList is sent with the list of rooms
	MessageTypeRoomsList = "roomsList"

	// MessageTypeError is sent when an error occurs
	MessageTypeError = "error"
)

// Character sets for ID generation
const (
	// RoomIDCharset is the character set for room ID generation
	RoomIDCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	// PlayerIDCharset is the character set for player ID generation
	PlayerIDCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// HTTP endpoints
const (
	// WebSocketEndpoint is the WebSocket endpoint path
	WebSocketEndpoint = "/ws"

	// APIRoomsEndpoint is the rooms API endpoint path
	APIRoomsEndpoint = "/api/rooms"

	// APIRoomEndpointPrefix is the prefix for individual room API endpoints
	APIRoomEndpointPrefix = "/api/room/"

	// HealthEndpoint is the health check endpoint path
	HealthEndpoint = "/health"
)

// HTTP methods
const (
	// MethodGet represents GET method
	MethodGet = "GET"

	// MethodPost represents POST method
	MethodPost = "POST"

	// MethodOptions represents OPTIONS method
	MethodOptions = "OPTIONS"
)

// CORS headers
const (
	// CORSOriginHeader is the Access-Control-Allow-Origin header
	CORSOriginHeader = "Access-Control-Allow-Origin"

	// CORSMethodsHeader is the Access-Control-Allow-Methods header
	CORSMethodsHeader = "Access-Control-Allow-Methods"

	// CORSHeadersHeader is the Access-Control-Allow-Headers header
	CORSHeadersHeader = "Access-Control-Allow-Headers"

	// CORSAllowAll is the value to allow all origins
	CORSAllowAll = "*"

	// CORSAllowedMethods lists all allowed HTTP methods
	CORSAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS"

	// CORSAllowedHeaders lists all allowed headers
	CORSAllowedHeaders = "Content-Type, Authorization"
)

// Content types
const (
	// ContentTypeJSON is the application/json content type
	ContentTypeJSON = "application/json"
)
