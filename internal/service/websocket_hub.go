// internal/service/websocket_hub.go
package service

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan []byte
	Register   chan *WebSocketClient
	Unregister chan *WebSocketClient
}

type WebSocketClient struct {
	Hub  *WebSocketHub
	Conn *websocket.Conn
	Send chan []byte
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Println("WebSocket client connected")

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Println("WebSocket client disconnected")
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message: // Queue message for sending
					// Message queued successfully
				default:
					// Client's send channel is full, close it
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WebSocketHub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// We're not processing incoming messages from clients in this implementation
	}
}

func (c *WebSocketClient) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}
}

func (h *WebSocketHub) GetClientCount() int {
	return len(h.clients)
}
