// Package network contains the network client for multiplayer gameplay.
// This package handles WebSocket communication with the backend server.
package network

import (
	"encoding/json"
	"sync"
	"syscall/js"
	"time"
)

// NetworkClient handles WebSocket connection to the server
type NetworkClient struct {
	ws         js.Value
	serverURL  string
	playerID   string
	playerName string
	roomID     string
	connected  bool
	mu         sync.RWMutex

	// Callbacks
	OnRoomCreated     func(room *NetworkRoom, player *NetworkPlayer)
	OnRoomJoined      func(room *NetworkRoom, player *NetworkPlayer)
	OnPlayerJoined    func(player *NetworkPlayer)
	OnPlayerLeft      func(player *NetworkPlayer)
	OnGameStarted     func(gameState *ServerGameState)
	OnGameStateUpdate func(gameState *ServerGameState)
	OnError           func(msg string)
}

// NetworkRoom represents a game room from server
type NetworkRoom struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	HostID      string           `json:"hostId"`
	Players     []*NetworkPlayer `json:"players"`
	MaxPlayers  int              `json:"maxPlayers"`
	GameStarted bool             `json:"gameStarted"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// NetworkPlayer represents a player from server
type NetworkPlayer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RoomID    string `json:"roomId"`
	PlayerIdx int    `json:"playerIdx"`
	IsHost    bool   `json:"isHost"`
	IsReady   bool   `json:"isReady"`
}

// ServerGameState represents game state from server
type ServerGameState struct {
	CurrentPlayer    int            `json:"currentPlayer"`
	RollResult       int            `json:"rollResult"`
	HasRolled        bool           `json:"hasRolled"`
	ShellStates      []bool         `json:"shellStates"`
	ExtraTurn        bool           `json:"extraTurn"`
	Winner           int            `json:"winner"`
	PlayerTokens     []PlayerTokens `json:"playerTokens"`
	NumActivePlayers int            `json:"numActivePlayers"`
}

// ServerTokenState represents a single token's state from server
type ServerTokenState struct {
	ID       int `json:"id"`
	State    int `json:"state"`
	Position int `json:"position"`
}

// PlayerTokens represents a player's tokens state from server
type PlayerTokens struct {
	PlayerIdx int                `json:"playerIdx"`
	Tokens    []ServerTokenState `json:"tokens"`
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewNetworkClient creates a new network client
func NewNetworkClient(serverURL string) *NetworkClient {
	return &NetworkClient{
		serverURL: serverURL,
	}
}

// IsConnected returns connection status
func (nc *NetworkClient) IsConnected() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.connected
}

// GetPlayerID returns the current player ID
func (nc *NetworkClient) GetPlayerID() string {
	return nc.playerID
}

// GetRoomID returns the current room ID
func (nc *NetworkClient) GetRoomID() string {
	return nc.roomID
}

// GetPlayerName returns the current player name
func (nc *NetworkClient) GetPlayerName() string {
	return nc.playerName
}

// GetServerURL returns the WebSocket server URL based on current location
func GetServerURL() string {
	// Check if running in browser
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		// Not in browser, return default
		return "ws://localhost:8080/ws"
	}

	// Get current location
	location := js.Global().Get("window").Get("location")
	protocol := location.Get("protocol").String()
	host := location.Get("host").String()

	// Convert http/https to ws/wss
	wsProtocol := "ws"
	if protocol == "https:" {
		wsProtocol = "wss"
	}

	// If we're running on localhost (common for development),
	// assume the backend is on port 8080 as specified in README.
	if hasPrefix(host, "localhost") || hasPrefix(host, "127.0.0.1") {
		hostName := splitHost(host)
		return wsProtocol + "://" + hostName + ":8080/ws"
	}

	return wsProtocol + "://" + host + "/ws"
}

// GetAPIURL gets the REST API URL based on current location
func GetAPIURL() string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return "http://localhost:8080/api/rooms"
	}

	location := js.Global().Get("window").Get("location")
	protocol := location.Get("protocol").String()
	host := location.Get("host").String()

	// If we're running on localhost, assume the backend is on port 8080
	if hasPrefix(host, "localhost") || hasPrefix(host, "127.0.0.1") {
		hostName := splitHost(host)
		return "http://" + hostName + ":8080/api/rooms"
	}

	return protocol + "//" + host + "/api/rooms"
}

// GetRoomFromURL gets room ID from URL query parameters
func GetRoomFromURL() string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return ""
	}

	location := js.Global().Get("window").Get("location")
	search := location.Get("search").String()

	// Parse query string
	if hasPrefix(search, "?") {
		search = search[1:]
	}

	pairs := splitString(search, "&")
	for _, pair := range pairs {
		parts := splitString(pair, "=")
		if len(parts) == 2 && parts[0] == "room" {
			return parts[1]
		}
	}

	return ""
}

// GetRoomURL returns the full shareable room URL
func GetRoomURL(roomID string) string {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return "?room=" + roomID
	}

	location := js.Global().Get("window").Get("location")
	origin := location.Get("origin").String()
	pathname := location.Get("pathname").String()
	return origin + pathname + "?room=" + roomID
}

// CopyToClipboard copies text to clipboard using browser API
func CopyToClipboard(text string) bool {
	nav := js.Global().Get("navigator")
	if nav.IsUndefined() {
		return false
	}
	clipboard := nav.Get("clipboard")
	if clipboard.IsUndefined() {
		return false
	}
	clipboard.Call("writeText", text)
	return true
}

// Helper functions for string manipulation (avoiding strings package in WASM if needed)
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func splitHost(host string) string {
	for i := 0; i < len(host); i++ {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// UnmarshalPayload unmarshals payload bytes to a given type
func UnmarshalPayload(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalMessage marshals a message to JSON
func MarshalMessage(msg WSMessage) ([]byte, error) {
	return json.Marshal(msg)
}
