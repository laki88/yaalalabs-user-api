package api

import (
	"database/sql"
	"encoding/json"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/nats"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
)

type Handler struct {
	Queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{Queries: queries}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirstName string `json:"first_name" validate:"required,min=2,max=50"`
		LastName  string `json:"last_name" validate:"required,min=2,max=50"`
		Email     string `json:"email" validate:"required,email"`
		Phone     string `json:"phone"`
		Age       *int32 `json:"age"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := internal.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert optional fields
	arg := db.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     internal.ToNullString(req.Phone),
		Age:       internal.ToNullInt32(req.Age),
		Status:    internal.ToNullString(req.Status),
	}

	user, err := h.Queries.CreateUser(r.Context(), arg)
	if err != nil {
		http.Error(w, "Could not create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	event, _ := json.Marshal(user)
	nats.Publish("users.updated", event)

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Queries.GetUser(r.Context(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Queries.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var req struct {
		FirstName string `json:"first_name" validate:"required,min=2,max=50"`
		LastName  string `json:"last_name" validate:"required,min=2,max=50"`
		Email     string `json:"email" validate:"required,email"`
		Phone     string `json:"phone"`
		Age       *int32 `json:"age"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := internal.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	arg := db.UpdateUserParams{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     internal.ToNullString(req.Phone),
		Age:       internal.ToNullInt32(req.Age),
		Status:    internal.ToNullString(req.Status),
	}

	user, err := h.Queries.UpdateUser(r.Context(), arg)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Could not update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	event, _ := json.Marshal(user)
	nats.Publish("users.updated", event)

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.Queries.DeleteUser(r.Context(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
