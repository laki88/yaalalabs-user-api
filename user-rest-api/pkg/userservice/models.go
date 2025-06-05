package userservice

import "github.com/google/uuid"

type User struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Age       *int32    `json:"age,omitempty"`
	Status    *string   `json:"status,omitempty"`
}

type CreateUserParams struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	Age       *int32  `json:"age,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type UpdateUserParams struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Age       *int32    `json:"age,omitempty"`
	Status    *string   `json:"status,omitempty"`
}
