// user-ws/ws/hub.go
package ws

import (
	"context"
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"user-ws-api/interfaces"
	"user-ws-api/models"
)

// MessageHandler is the interface each handler must implement
type MessageHandler interface {
	HandleMessage(c *Client, ctx context.Context, msg WSMessage)
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan BroadcastMessage
	register    chan *Client
	unregister  chan *Client
	userService userservice.UserService
	router      interfaces.OrderSubmitter
	// handlers registry: entity -> type -> handler
	handlers map[string]map[string]MessageHandler
}

func NewHub(userService userservice.UserService, router interfaces.OrderSubmitter) *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),

		userService: userService,
		router:      router,
	}

	h.registerHandlers()

	return h
}

func (h *Hub) SetTradeChannel(tradeCh <-chan models.Trade) {
	go func() {
		for trade := range tradeCh {
			h.SendTradeToUsers(trade)
		}
	}()
}

func (h *Hub) registerHandlers() {
	h.handlers = map[string]map[string]MessageHandler{
		"users": {
			"create":    &CreateUserHandler{service: h.userService},
			"update":    &UpdateUserHandler{service: h.userService},
			"delete":    &DeleteUserHandler{service: h.userService},
			"get":       &GetUsersHandler{service: h.userService},
			"get_by_id": &GetUserByIDHandler{service: h.userService},
		},
		"orders": {
			"order": &CreateOrderHandler{router: h.router},
		},
	}
}

func (h *Hub) SendTradeToUsers(trade models.Trade) {
	payload, _ := json.Marshal(trade)
	msg := WSMessage{
		Type:    "trade",
		Entity:  "orders",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)

	for client := range h.clients {
		if client.userID == trade.BuyerID || client.userID == trade.SellerID {
			client.send <- data
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case broadcastMessage := <-h.broadcast:
			for client := range h.clients {
				if client.subscribedEntities[broadcastMessage.Entity] {
					select {
					case client.send <- broadcastMessage.Message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
