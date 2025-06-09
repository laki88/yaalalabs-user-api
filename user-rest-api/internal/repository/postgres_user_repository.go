// internal/repository/postgres_user_repository.go
package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
)

type PostgresUserRepository struct {
	q *db.Queries
}

func NewPostgresUserRepository(q *db.Queries) *PostgresUserRepository {
	return &PostgresUserRepository{q: q}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, arg)
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteUser(ctx, userID)
}

func (r *PostgresUserRepository) GetUser(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return r.q.GetUser(ctx, userID)
}

func (r *PostgresUserRepository) ListUsers(ctx context.Context) ([]db.User, error) {
	return r.q.ListUsers(ctx)
}
