package userservice_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/repository/mocks"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := userservice.NewService(repo)

	phone := "123"
	status := "active"
	age := int32(25)

	arg := userservice.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     &phone,
		Age:       &age,
		Status:    &status,
	}

	expectedUser := db.User{
		UserID:    uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     internal.ToNullString(phone),
		Age:       internal.ToNullInt32(&age),
		Status:    internal.ToNullString(status),
	}

	repo.On("CreateUser", mock.Anything, mock.AnythingOfType("db.CreateUserParams")).
		Return(expectedUser, nil)

	user, err := svc.CreateUser(context.Background(), arg)

	assert.NoError(t, err)
	assert.Equal(t, "John", user.FirstName)
	repo.AssertExpectations(t)
}

func TestUpdateUser(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := userservice.NewService(repo)

	phone := "456"
	status := "inactive"
	age := int32(30)
	id := uuid.New()

	arg := userservice.UpdateUserParams{
		UserID:    id,
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Phone:     &phone,
		Age:       &age,
		Status:    &status,
	}

	expected := db.User{
		UserID:    id,
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Phone:     internal.ToNullString(phone),
		Age:       internal.ToNullInt32(&age),
		Status:    internal.ToNullString(status),
	}

	repo.On("UpdateUser", mock.Anything, mock.AnythingOfType("db.UpdateUserParams")).
		Return(expected, nil)

	user, err := svc.UpdateUser(context.Background(), arg)

	assert.NoError(t, err)
	assert.Equal(t, "Jane", user.FirstName)
	repo.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := userservice.NewService(repo)

	id := uuid.New()
	repo.On("DeleteUser", mock.Anything, id).Return(nil)

	err := svc.DeleteUser(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := userservice.NewService(repo)

	id := uuid.New()
	expected := db.User{
		UserID:    id,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
	}

	repo.On("GetUser", mock.Anything, id).Return(expected, nil)

	user, err := svc.GetUser(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, "Test", user.FirstName)
	repo.AssertExpectations(t)
}

func TestGetAllUsers(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := userservice.NewService(repo)

	expected := []db.User{
		{UserID: uuid.New(), FirstName: "John"},
		{UserID: uuid.New(), FirstName: "Jane"},
	}

	repo.On("ListUsers", mock.Anything).Return(expected, nil)

	users, err := svc.GetAllUsers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	repo.AssertExpectations(t)
}
