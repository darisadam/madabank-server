package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-for-testing", 24)

	userID := uuid.New()
	email := "test@madabank.com"
	role := "customer"

	token, expiresAt, err := jwtService.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("Token should not be empty")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatal("ExpiresAt should be in the future")
	}
}

func TestValidateToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-for-testing", 24)

	userID := uuid.New()
	email := "test@madabank.com"
	role := "customer"

	token, _, err := jwtService.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-for-testing", 24)

	invalidToken := "invalid.token.here"

	_, err := jwtService.ValidateToken(invalidToken)
	if err == nil {
		t.Fatal("ValidateToken should fail for invalid token")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("secret-key-1", 24)
	jwtService2 := NewJWTService("secret-key-2", 24)

	userID := uuid.New()
	token, _, err := jwtService1.GenerateToken(userID, "test@madabank.com", "customer")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = jwtService2.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail when using wrong secret")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Create JWT service with very short expiration (negative hours to create expired token)
	jwtService := NewJWTService("test-secret-key-for-testing", -1)

	userID := uuid.New()
	token, _, err := jwtService.GenerateToken(userID, "test@madabank.com", "customer")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// This token should be expired immediately
	_, err = jwtService.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail for expired token")
	}
}
