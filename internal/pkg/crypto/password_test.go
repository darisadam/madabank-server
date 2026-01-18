package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "SecurePassword123!"
	wrongPassword := "WrongPassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if !CheckPassword(password, hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}

	// Test wrong password
	if CheckPassword(wrongPassword, hash) {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "SecurePassword123!"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Bcrypt should generate different hashes each time (due to random salt)
	if hash1 == hash2 {
		t.Fatal("Two hashes of the same password should be different")
	}

	// But both should validate correctly
	if !CheckPassword(password, hash1) || !CheckPassword(password, hash2) {
		t.Fatal("Both hashes should validate the password")
	}
}
