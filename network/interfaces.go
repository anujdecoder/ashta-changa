// Package network contains the network client for multiplayer gameplay.
package network

// NetworkClientInterface defines the interface for network operations.
// This interface allows for mocking in tests and alternative implementations.
type NetworkClientInterface interface {
	// Connect connects to the WebSocket server
	Connect() error

	// Disconnect closes the connection
	Disconnect()

	// IsConnected returns connection status
	IsConnected() bool

	// GetPlayerID returns the current player ID
	GetPlayerID() string

	// GetRoomID returns the current room ID
	GetRoomID() string

	// GetPlayerName returns the current player name
	GetPlayerName() string

	// CreateRoom creates a new room
	CreateRoom(roomName, playerName string)

	// JoinRoom joins an existing room
	JoinRoom(roomID, playerName string, playerID string)

	// StartGame starts the game (host only)
	StartGame()

	// SendGameAction sends a game action to the server
	SendGameAction(action string, tokenIdx int)

	// SendRollAction sends a roll action request to the server
	SendRollAction()

	// SetReady sets player ready status
	SetReady(isReady bool)

	// SetOnRoomCreated sets the callback for room created event
	SetOnRoomCreated(callback func(room *NetworkRoom, player *NetworkPlayer))

	// SetOnRoomJoined sets the callback for room joined event
	SetOnRoomJoined(callback func(room *NetworkRoom, player *NetworkPlayer))

	// SetOnPlayerJoined sets the callback for player joined event
	SetOnPlayerJoined(callback func(player *NetworkPlayer))

	// SetOnPlayerLeft sets the callback for player left event
	SetOnPlayerLeft(callback func(player *NetworkPlayer))

	// SetOnGameStarted sets the callback for game started event
	SetOnGameStarted(callback func(gameState *ServerGameState))

	// SetOnGameStateUpdate sets the callback for game state update event
	SetOnGameStateUpdate(callback func(gameState *ServerGameState))

	// SetOnError sets the callback for error events
	SetOnError(callback func(msg string))
}

// Ensure NetworkClient implements NetworkClientInterface
var _ NetworkClientInterface = (*NetworkClient)(nil)

// SetOnRoomCreated sets the callback for room created event
func (nc *NetworkClient) SetOnRoomCreated(callback func(room *NetworkRoom, player *NetworkPlayer)) {
	nc.OnRoomCreated = callback
}

// SetOnRoomJoined sets the callback for room joined event
func (nc *NetworkClient) SetOnRoomJoined(callback func(room *NetworkRoom, player *NetworkPlayer)) {
	nc.OnRoomJoined = callback
}

// SetOnPlayerJoined sets the callback for player joined event
func (nc *NetworkClient) SetOnPlayerJoined(callback func(player *NetworkPlayer)) {
	nc.OnPlayerJoined = callback
}

// SetOnPlayerLeft sets the callback for player left event
func (nc *NetworkClient) SetOnPlayerLeft(callback func(player *NetworkPlayer)) {
	nc.OnPlayerLeft = callback
}

// SetOnGameStarted sets the callback for game started event
func (nc *NetworkClient) SetOnGameStarted(callback func(gameState *ServerGameState)) {
	nc.OnGameStarted = callback
}

// SetOnGameStateUpdate sets the callback for game state update event
func (nc *NetworkClient) SetOnGameStateUpdate(callback func(gameState *ServerGameState)) {
	nc.OnGameStateUpdate = callback
}

// SetOnError sets the callback for error events
func (nc *NetworkClient) SetOnError(callback func(msg string)) {
	nc.OnError = callback
}
