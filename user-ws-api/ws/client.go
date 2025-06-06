// user-ws/ws/client.go
package ws

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log"
	"net/http"
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
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	hub.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		log.Printf("Received message from client: %s\n", message)

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		ctx := context.Background()

		if msg.Entity != "users" {
			log.Println("Unsupported entity:", msg.Entity)
			continue
		}

		switch msg.Type {
		case "subscribe":
			c.hub.Subscribe(c, msg.Entity)
			log.Printf("Client subscribed to %s", msg.Entity)
			ack := map[string]string{
				"type":   "subscribe_ack",
				"entity": msg.Entity,
				"status": "subscribed",
			}
			ackBytes, _ := json.Marshal(ack)
			c.send <- ackBytes
			continue

		case "create":
			var user userservice.CreateUserParams
			if err := json.Unmarshal(msg.Payload, &user); err != nil {
				log.Println("Invalid CreateUser payload:", err)
				continue
			}
			createdUser, err := c.hub.userService.CreateUser(ctx, user)
			if err != nil {
				log.Println("CreateUser error:", err)
				continue
			}
			resp, _ := json.Marshal(createdUser)
			c.send <- resp
			c.hub.BroadcastToSubscribers("users", resp)

		case "update":
			var update userservice.UpdateUserParams
			if err := json.Unmarshal(msg.Payload, &update); err != nil {
				log.Println("Invalid UpdateUser payload:", err)
				continue
			}
			updatedUser, err := c.hub.userService.UpdateUser(ctx, update)
			if err != nil {
				log.Println("UpdateUser error:", err)
				continue
			}
			resp, _ := json.Marshal(updatedUser)
			c.send <- resp
			c.hub.BroadcastToSubscribers("users", resp)

		case "delete":
			var payload struct {
				UserID uuid.UUID `json:"user_id"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Println("Invalid DeleteUser payload:", err)
				continue
			}
			if err := c.hub.userService.DeleteUser(ctx, payload.UserID); err != nil {
				log.Println("DeleteUser error:", err)
				continue
			}
			c.send <- []byte(`{"status":"deleted"}`)
			//c.hub.BroadcastToSubscribers("users", resp) // create response to send

		case "get":
			users, err := c.hub.userService.GetAllUsers(ctx)
			if err != nil {
				log.Println("GetAllUsers error:", err)
				continue
			}
			resp, _ := json.Marshal(users)
			c.send <- resp

		case "get_by_id":
			var payload struct {
				UserID uuid.UUID `json:"user_id"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Println("Invalid GetUserByID payload:", err)
				continue
			}
			user, err := c.hub.userService.GetUser(ctx, payload.UserID)
			if err != nil {
				log.Println("GetUser error:", err)
				continue
			}
			resp, _ := json.Marshal(user)
			c.send <- resp

		default:
			log.Println("Unsupported operation type:", msg.Type)
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
