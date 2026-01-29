package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEncryptor_ValidKey(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	enc, err := NewEncryptor(key)
	assert.NoError(t, err)
	assert.NotNil(t, enc)
}

func TestNewEncryptor_InvalidKeyLength(t *testing.T) {
	key := "short"
	enc, err := NewEncryptor(key)
	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Contains(t, err.Error(), "must be 32 bytes")
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, err := NewEncryptor(key)
	assert.NoError(t, err)

	plaintext := "4111111111111111"

	ciphertext, err := enc.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := enc.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewEncryptor(key)

	ciphertext, err := enc.Encrypt("")
	assert.NoError(t, err)

	decrypted, err := enc.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestEncrypt_DifferentNonces(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewEncryptor(key)

	plaintext := "secret"

	ciphertext1, _ := enc.Encrypt(plaintext)
	ciphertext2, _ := enc.Encrypt(plaintext)

	// Same plaintext should produce different ciphertext (random nonce)
	assert.NotEqual(t, ciphertext1, ciphertext2)
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewEncryptor(key)

	_, err := enc.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base64")
}

func TestDecrypt_TooShortCiphertext(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewEncryptor(key)

	// Minimal base64 that decodes to less than nonce size
	_, err := enc.Decrypt("aGVsbG8=")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewEncryptor(key)

	plaintext := "secret"
	ciphertext, _ := enc.Encrypt(plaintext)

	// Tamper with ciphertext
	tamperedBytes := []byte(ciphertext)
	if len(tamperedBytes) > 5 {
		tamperedBytes[5] = 'X'
	}
	tampered := string(tamperedBytes)

	_, err := enc.Decrypt(tampered)
	assert.Error(t, err)
}

func TestMaskCardNumber_Standard(t *testing.T) {
	masked := MaskCardNumber("4111111111111111")
	assert.Equal(t, "************1111", masked)
}

func TestMaskCardNumber_Short(t *testing.T) {
	masked := MaskCardNumber("123")
	assert.Equal(t, "****", masked)
}

func TestMaskCardNumber_Empty(t *testing.T) {
	masked := MaskCardNumber("")
	assert.Equal(t, "****", masked)
}

func TestValidateCardNumber_ValidLuhn(t *testing.T) {
	// Valid Visa test card number
	assert.True(t, ValidateCardNumber("4111111111111111"))
	// Valid Mastercard test number
	assert.True(t, ValidateCardNumber("5500000000000004"))
}

func TestValidateCardNumber_InvalidLuhn(t *testing.T) {
	assert.False(t, ValidateCardNumber("4111111111111112"))
	assert.False(t, ValidateCardNumber("1234567890123456"))
}

func TestValidateCardNumber_NonDigit(t *testing.T) {
	assert.False(t, ValidateCardNumber("4111-1111-1111-1111"))
	assert.False(t, ValidateCardNumber("4111abcd11111111"))
}

func TestValidateCVV_Valid(t *testing.T) {
	assert.True(t, ValidateCVV("123"))
	assert.True(t, ValidateCVV("1234")) // Amex
}

func TestValidateCVV_Invalid(t *testing.T) {
	assert.False(t, ValidateCVV("12"))
	assert.False(t, ValidateCVV("12345"))
	assert.False(t, ValidateCVV("abc"))
}
