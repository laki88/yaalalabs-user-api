// user-ws/ws/subscriber.go
package ws

import (
	"log"

	nats "github.com/nats-io/nats.go"
)

func StartNATSListener(hub *Hub) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println("[WARN] NATS not available, continuing without external REST updates.")
		return
	}
	log.Println("[INFO] Connected to NATS for external REST updates")

	_, err = nc.Subscribe("users.updated", func(m *nats.Msg) {
		log.Println("NATS message received:", string(m.Data))
		hub.BroadcastToSubscribers("users", m.Data)
	})
	if err != nil {
		log.Println("[ERROR] Failed to subscribe to users.updated:", err)
	}
}
