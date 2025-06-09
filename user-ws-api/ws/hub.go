// user-ws/ws/hub.go
package ws

import "github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan BroadcastMessage
	register    chan *Client
	unregister  chan *Client
	userService userservice.UserService
}

func NewHub(userService userservice.UserService) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan BroadcastMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		userService: userService,
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
