package ws

import (
	"context"
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log/slog"
	"user-ws-api/common"
)

type CreateUserHandler struct {
	service userservice.UserService
}

func (h *CreateUserHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	var user userservice.CreateUserParams
	if err := json.Unmarshal(msg.Payload, &user); err != nil {
		slog.Error("Invalid create payload:", err)
		errMsg := map[string]string{"error": "Invalid user payload"}
		c.send <- common.MakeWSResponse("error", "users", "create", errMsg)
		return
	}
	created, err := h.service.CreateUser(ctx, user)
	if err != nil {
		slog.Error("Create error:", err)
		errMsg := map[string]string{"error": err.Error()}
		c.send <- common.MakeWSResponse("error", "users", "create", errMsg)
		return
	}
	resp, _ := json.Marshal(created)
	c.send <- resp
	c.hub.broadcast <- BroadcastMessage{"users", resp}
}
