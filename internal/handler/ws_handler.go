package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket" // Note: need to add this to go.mod if not present
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For demo, allow all
	},
}

// Client represents a connected user.
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastSeatUpdate(showtimeID, seatID, status string) {
	msg := map[string]string{
		"type":        "seat_update",
		"showtime_id": showtimeID,
		"seat_id":     seatID,
		"status":      status,
	}
	payload, _ := json.Marshal(msg)
	h.broadcast <- payload
}

func (h *Client) ReadPump() {
	defer func() {
		h.Hub.unregister <- h
		h.Conn.Close()
	}()
	for {
		_, _, err := h.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Currently we only push from server to client
	}
}

func (h *Client) WritePump() {
	for {
		select {
		case message, ok := <-h.Send:
			if !ok {
				h.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			h.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading web socket: %v", err)
		return
	}
	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
