package ws

import (
	"context"
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log/slog"
	"user-ws-api/common"
)

type GetUsersHandler struct {
	service userservice.UserService
}

func (h *GetUsersHandler) HandleMessage(c *Client, ctx context.Context, msg WSMessage) {
	users, err := h.service.GetAllUsers(ctx)
	if err != nil {
		slog.Error("GetAllUsers error:", err)
		errMsg := map[string]string{"error": err.Error()}
		c.send <- common.MakeWSResponse("error", "users", "get", errMsg)
		return
	}
	resp, _ := json.Marshal(users)
	c.send <- resp
}
