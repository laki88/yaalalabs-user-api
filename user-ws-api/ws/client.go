// user-ws/ws/client.go
package ws

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"time"
	"user-ws-api/common"
)

type WSMessage struct {
	Type    string          `json:"type"`
	Entity  string          `json:"entity"`
	Payload json.RawMessage `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	hub                *Hub
	conn               *websocket.Conn
	send               chan []byte
	subscribedEntities map[string]bool
	userID             string
}

type BroadcastMessage struct {
	Entity  string
	Message []byte
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade error:", err)
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		slog.Error("Missing user_id in web socket connection")
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}
	client := &Client{
		hub:                hub,
		conn:               conn,
		send:               make(chan []byte, 256),
		subscribedEntities: make(map[string]bool),
		userID:             userID,
	}
	hub.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()
	c.conn.SetReadLimit(512 * 1024) // Max message size (512KB)

	if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return
	}
	// Initial handshake timeout
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				slog.Error("read error:", err)
			}
			break
		}
		// Reset timeout after each successful read
		if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return
		}
		slog.Info("Received message from client: ", "message", message)

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("Invalid JSON:", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		if !(msg.Entity == "users" || msg.Entity == "orders") {
			handleErrorMessage(cancel, "Unsupported entity:", msg, c)
			continue
		}

		// Dispatch based on entity and type using the Hub's handler registry
		entityHandlers, ok := c.hub.handlers[msg.Entity]
		if !ok {
			handleErrorMessage(cancel, "Unsupported entity:", msg, c)
			continue
		}
		handler, ok := entityHandlers[msg.Type]
		if !ok {
			handleErrorMessage(cancel, "Unsupported type:", msg, c)
			continue
		}
		handler.HandleMessage(c, ctx, msg)
		cancel()
	}
}

func handleErrorMessage(cancel context.CancelFunc, errorMessage string, msg WSMessage, c *Client) {
	slog.Error(errorMessage, msg.Entity)
	c.send <- common.MakeWSResponse("error", msg.Entity, msg.Type, map[string]string{"error": "Unsupported entity"})
	cancel()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // Ping timeout
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
