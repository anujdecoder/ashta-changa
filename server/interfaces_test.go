package server

import (
	"testing"
)

// TestRoomManagerInterface verifies RoomManager implements RoomManagerInterface
func TestRoomManagerInterface(t *testing.T) {
	var _ RoomManagerInterface = (*RoomManager)(nil)
}

// TestRoomInterface verifies Room implements RoomInterface
func TestRoomInterface(t *testing.T) {
	var _ RoomInterface = (*Room)(nil)
}

// TestRoomManagerCreateRoom verifies room creation
func TestRoomManagerCreateRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("TestRoom", "host123")

	if room == nil {
		t.Fatal("CreateRoom should not return nil")
	}
	if room.Name != "TestRoom" {
		t.Errorf("Room name = %s, want TestRoom", room.Name)
	}
	if room.HostID != "host123" {
		t.Errorf("Host ID = %s, want host123", room.HostID)
	}
	if len(room.Players) != 0 {
		t.Errorf("New room should have 0 players, got %d", len(room.Players))
	}
}

// TestRoomManagerGetRoom verifies room retrieval
func TestRoomManagerGetRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	// Should find the room
	found, exists := rm.GetRoom(room.ID)
	if !exists {
		t.Error("Should find created room")
	}
	if found.ID != room.ID {
		t.Error("Found room should have same ID")
	}

	// Should not find non-existent room
	_, exists = rm.GetRoom("NONEXISTENT")
	if exists {
		t.Error("Should not find non-existent room")
	}
}

// TestRoomManagerListRooms verifies room listing
func TestRoomManagerListRooms(t *testing.T) {
	rm := NewRoomManager()

	// Empty initially
	rooms := rm.ListRooms()
	if len(rooms) != 0 {
		t.Errorf("Expected 0 rooms initially, got %d", len(rooms))
	}

	// Add rooms
	rm.CreateRoom("Room1", "host1")
	rm.CreateRoom("Room2", "host2")

	rooms = rm.ListRooms()
	if len(rooms) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(rooms))
	}
}

// TestRoomManagerRemoveRoom verifies room removal
func TestRoomManagerRemoveRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")
	roomID := room.ID

	// Remove the room
	rm.RemoveRoom(roomID)

	// Should not find it anymore
	_, exists := rm.GetRoom(roomID)
	if exists {
		t.Error("Room should be removed")
	}
}

// TestRoomAddPlayer verifies adding players to room
func TestRoomAddPlayer(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	player := &Player{ID: "p1", Name: "Player1"}
	if !room.AddPlayer(player) {
		t.Error("Should add first player")
	}

	if len(room.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(room.Players))
	}

	// Should not add more than MaxPlayers
	room.AddPlayer(&Player{ID: "p2", Name: "Player2"})
	room.AddPlayer(&Player{ID: "p3", Name: "Player3"})
	room.AddPlayer(&Player{ID: "p4", Name: "Player4"})

	if len(room.Players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(room.Players))
	}

	// Should not add 5th player
	if room.AddPlayer(&Player{ID: "p5", Name: "Player5"}) {
		t.Error("Should not add 5th player")
	}
}

// TestRoomRemovePlayer verifies removing players
func TestRoomRemovePlayer(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	room.AddPlayer(&Player{ID: "p1", Name: "Player1"})
	room.AddPlayer(&Player{ID: "p2", Name: "Player2"})

	room.RemovePlayer("p1")

	if len(room.Players) != 1 {
		t.Errorf("Expected 1 player after removal, got %d", len(room.Players))
	}

	// Remove non-existent player should not panic
	room.RemovePlayer("nonexistent")
}

// TestRoomGetPlayer verifies player retrieval
func TestRoomGetPlayer(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	player := &Player{ID: "p1", Name: "Player1"}
	room.AddPlayer(player)

	found, exists := room.GetPlayer("p1")
	if !exists {
		t.Error("Should find added player")
	}
	if found.Name != "Player1" {
		t.Error("Found player should have correct name")
	}

	_, exists = room.GetPlayer("nonexistent")
	if exists {
		t.Error("Should not find non-existent player")
	}
}

// TestRoomStartGame verifies game starting
func TestRoomStartGame(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	// Cannot start with 0 players
	if room.StartGame() {
		t.Error("Should not start with 0 players")
	}

	room.AddPlayer(&Player{ID: "p1", Name: "Player1"})

	if !room.StartGame() {
		t.Error("Should start with 1 player")
	}

	if !room.GameStarted {
		t.Error("Game should be marked as started")
	}

	// Cannot start again
	if room.StartGame() {
		t.Error("Should not start already started game")
	}
}

// TestRoomBroadcast verifies broadcasting messages
func TestRoomBroadcast(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	// Just verify it doesn't panic
	room.Broadcast(Message{Type: "test", Payload: "data"})
}

// TestRoomBroadcastToOthers verifies broadcasting to others
func TestRoomBroadcastToOthers(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("Test", "host")

	// Just verify it doesn't panic
	room.BroadcastToOthers("p1", Message{Type: "test", Payload: "data"})
}

// TestConstants verifies all constants are defined
func TestConstants(t *testing.T) {
	// WebSocket constants
	if WriteWait == 0 {
		t.Error("WriteWait should be defined")
	}
	if PongWait == 0 {
		t.Error("PongWait should be defined")
	}
	if PingPeriod == 0 {
		t.Error("PingPeriod should be defined")
	}
	if MaxMessageSize == 0 {
		t.Error("MaxMessageSize should be defined")
	}

	// Buffer sizes
	if SendChannelSize == 0 {
		t.Error("SendChannelSize should be defined")
	}
	if ReadBufferSize == 0 {
		t.Error("ReadBufferSize should be defined")
	}
	if WriteBufferSize == 0 {
		t.Error("WriteBufferSize should be defined")
	}

	// Game constants
	if DefaultMaxPlayers == 0 {
		t.Error("DefaultMaxPlayers should be defined")
	}
	if RoomIDLength == 0 {
		t.Error("RoomIDLength should be defined")
	}

	// Token states
	if TokenStateAtStart != 0 {
		t.Error("TokenStateAtStart should be 0")
	}
	if TokenStateOnBoard != 1 {
		t.Error("TokenStateOnBoard should be 1")
	}
	if TokenStateFinished != 2 {
		t.Error("TokenStateFinished should be 2")
	}

	// Message types
	if MessageTypeRoomCreated == "" {
		t.Error("MessageTypeRoomCreated should be defined")
	}
	if MessageTypeRoomJoined == "" {
		t.Error("MessageTypeRoomJoined should be defined")
	}
	if MessageTypeGameStarted == "" {
		t.Error("MessageTypeGameStarted should be defined")
	}

	// Actions
	if ActionRoll == "" {
		t.Error("ActionRoll should be defined")
	}
	if ActionMove == "" {
		t.Error("ActionMove should be defined")
	}
}

// TestGenerateRoomID verifies room ID generation
func TestGenerateRoomID(t *testing.T) {
	id1 := generateRoomID()
	id2 := generateRoomID()

	if len(id1) != RoomIDLength {
		t.Errorf("Room ID length = %d, want %d", len(id1), RoomIDLength)
	}

	if id1 == id2 {
		t.Error("Generated room IDs should be unique")
	}
}

// TestGeneratePlayerID verifies player ID generation
func TestGeneratePlayerID(t *testing.T) {
	id1 := generatePlayerID()
	id2 := generatePlayerID()

	if len(id1) == 0 {
		t.Error("Player ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated player IDs should be unique")
	}
}
