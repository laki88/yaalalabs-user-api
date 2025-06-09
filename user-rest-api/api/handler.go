package api

import (
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/nats"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
)

type Handler struct {
	Service userservice.UserService
}

var req struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone"`
	Age       *int32 `json:"age"`
	Status    string `json:"status"`
}

func NewHandler(service userservice.UserService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := internal.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the service layer with decoupled types
	user, err := h.Service.CreateUser(r.Context(), userservice.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     &req.Phone,
		Age:       req.Age,
		Status:    &req.Status,
	})

	if err != nil {
		http.Error(w, "Could not create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish to NATS
	event, _ := json.Marshal(user)
	nats.Publish("users.updated", event)

	// Respond with the created user
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := internal.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.Service.UpdateUser(r.Context(), userservice.UpdateUserParams{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     &req.Phone,
		Age:       req.Age,
		Status:    &req.Status,
	})

	if err != nil {
		http.Error(w, "Could not update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	event, _ := json.Marshal(user)
	nats.Publish("users.updated", event)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
