package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of service.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(req *user.CreateUserRequest) (*user.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Login(req *user.LoginRequest) (*user.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetProfile(userID uuid.UUID) (*user.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(userID uuid.UUID, req *user.UpdateUserRequest) (*user.User, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) DeleteAccount(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserService) RefreshToken(token string) (*user.LoginResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *MockUserService) ForgotPassword(req *user.ForgotPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(req *user.ResetPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== Register Tests ====================

func TestUserHandler_Register_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/register", handler.Register)

	expectedUser := &user.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockService.On("Register", mock.AnythingOfType("*user.CreateUserRequest")).Return(expectedUser, nil)

	reqBody := `{"email":"test@example.com","password":"password123","first_name":"John","last_name":"Doe"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_InvalidRequest(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/register", handler.Register)

	reqBody := `{"email":"invalid"}` // Missing required fields
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Register_DuplicateEmail(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/register", handler.Register)

	mockService.On("Register", mock.AnythingOfType("*user.CreateUserRequest")).
		Return(nil, assert.AnError)

	reqBody := `{"email":"existing@example.com","password":"password123","first_name":"John","last_name":"Doe"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== Login Tests ====================

func TestUserHandler_Login_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/login", handler.Login)

	expectedResponse := &user.LoginResponse{
		Token:        "jwt-token",
		RefreshToken: "refresh-token",
		User:         &user.User{Email: "test@example.com"},
	}

	mockService.On("Login", mock.AnythingOfType("*user.LoginRequest")).Return(expectedResponse, nil)

	reqBody := `{"email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response user.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/login", handler.Login)

	mockService.On("Login", mock.AnythingOfType("*user.LoginRequest")).
		Return(nil, assert.AnError)

	reqBody := `{"email":"test@example.com","password":"wrongpassword"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidRequest(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/login", handler.Login)

	reqBody := `{"email":"not-an-email"}` // Invalid request
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== GetProfile Tests ====================

func TestUserHandler_GetProfile_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	userID := uuid.New()

	// Middleware to set user_id in context
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetProfile(c)
	})

	expectedUser := &user.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockService.On("GetProfile", userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_Unauthorized(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.GET("/profile", handler.GetProfile) // No user_id in context

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	userID := uuid.New()

	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetProfile(c)
	})

	mockService.On("GetProfile", userID).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== UpdateProfile Tests ====================

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	userID := uuid.New()

	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateProfile(c)
	})

	updatedUser := &user.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Updated",
	}

	mockService.On("UpdateProfile", userID, mock.AnythingOfType("*user.UpdateUserRequest")).Return(updatedUser, nil)

	reqBody := `{"first_name":"Updated"}`
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_Unauthorized(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.PUT("/profile", handler.UpdateProfile) // No user_id

	reqBody := `{"first_name":"Updated"}`
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== DeleteAccount Tests ====================

func TestUserHandler_DeleteAccount_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	userID := uuid.New()

	router.DELETE("/profile", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.DeleteAccount(c)
	})

	mockService.On("DeleteAccount", userID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteAccount_Unauthorized(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.DELETE("/profile", handler.DeleteAccount) // No user_id

	req, _ := http.NewRequest("DELETE", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== RefreshToken Tests ====================

func TestUserHandler_RefreshToken_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/refresh", handler.RefreshToken)

	expectedResponse := &user.LoginResponse{
		Token:        "new-jwt-token",
		RefreshToken: "new-refresh-token",
		User:         &user.User{Email: "test@example.com"},
	}

	mockService.On("RefreshToken", "valid-refresh-token").Return(expectedResponse, nil)

	reqBody := `{"refresh_token":"valid-refresh-token"}`
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_RefreshToken_Invalid(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/refresh", handler.RefreshToken)

	mockService.On("RefreshToken", "invalid-token").Return(nil, assert.AnError)

	reqBody := `{"refresh_token":"invalid-token"}`
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== ForgotPassword Tests ====================

func TestUserHandler_ForgotPassword_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/forgot-password", handler.ForgotPassword)

	mockService.On("ForgotPassword", mock.AnythingOfType("*user.ForgotPasswordRequest")).Return(nil)

	reqBody := `{"email":"test@example.com"}`
	req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ForgotPassword_InvalidRequest(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/forgot-password", handler.ForgotPassword)

	reqBody := `{}` // Missing email
	req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== ResetPassword Tests ====================

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/reset-password", handler.ResetPassword)

	mockService.On("ResetPassword", mock.AnythingOfType("*user.ResetPasswordRequest")).Return(nil)

	reqBody := `{"email":"test@example.com","otp":"123456","new_password":"newpassword123"}`
	req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ResetPassword_InvalidOTP(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/reset-password", handler.ResetPassword)

	mockService.On("ResetPassword", mock.AnythingOfType("*user.ResetPasswordRequest")).Return(assert.AnError)

	reqBody := `{"email":"test@example.com","otp":"000000","new_password":"newpassword123"}`
	req, _ := http.NewRequest("POST", "/reset-password", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
