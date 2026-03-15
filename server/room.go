package server

import (
	"encoding/json"
	"sync"
	"time"
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

// RoomManager manages all game rooms
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new game room
func (rm *RoomManager) CreateRoom(name, hostID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := generateRoomID()
	room := &Room{
		ID:         roomID,
		Name:       name,
		HostID:     hostID,
		Players:    make([]*Player, 0),
		MaxPlayers: 4,
		GameState: &GameState{
			CurrentPlayer:    0,
			RollResult:       0,
			HasRolled:        false,
			ShellStates:      make([]bool, 4),
			ExtraTurn:        false,
			Winner:           -1,
			PlayerTokens:     make([]PlayerTokens, 0),
			NumActivePlayers: 0,
		},
		CreatedAt: time.Now(),
	}

	rm.rooms[roomID] = room
	return room
}

// GetRoom gets a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	return room, exists
}

// RemoveRoom removes a room
func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.rooms, roomID)
}

// ListRooms returns all active rooms
func (rm *RoomManager) ListRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// AddPlayer adds a player to a room
func (r *Room) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.MaxPlayers || r.GameStarted {
		return false
	}

	player.RoomID = r.ID
	player.PlayerIdx = len(r.Players)

	// First player is host
	if len(r.Players) == 0 {
		player.IsHost = true
		r.HostID = player.ID
	}

	r.Players = append(r.Players, player)
	r.GameState.NumActivePlayers = len(r.Players)
	return true
}

// RemovePlayer removes a player from a room
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.Players {
		if p.ID == playerID {
			// Remove player from slice
			r.Players = append(r.Players[:i], r.Players[i+1:]...)

			// Reassign player indices
			for j, player := range r.Players {
				player.PlayerIdx = j
			}

			// If host left, assign new host
			if p.IsHost && len(r.Players) > 0 {
				r.Players[0].IsHost = true
				r.HostID = r.Players[0].ID
			}

			r.GameState.NumActivePlayers = len(r.Players)
			break
		}
	}
}

// GetPlayer gets a player by ID
func (r *Room) GetPlayer(playerID string) (*Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.Players {
		if p.ID == playerID {
			return p, true
		}
	}
	return nil, false
}

// StartGame starts the game
func (r *Room) StartGame() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) < 1 || r.GameStarted {
		return false
	}

	r.GameStarted = true
	r.GameState.NumActivePlayers = len(r.Players)

	// Reset game state fields
	r.GameState.CurrentPlayer = 0
	r.GameState.RollResult = 0
	r.GameState.HasRolled = false
	r.GameState.ExtraTurn = false
	r.GameState.Winner = -1
	r.GameState.PlayerTokens = make([]PlayerTokens, 0)
	r.GameState.ShellStates = make([]bool, 4)

	// Initialize player tokens - starting positions are center safe houses (not corners)
	// Player 0: position 2, Player 1: position 6, Player 2: position 10, Player 3: position 14
	for _, player := range r.Players {
		pt := PlayerTokens{
			PlayerIdx: player.PlayerIdx,
			Tokens:    make([]TokenState, 4),
		}
		startPos := player.PlayerIdx*4 + 2 // Center safe house positions: 2, 6, 10, 14
		for i := 0; i < 4; i++ {
			pt.Tokens[i] = TokenState{
				ID:       i,
				State:    0, // atStart
				Position: startPos,
			}
		}
		r.GameState.PlayerTokens = append(r.GameState.PlayerTokens, pt)
	}

	return true
}

// Broadcast sends a message to all players in the room
func (r *Room) Broadcast(msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, p := range r.Players {
		if p.Conn != nil {
			p.Conn.send <- data
		}
	}
}

// BroadcastToOthers sends a message to all players except the sender
func (r *Room) BroadcastToOthers(senderID string, msg interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, p := range r.Players {
		if p.Conn != nil && p.ID != senderID {
			p.Conn.send <- data
		}
	}
}

// generateRoomID generates a unique room ID
func generateRoomID() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
