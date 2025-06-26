package common

import "encoding/json"

type WSResponse struct {
	Status string          `json:"status"` // "ok" or "error"
	Entity string          `json:"entity"`
	Type   string          `json:"type"` // e.g. "create", "update", etc.
	Data   json.RawMessage `json:"data"` // flexible payload
}

func MakeWSResponse(status, entity, msgType string, payload any) []byte {
	data, _ := json.Marshal(payload)
	resp := WSResponse{
		Status: status,
		Entity: entity,
		Type:   msgType,
		Data:   data,
	}
	raw, _ := json.Marshal(resp)
	return raw
}
