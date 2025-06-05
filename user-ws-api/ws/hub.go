// user-ws/ws/hub.go
package ws

import "github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"

type Hub struct {
	clients       map[*Client]bool
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
	subscriptions map[*Client]map[string]bool // client -> set of entities
	userService   userservice.UserService
}

func NewHub(userService userservice.UserService) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscriptions: make(map[*Client]map[string]bool),
		userService:   userService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.subscriptions[client] = make(map[string]bool)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.subscriptions, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					delete(h.subscriptions, client)
				}
			}
		}
	}
}

func (h *Hub) Subscribe(client *Client, entity string) {
	if _, ok := h.subscriptions[client]; ok {
		h.subscriptions[client][entity] = true
	}
}

func (h *Hub) BroadcastToSubscribers(entity string, message []byte) {
	for client := range h.clients {
		if h.subscriptions[client][entity] {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(h.subscriptions, client)
			}
		}
	}
}
