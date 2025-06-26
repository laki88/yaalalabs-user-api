package ws

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log/slog"
	"user-ws-api/common"
)

type DeleteUserHandler struct {
	service userservice.UserService
}

func (h *DeleteUserHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	var payload struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		slog.Error("Invalid delete payload:", err)
		errMsg := map[string]string{"error": "Invalid user payload"}
		c.send <- common.MakeWSResponse("error", "users", "delete", errMsg)
		return
	}
	if err := h.service.DeleteUser(ctx, payload.UserID); err != nil {
		slog.Error("Delete error:", err)
		errMsg := map[string]string{"error": err.Error()}
		c.send <- common.MakeWSResponse("error", "users", "delete", errMsg)
		return
	}
	msgResp := []byte(`{"status":"deleted"}`)
	c.send <- msgResp
	c.hub.broadcast <- BroadcastMessage{"users", msgResp}
}
