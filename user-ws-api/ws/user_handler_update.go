package ws

import (
	"context"
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log/slog"
	"user-ws-api/common"
)

type UpdateUserHandler struct {
	service userservice.UserService
}

func (h *UpdateUserHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	var user userservice.UpdateUserParams
	if err := json.Unmarshal(msg.Payload, &user); err != nil {
		slog.Error("Invalid update payload:", err)
		errMsg := map[string]string{"error": "Invalid update payload"}
		c.send <- common.MakeWSResponse("error", "users", "update", errMsg)
		return
	}
	updated, err := h.service.UpdateUser(ctx, user)
	if err != nil {
		slog.Error("Update error:", err)
		errMsg := map[string]string{"error": err.Error()}
		c.send <- common.MakeWSResponse("error", "users", "update", errMsg)
		return
	}
	resp, _ := json.Marshal(updated)
	c.send <- resp
	c.hub.broadcast <- BroadcastMessage{"users", resp}
}
