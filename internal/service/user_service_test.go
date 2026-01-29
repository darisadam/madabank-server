package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/darisadam/madabank-server/internal/domain/card"
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

func (m *MockUserRepository) GetByPhone(phone string) (*user.User, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// MockAccountRepositoryForUser is a mock implementation for UserService tests
type MockAccountRepositoryForUser struct {
	mock.Mock
}

func (m *MockAccountRepositoryForUser) Create(a *account.Account) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockAccountRepositoryForUser) GetByID(id uuid.UUID) (*account.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepositoryForUser) GetByUserID(userID uuid.UUID) ([]*account.Account, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountRepositoryForUser) GetByAccountNumber(number string) (*account.Account, error) {
	args := m.Called(number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepositoryForUser) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockAccountRepositoryForUser) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAccountRepositoryForUser) GenerateAccountNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAccountRepositoryForUser) UpdateBalance(id uuid.UUID, amount float64) error {
	args := m.Called(id, amount)
	return args.Error(0)
}

// MockCardRepositoryForUser is a mock implementation for UserService tests
type MockCardRepositoryForUser struct {
	mock.Mock
}

func (m *MockCardRepositoryForUser) Create(c *card.Card) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockCardRepositoryForUser) GetByID(id uuid.UUID) (*card.Card, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.Card), args.Error(1)
}

func (m *MockCardRepositoryForUser) GetByAccountID(accountID uuid.UUID) ([]*card.Card, error) {
	args := m.Called(accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*card.Card), args.Error(1)
}

func (m *MockCardRepositoryForUser) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCardRepositoryForUser) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCardRepositoryForUser) GenerateCardNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockCardRepositoryForUser) GenerateCVV() string {
	args := m.Called()
	return args.String(0)
}

func setupTest(t *testing.T) (*userService, *MockUserRepository, *MockAccountRepositoryForUser, *MockCardRepositoryForUser, *miniredis.Miniredis) {
	// Initialize Logger
	logger.Init("test")

	// Setup Miniredis
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Setup Mock Repos
	mockUserRepo := new(MockUserRepository)
	mockAccountRepo := new(MockAccountRepositoryForUser)
	mockCardRepo := new(MockCardRepositoryForUser)

	// Setup JWT Service
	jwtSvc := jwt.NewJWTService("secret", 1)

	// Setup Encryptor (32-byte key for AES-256)
	encryptor, _ := crypto.NewEncryptor("12345678901234567890123456789012")

	// Create Service
	svc := NewUserService(mockUserRepo, mockAccountRepo, mockCardRepo, jwtSvc, redisClient, encryptor).(*userService)

	return svc, mockUserRepo, mockAccountRepo, mockCardRepo, mr
}

func TestForgotPassword_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
	email := "unknown@example.com"

	// Mock User Not Found
	mockRepo.On("GetByEmail", email).Return((*user.User)(nil), fmt.Errorf("user not found"))

	err := svc.ForgotPassword(&user.ForgotPasswordRequest{Email: email})
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestResetPassword_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, _, _, _, _ := setupTest(t)
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
	svc, _, _, _, _ := setupTest(t)
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
	svc, mockRepo, mockAccountRepo, mockCardRepo, _ := setupTest(t)
	req := &user.CreateUserRequest{
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Expect GetByEmail to return nil (user not found)
	mockRepo.On("GetByEmail", req.Email).Return((*user.User)(nil), fmt.Errorf("user not found"))
	// Expect Create user
	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	// Auto-onboarding mocks
	mockAccountRepo.On("GenerateAccountNumber").Return("1234567890", nil)
	mockAccountRepo.On("Create", mock.AnythingOfType("*account.Account")).Return(nil)
	mockCardRepo.On("GenerateCardNumber").Return("4111111111111111", nil)
	mockCardRepo.On("GenerateCVV").Return("123")
	mockCardRepo.On("Create", mock.AnythingOfType("*card.Card")).Return(nil)

	u, err := svc.Register(req)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, req.Email, u.Email)
	assert.Empty(t, u.PasswordHash) // Should be cleared
	mockAccountRepo.AssertExpectations(t)
	mockCardRepo.AssertExpectations(t)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
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

