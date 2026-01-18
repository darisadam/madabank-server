package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/darisadam/madabank-server/internal/pkg/crypto"
	"github.com/darisadam/madabank-server/internal/pkg/jwt"
	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*user.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*user.User, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) SaveRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	args := m.Called(userID, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) GetRefreshToken(tokenHash string) (uuid.UUID, time.Time, error) {
	args := m.Called(tokenHash)
	return args.Get(0).(uuid.UUID), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockUserRepository) RevokeRefreshToken(tokenHash string) error {
	args := m.Called(tokenHash)
	return args.Error(0)
}

func setupTest(t *testing.T) (*userService, *MockUserRepository, *miniredis.Miniredis) {
	// Initialize Logger
	logger.Init("test")

	// Setup Miniredis
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Setup Mock Repo
	mockRepo := new(MockUserRepository)

	// Setup JWT Service
	jwtSvc := jwt.NewJWTService("secret", 1)

	// Create Service
	svc := NewUserService(mockRepo, jwtSvc, redisClient).(*userService)

	return svc, mockRepo, mr
}

func TestForgotPassword_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "test@example.com"

	// Mock User Exists
	mockRepo.On("GetByEmail", email).Return(&user.User{ID: uuid.New(), Email: email}, nil)

	err := svc.ForgotPassword(&user.ForgotPasswordRequest{Email: email})
	assert.NoError(t, err)

	// Verify OTP is stored in Redis
	otpKey := fmt.Sprintf("otp:%s", email)
	exists, _ := svc.redisClient.Exists(context.Background(), otpKey).Result()
	assert.Equal(t, int64(1), exists)

	// Verify Rate Limit is set
	rateLimitKey := fmt.Sprintf("rate_limit:otp:%s", email)
	rlExists, _ := svc.redisClient.Exists(context.Background(), rateLimitKey).Result()
	assert.Equal(t, int64(1), rlExists)
}

func TestForgotPassword_RateLimited(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "rate@example.com"

	// Mock User Exists
	mockRepo.On("GetByEmail", email).Return(&user.User{ID: uuid.New(), Email: email}, nil)

	// Set Rate Limit Key directly
	rateLimitKey := fmt.Sprintf("rate_limit:otp:%s", email)
	svc.redisClient.Set(context.Background(), rateLimitKey, "1", 15*time.Minute)

	err := svc.ForgotPassword(&user.ForgotPasswordRequest{Email: email})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "please wait 15 minutes")
}

func TestForgotPassword_UserNotFound(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "unknown@example.com"

	// Mock User Not Found
	mockRepo.On("GetByEmail", email).Return((*user.User)(nil), fmt.Errorf("user not found"))

	err := svc.ForgotPassword(&user.ForgotPasswordRequest{Email: email})
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestResetPassword_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "reset@example.com"
	otp := "123456"
	newPassword := "newSecret123"
	uid := uuid.New()

	// Setup Redis with Valid OTP
	otpKey := fmt.Sprintf("otp:%s", email)
	svc.redisClient.Set(context.Background(), otpKey, otp, 15*time.Minute)

	// Mock Expectations
	mockRepo.On("GetByEmail", email).Return(&user.User{ID: uid, Email: email}, nil)
	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	err := svc.ResetPassword(&user.ResetPasswordRequest{
		Email:       email,
		OTP:         otp,
		NewPassword: newPassword,
	})
	assert.NoError(t, err)

	// Verify OTP is deleted
	exists, _ := svc.redisClient.Exists(context.Background(), otpKey).Result()
	assert.Equal(t, int64(0), exists)
}

func TestResetPassword_InvalidOTP(t *testing.T) {
	svc, _, _ := setupTest(t)
	email := "invalid@example.com"

	// Setup Redis with Valid OTP
	otpKey := fmt.Sprintf("otp:%s", email)
	svc.redisClient.Set(context.Background(), otpKey, "123456", 15*time.Minute)

	err := svc.ResetPassword(&user.ResetPasswordRequest{
		Email:       email,
		OTP:         "000000", // Wrong OTP
		NewPassword: "new",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid OTP code")
}

func TestResetPassword_ExpiredOTP(t *testing.T) {
	svc, _, _ := setupTest(t)
	email := "expired@example.com"

	// No OTP in Redis represents expired/missing
	err := svc.ResetPassword(&user.ResetPasswordRequest{
		Email:       email,
		OTP:         "123456",
		NewPassword: "new",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired OTP")
}

func TestRegister_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	req := &user.CreateUserRequest{
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Expect GetByEmail to return nil (user not found)
	mockRepo.On("GetByEmail", req.Email).Return((*user.User)(nil), fmt.Errorf("user not found"))
	// Expect Create
	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	u, err := svc.Register(req)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, req.Email, u.Email)
	assert.Empty(t, u.PasswordHash) // Should be cleared
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "duplicate@example.com"
	req := &user.CreateUserRequest{Email: email, Password: "password123"}

	// Expect GetByEmail to return existing user
	mockRepo.On("GetByEmail", email).Return(&user.User{ID: uuid.New(), Email: email}, nil)

	u, err := svc.Register(req)
	assert.Error(t, err)
	assert.Nil(t, u)
	assert.Contains(t, err.Error(), "user with this email already exists")
}

func TestLogin_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "login@example.com"
	password := "password123"

	// Create user with hashed password
	hash, _ := crypto.HashPassword(password)
	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}

	mockRepo.On("GetByEmail", email).Return(u, nil)
	// Expect SaveRefreshToken to be called
	mockRepo.On("SaveRefreshToken", u.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	resp, err := svc.Login(&user.LoginRequest{Email: email, Password: password})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	email := "badpass@example.com"

	hash, _ := crypto.HashPassword("correctPassword")
	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}

	mockRepo.On("GetByEmail", email).Return(u, nil)

	resp, err := svc.Login(&user.LoginRequest{Email: email, Password: "wrongPassword"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestGetProfile_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	uid := uuid.New()
	u := &user.User{ID: uid, Email: "profile@example.com"}

	mockRepo.On("GetByID", uid).Return(u, nil)

	res, err := svc.GetProfile(uid)
	assert.NoError(t, err)
	assert.Equal(t, u.Email, res.Email)
}

func TestUpdateProfile_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	uid := uuid.New()

	first := "New"
	req := &user.UpdateUserRequest{FirstName: &first}

	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	// After update it fetches profile again
	mockRepo.On("GetByID", uid).Return(&user.User{ID: uid, FirstName: "New"}, nil)

	res, err := svc.UpdateProfile(uid, req)
	assert.NoError(t, err)
	assert.Equal(t, "New", res.FirstName)
}

func TestRefreshToken_Success(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	token := "valid_refresh_token"
	uid := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	// Mock GetRefreshToken
	mockRepo.On("GetRefreshToken", token).Return(uid, expiresAt, nil)
	// Mock GetByID
	mockRepo.On("GetByID", uid).Return(&user.User{ID: uid, Email: "refresh@example.com"}, nil)

	resp, err := svc.RefreshToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
}

func TestRefreshToken_Expired(t *testing.T) {
	svc, mockRepo, _ := setupTest(t)
	token := "expired_token"
	uid := uuid.New()
	expiresAt := time.Now().Add(-time.Hour) // Expired

	mockRepo.On("GetRefreshToken", token).Return(uid, expiresAt, nil)

	resp, err := svc.RefreshToken(token)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "refresh token expired")
}
