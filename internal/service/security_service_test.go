package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityService(t *testing.T) {
	svc := NewSecurityService()
	assert.NotNil(t, svc)
}

func TestSecurityService_GetPublicKeyPEM(t *testing.T) {
	svc := NewSecurityService()

	publicKeyPEM := svc.GetPublicKeyPEM()
	assert.NotEmpty(t, publicKeyPEM)
	assert.Contains(t, publicKeyPEM, "-----BEGIN PUBLIC KEY-----")
	assert.Contains(t, publicKeyPEM, "-----END PUBLIC KEY-----")
}

func TestSecurityService_Decrypt_Success(t *testing.T) {
	svc := NewSecurityService().(*securityService)

	// Encrypt a message using the public key
	plaintext := "test-secret-data"
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		svc.publicKey,
		[]byte(plaintext),
		nil,
	)
	assert.NoError(t, err)

	// Encode to base64 (what the client would send)
	encryptedBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	// Decrypt
	result, err := svc.Decrypt(encryptedBase64)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, result)
}

func TestSecurityService_Decrypt_InvalidBase64(t *testing.T) {
	svc := NewSecurityService()

	result, err := svc.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "invalid base64")
}

func TestSecurityService_Decrypt_InvalidCiphertext(t *testing.T) {
	svc := NewSecurityService()

	// Valid base64 but not valid RSA ciphertext
	invalidCiphertext := base64.StdEncoding.EncodeToString([]byte("not-encrypted-data"))

	result, err := svc.Decrypt(invalidCiphertext)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "decryption failed")
}
