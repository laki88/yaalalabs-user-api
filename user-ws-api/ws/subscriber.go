// user-ws/ws/subscriber.go
package ws

import (
	"log/slog"

	nats "github.com/nats-io/nats.go"
)

func StartNATSListener(hub *Hub, natsURL string) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		slog.Error("[WARN] NATS not available, continuing without external REST updates.")
		return
	}
	slog.Info("[INFO] Connected to NATS for external REST updates")

	_, err = nc.Subscribe("users.updated", func(m *nats.Msg) {
		slog.Info("NATS message received:", "Message", string(m.Data))
		hub.broadcast <- BroadcastMessage{
			Entity:  "users",
			Message: m.Data,
		}
	})
	if err != nil {
		slog.Error("[ERROR] Failed to subscribe to users.updated:", "Error", err)
	}
}
