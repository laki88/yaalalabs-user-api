package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"user-ws-api/common"
	"user-ws-api/interfaces"

	"user-ws-api/models"
)

type CreateOrderHandler struct {
	router interfaces.OrderSubmitter
}

func (h *CreateOrderHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	var order models.Order
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		slog.Error("Invalid order payload:", err)
		errMsg := map[string]string{"error": "Invalid order payload"}
		c.send <- common.MakeWSResponse("error", "orders", "create", errMsg)
		return
	}
	h.router.Submit(order)
}
