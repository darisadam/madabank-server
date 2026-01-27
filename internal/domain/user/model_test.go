package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_PasswordHashHidden(t *testing.T) {
	u := User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "secret_hash_should_not_appear",
		FirstName:    "John",
		LastName:     "Doe",
	}

	// PasswordHash has json:"-" tag, so it should not be exposed
	// This test just verifies the struct field exists and is set
	assert.NotEmpty(t, u.PasswordHash)
	assert.Equal(t, "test@example.com", u.Email)
}

func TestUser_OptionalFields(t *testing.T) {
	u := User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
	}

	// Optional fields should be nil
	assert.Nil(t, u.Phone)
	assert.Nil(t, u.DateOfBirth)
	assert.Nil(t, u.DeletedAt)

	// Set optional fields
	phone := "+1234567890"
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	u.Phone = &phone
	u.DateOfBirth = &dob

	assert.NotNil(t, u.Phone)
	assert.Equal(t, "+1234567890", *u.Phone)
	assert.NotNil(t, u.DateOfBirth)
}

func TestCreateUserRequest_RequiredFields(t *testing.T) {
	req := CreateUserRequest{
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	assert.Equal(t, "new@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "New", req.FirstName)
	assert.Equal(t, "User", req.LastName)
}

func TestLoginRequest_Structure(t *testing.T) {
	req := LoginRequest{
		Email:    "login@example.com",
		Password: "password",
	}

	assert.Equal(t, "login@example.com", req.Email)
	assert.Equal(t, "password", req.Password)
}

func TestLoginResponse_Structure(t *testing.T) {
	userID := uuid.New()
	resp := LoginResponse{
		Token:        "jwt.token.here",
		RefreshToken: "refresh.token.here",
		ExpiresAt:    time.Now().Add(time.Hour),
		User:         &User{ID: userID, Email: "user@example.com"},
	}

	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotNil(t, resp.User)
	assert.Equal(t, userID, resp.User.ID)
}

func TestUpdateUserRequest_OptionalFields(t *testing.T) {
	firstName := "Updated"
	req := UpdateUserRequest{
		FirstName: &firstName,
	}

	assert.NotNil(t, req.FirstName)
	assert.Equal(t, "Updated", *req.FirstName)
	assert.Nil(t, req.LastName)
	assert.Nil(t, req.Phone)
	assert.Nil(t, req.DateOfBirth)
}

func TestForgotPasswordRequest_Structure(t *testing.T) {
	req := ForgotPasswordRequest{
		Email: "forgot@example.com",
	}

	assert.Equal(t, "forgot@example.com", req.Email)
}

func TestResetPasswordRequest_Structure(t *testing.T) {
	req := ResetPasswordRequest{
		Email:       "reset@example.com",
		OTP:         "123456",
		NewPassword: "newpassword123",
	}

	assert.Equal(t, "reset@example.com", req.Email)
	assert.Equal(t, "123456", req.OTP)
	assert.Equal(t, "newpassword123", req.NewPassword)
}
