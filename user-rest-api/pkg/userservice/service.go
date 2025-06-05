package userservice

import (
	"context"
	"database/sql"
	"github.com/google/uuid"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
)

type UserService interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (User, error)
}

type service struct {
	q *db.Queries
}

func NewService(sqlDB *sql.DB) UserService {
	queries := db.New(sqlDB)
	return &service{q: queries}
}

func (s *service) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	dbArg := db.CreateUserParams{
		FirstName: arg.FirstName,
		LastName:  arg.LastName,
		Email:     arg.Email,
		Phone:     internal.ToNullString(*arg.Phone),
		Age:       internal.ToNullInt32(arg.Age),
		Status:    internal.ToNullString(*arg.Status),
	}
	user, err := s.q.CreateUser(ctx, dbArg)
	if err != nil {
		return User{}, err
	}
	return toPublicUser(user), nil
}

func (s *service) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	dbArg := db.UpdateUserParams{
		UserID:    arg.UserID,
		FirstName: arg.FirstName,
		LastName:  arg.LastName,
		Email:     arg.Email,
		Phone:     internal.ToNullString(*arg.Phone),
		Age:       internal.ToNullInt32(arg.Age),
		Status:    internal.ToNullString(*arg.Status),
	}
	user, err := s.q.UpdateUser(ctx, dbArg)
	if err != nil {
		return User{}, err
	}
	return toPublicUser(user), nil
}

func (s *service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.q.DeleteUser(ctx, userID)
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := s.q.GetUser(ctx, userID)
	if err != nil {
		return User{}, err
	}
	return toPublicUser(user), nil
}
func toPublicUser(u db.User) User {
	var phone *string
	if u.Phone.Valid {
		phone = &u.Phone.String
	}
	var age *int32
	if u.Age.Valid {
		age = &u.Age.Int32
	}
	var status *string
	if u.Status.Valid {
		status = &u.Status.String
	}
	return User{
		UserID:    u.UserID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     phone,
		Age:       age,
		Status:    status,
	}
}
