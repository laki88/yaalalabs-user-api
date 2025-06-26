package api_test

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/api"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) CreateUser(ctx context.Context, arg userservice.CreateUserParams) (userservice.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(userservice.User), args.Error(1)
}

func (m *mockUserService) GetUser(ctx context.Context, id uuid.UUID) (userservice.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(userservice.User), args.Error(1)
}

func (m *mockUserService) GetAllUsers(ctx context.Context) ([]userservice.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]userservice.User), args.Error(1)
}

func (m *mockUserService) UpdateUser(ctx context.Context, arg userservice.UpdateUserParams) (userservice.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(userservice.User), args.Error(1)
}

func (m *mockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	handler := api.NewHandler(mockService)
	internal.InitValidator()

	reqBody := `{"first_name":"Alice","last_name":"Smith","email":"alice@example.com", "phone": "1234567890", "age": 30, "status": "Active"}`
	r := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(reqBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	phoneNumber := "1234567890"
	age := int32(30)
	status := "Active"
	expected := userservice.User{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com", Phone: &phoneNumber, Age: &age, Status: &status}
	mockService.On("CreateUser", mock.Anything, mock.Anything).Return(expected, nil)

	handler.CreateUser(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	handler := api.NewHandler(mockService)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()

	expected := userservice.User{UserID: userID, Email: "bob@example.com"}
	mockService.On("GetUser", mock.Anything, userID).Return(expected, nil)

	handler.GetUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	handler := api.NewHandler(mockService)
	internal.InitValidator()

	userID := uuid.New()
	body := `{"first_name":"Updated","last_name":"User","email":"updated@example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBufferString(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	expected := userservice.User{UserID: userID, Email: "updated@example.com"}
	mockService.On("UpdateUser", mock.Anything, mock.Anything).Return(expected, nil)

	handler.UpdateUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	handler := api.NewHandler(mockService)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	w := httptest.NewRecorder()

	mockService.On("DeleteUser", mock.Anything, userID).Return(nil)

	handler.DeleteUser(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
