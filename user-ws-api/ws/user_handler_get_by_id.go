package ws

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log/slog"
	"user-ws-api/common"
)

type GetUserByIDHandler struct {
	service userservice.UserService
}

func (h *GetUserByIDHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	var payload struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		slog.Error("Invalid get_by_id payload:", err)
		errMsg := map[string]string{"error": "Invalid user get_by_id payload"}
		c.send <- common.MakeWSResponse("error", "users", "get_by_id", errMsg)
		return
	}
	user, err := h.service.GetUser(ctx, payload.UserID)
	if err != nil {
		slog.Error("GetUser error:", err)
		return
	}
	resp, _ := json.Marshal(user)
	c.send <- resp
}
