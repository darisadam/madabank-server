package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
)

type SecurityService interface {
	GetPublicKeyPEM() string
	Decrypt(encryptedBase64 string) (string, error)
}

type securityService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewSecurityService() SecurityService {
	// Generate RSA 2048-bit key pair on startup
	// In production, you might want to persist these keys or rotate them periodically.
	// For this portfolio, ephemeral keys (regenerated on restart) are acceptable and safer than hardcoded ones.
	log.Println("Generating RSA Key Pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate RSA keys: %v", err)
	}

	return &securityService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

func (s *securityService) GetPublicKeyPEM() string {
	pubASN1, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		return ""
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return string(pubPEM)
}

func (s *securityService) Decrypt(encryptedBase64 string) (string, error) {
	// 1. Decode Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}

	// 2. Decrypt using RSA-OAEP with SHA-256
	// This matches what JSEncrypt or standard WebCrypto uses for RSA-OAEP.
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		s.privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
