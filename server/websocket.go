package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Ensure Hub implements HubInterface
var _ HubInterface = (*Hub)(nil)

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  ReadBufferSize,
	WriteBufferSize: WriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// NewHub creates a new hub
func NewHub(roomManager RoomManagerInterface) *Hub {
	return &Hub{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		roomManager: roomManager,
	}
}

// Run runs the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove player from room
				if client.player != nil {
					if room, exists := h.roomManager.GetRoom(client.player.RoomID); exists {
						room.RemovePlayer(client.player.ID)

						// Notify other players
						room.Broadcast(Message{
							Type:    MessageTypePlayerLeft,
							Payload: client.player,
						})

						// Remove room if empty
						if len(room.Players) == 0 {
							h.roomManager.RemoveRoom(room.ID)
						}
					}
				}
			}
		}
	}
}

// ServeWs handles WebSocket requests from the peer
func ServeWs(hub *Hub, roomManager RoomManagerInterface, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, SendChannelSize),
	}

	hub.Register(client)

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump(roomManager)
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump(roomManager RoomManagerInterface) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle incoming message
		handleMessage(c, message, roomManager)
	}
}

// ReadPump is the public interface for readPump
func (c *Client) ReadPump(roomManager RoomManagerInterface) {
	c.readPump(roomManager)
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WritePump is the public interface for writePump
func (c *Client) WritePump() {
	c.writePump()
}

// Send sends data to the client's send channel
func (c *Client) Send(data []byte) {
	c.send <- data
}

// Close closes the client connection
func (c *Client) Close() {
	c.conn.Close()
}

// Register registers a client with the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
