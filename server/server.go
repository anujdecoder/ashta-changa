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
	roomManager RoomManagerInterface
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
	mux.HandleFunc(WebSocketEndpoint, func(w http.ResponseWriter, r *http.Request) {
		ServeWs(s.hub, s.roomManager, w, r)
	})

	// API endpoints
	mux.HandleFunc(APIRoomsEndpoint, s.handleRooms)
	mux.HandleFunc(APIRoomEndpointPrefix, s.handleRoom)

	// Health check
	mux.HandleFunc(HealthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// CORS middleware
	handler := withCORS(mux)

	addr := ":" + s.port
	log.Printf("Starting server on %s", addr)
	log.Printf("WebSocket endpoint: ws://localhost%s%s", addr, WebSocketEndpoint)
	log.Printf("API endpoint: http://localhost%s%s", addr, APIRoomsEndpoint)

	return http.ListenAndServe(addr, handler)
}

// handleRooms handles room list requests
func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", ContentTypeJSON)

	switch r.Method {
	case MethodGet:
		rooms := s.roomManager.ListRooms()
		json.NewEncoder(w).Encode(rooms)
	case MethodPost:
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
	w.Header().Set("Content-Type", ContentTypeJSON)

	// Extract room ID from URL path
	roomID := r.URL.Path[len(APIRoomEndpointPrefix):]

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case MethodGet:
		json.NewEncoder(w).Encode(room)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// withCORS adds CORS headers to responses
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CORSOriginHeader, CORSAllowAll)
		w.Header().Set(CORSMethodsHeader, CORSAllowedMethods)
		w.Header().Set(CORSHeadersHeader, CORSAllowedHeaders)

		if r.Method == MethodOptions {
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
