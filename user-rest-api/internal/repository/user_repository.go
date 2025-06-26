//go:generate mockery --name UserRepository --structname MockUserRepository --output ./mocks --case underscore
package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
}