func TestLogin_ByPhone_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	phone := "+6281234567890"
	password := "password123"

	// Create user with hashed password
	hash, _ := crypto.HashPassword(password)
	u := &user.User{
		ID:           uuid.New(),
		Email:        "phone-user@example.com",
		Phone:        &phone,
		PasswordHash: hash,
		IsActive:     true,
	}

	mockRepo.On("GetByPhone", phone).Return(u, nil)
	mockRepo.On("SaveRefreshToken", u.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	resp, err := svc.Login(&user.LoginRequest{Phone: phone, Password: password})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestLogin_NoCredentials(t *testing.T) {
	svc, _, _, _, _ := setupTest(t)

	resp, err := svc.Login(&user.LoginRequest{Password: "password123"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "email or phone number is required")
}

func TestGetProfile_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()
	u := &user.User{ID: uid, Email: "profile@example.com"}

	mockRepo.On("GetByID", uid).Return(u, nil)

	res, err := svc.GetProfile(uid)
	assert.NoError(t, err)
	assert.Equal(t, u.Email, res.Email)
}

func TestUpdateProfile_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
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
	svc, mockRepo, _, _, _ := setupTest(t)
	token := "expired_token"
	uid := uuid.New()
	expiresAt := time.Now().Add(-time.Hour) // Expired

	mockRepo.On("GetRefreshToken", token).Return(uid, expiresAt, nil)

	resp, err := svc.RefreshToken(token)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "refresh token expired")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	token := "invalid_token"

	mockRepo.On("GetRefreshToken", token).Return(uuid.Nil, time.Time{}, fmt.Errorf("invalid or revoked token"))

	resp, err := svc.RefreshToken(token)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	token := "valid_token"
	uid := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	mockRepo.On("GetRefreshToken", token).Return(uid, expiresAt, nil)
	mockRepo.On("GetByID", uid).Return(nil, fmt.Errorf("user not found"))

	resp, err := svc.RefreshToken(token)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ==================== DeleteAccount Tests ====================

func TestDeleteAccount_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	mockRepo.On("Delete", uid).Return(nil)

	err := svc.DeleteAccount(uid)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteAccount_Error(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	mockRepo.On("Delete", uid).Return(fmt.Errorf("database error"))

	err := svc.DeleteAccount(uid)
	assert.Error(t, err)
}

// ==================== UpdateProfile Tests ====================

func TestUpdateProfile_WithMultipleFields_Success(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	firstName := "Updated"
	lastName := "Name"
	req := &user.UpdateUserRequest{
		FirstName: &firstName,
		LastName:  &lastName,
	}

	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRepo.On("GetByID", uid).Return(&user.User{
		ID:        uid,
		FirstName: firstName,
		LastName:  lastName,
	}, nil)

	result, err := svc.UpdateProfile(uid, req)
	assert.NoError(t, err)
	assert.Equal(t, firstName, result.FirstName)
	assert.Equal(t, lastName, result.LastName)
}

func TestUpdateProfile_NoChanges(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	req := &user.UpdateUserRequest{} // No fields set

	// Should just return profile without updating
	mockRepo.On("GetByID", uid).Return(&user.User{
		ID:    uid,
		Email: "test@example.com",
	}, nil)

	result, err := svc.UpdateProfile(uid, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateProfile_InvalidDateOfBirth(t *testing.T) {
	svc, _, _, _, _ := setupTest(t)
	uid := uuid.New()

	invalidDOB := "not-a-date"
	req := &user.UpdateUserRequest{
		DateOfBirth: &invalidDOB,
	}

	result, err := svc.UpdateProfile(uid, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid date of birth format")
}

func TestUpdateProfile_UpdateFails(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	firstName := "Updated"
	req := &user.UpdateUserRequest{
		FirstName: &firstName,
	}

	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(fmt.Errorf("database error"))

	result, err := svc.UpdateProfile(uid, req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateProfile_WithPhone(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	phone := "+6281234567890"
	req := &user.UpdateUserRequest{
		Phone: &phone,
	}

	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRepo.On("GetByID", uid).Return(&user.User{
		ID:    uid,
		Phone: &phone,
	}, nil)

	result, err := svc.UpdateProfile(uid, req)
	assert.NoError(t, err)
	assert.Equal(t, &phone, result.Phone)
}

func TestUpdateProfile_WithValidDateOfBirth(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	dob := "1990-05-15"
	req := &user.UpdateUserRequest{
		DateOfBirth: &dob,
	}

	mockRepo.On("Update", uid, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRepo.On("GetByID", uid).Return(&user.User{
		ID: uid,
	}, nil)

	result, err := svc.UpdateProfile(uid, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// ==================== GetProfile Tests ====================

func TestGetProfile_Error(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	uid := uuid.New()

	mockRepo.On("GetByID", uid).Return(nil, fmt.Errorf("database error"))

	result, err := svc.GetProfile(uid)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Login Additional Tests ====================

func TestLogin_InactiveUser(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	email := "inactive@example.com"
	password := "password123"

	hash, _ := crypto.HashPassword(password)
	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		IsActive:     false, // Inactive user
	}

	mockRepo.On("GetByEmail", email).Return(u, nil)

	resp, err := svc.Login(&user.LoginRequest{Email: email, Password: password})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "account is inactive")
}

func TestLogin_ByPhone_UserNotFound(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	phone := "+6281234567890"

	mockRepo.On("GetByPhone", phone).Return(nil, fmt.Errorf("user not found"))

	resp, err := svc.Login(&user.LoginRequest{Phone: phone, Password: "password"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid phone number or password")
}

func TestLogin_ByPhone_WrongPassword(t *testing.T) {
	svc, mockRepo, _, _, _ := setupTest(t)
	phone := "+6281234567890"

	hash, _ := crypto.HashPassword("correct_password")
	u := &user.User{
		ID:           uuid.New(),
		Phone:        &phone,
		PasswordHash: hash,
		IsActive:     true,
	}

	mockRepo.On("GetByPhone", phone).Return(u, nil)

	resp, err := svc.Login(&user.LoginRequest{Phone: phone, Password: "wrong_password"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid phone number or password")
}
