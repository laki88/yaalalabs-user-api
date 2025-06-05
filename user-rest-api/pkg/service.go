package pkg

import (
	"context"
	"github.com/google/uuid"

	"user-rest-api/internal/db"
)

type UserService interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (db.User, error)
}

type service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) UserService {
	return &service{queries: queries}
}

func (s *service) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return s.queries.CreateUser(ctx, arg)
}

func (s *service) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return s.queries.UpdateUser(ctx, arg)
}

func (s *service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.queries.DeleteUser(ctx, userID)
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return s.queries.GetUser(ctx, userID)
}
