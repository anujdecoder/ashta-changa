package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server represents the game server
type Server struct {
	hub         *Hub
	roomManager *RoomManager
	port        string
}

// NewServer creates a new game server
func NewServer(port string) *Server {
	roomManager := NewRoomManager()
	hub := NewHub(roomManager)

	return &Server{
		hub:         hub,
		roomManager: roomManager,
		port:        port,
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Start the hub
	go s.hub.Run()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(s.hub, s.roomManager, w, r)
	})

	// API endpoints
	mux.HandleFunc("/api/rooms", s.handleRooms)
	mux.HandleFunc("/api/room/", s.handleRoom)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// CORS middleware
	handler := withCORS(mux)

	addr := ":" + s.port
	log.Printf("Starting server on %s", addr)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", addr)
	log.Printf("API endpoint: http://localhost%s/api/rooms", addr)

	return http.ListenAndServe(addr, handler)
}

// handleRooms handles room list requests
func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rooms := s.roomManager.ListRooms()
		json.NewEncoder(w).Encode(rooms)
	case http.MethodPost:
		var req struct {
			RoomName   string `json:"roomName"`
			PlayerName string `json:"playerName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		playerID := generatePlayerID()
		room := s.roomManager.CreateRoom(req.RoomName, playerID)

		response := map[string]interface{}{
			"room":     room,
			"playerId": playerID,
		}
		json.NewEncoder(w).Encode(response)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRoom handles individual room requests
func (s *Server) handleRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract room ID from URL path
	roomID := r.URL.Path[len("/api/room/"):]

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(room)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// withCORS adds CORS headers to responses
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetRoomLink generates a shareable room link
func GetRoomLink(roomID string, baseURL string) string {
	return fmt.Sprintf("%s?room=%s", baseURL, roomID)
}
