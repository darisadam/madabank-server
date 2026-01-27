package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecurityService is a mock implementation of service.SecurityService
type MockSecurityService struct {
	mock.Mock
}

func (m *MockSecurityService) GetPublicKeyPEM() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSecurityService) Decrypt(encryptedBase64 string) (string, error) {
	args := m.Called(encryptedBase64)
	return args.String(0), args.Error(1)
}

func setupSecurityRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== GetPublicKey Tests ====================

func TestSecurityHandler_GetPublicKey_Success(t *testing.T) {
	mockService := new(MockSecurityService)
	handler := NewSecurityHandler(mockService)

	router := setupSecurityRouter()
	router.GET("/security/public-key", handler.GetPublicKey)

	mockPublicKey := "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkq...\n-----END PUBLIC KEY-----"
	mockService.On("GetPublicKeyPEM").Return(mockPublicKey)

	req, _ := http.NewRequest("GET", "/security/public-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// The handler returns plain text PEM, not JSON
	assert.Contains(t, w.Body.String(), "-----BEGIN PUBLIC KEY-----")
	mockService.AssertExpectations(t)
}

func TestSecurityHandler_GetPublicKey_Empty(t *testing.T) {
	mockService := new(MockSecurityService)
	handler := NewSecurityHandler(mockService)

	router := setupSecurityRouter()
	router.GET("/security/public-key", handler.GetPublicKey)

	mockService.On("GetPublicKeyPEM").Return("")

	req, _ := http.NewRequest("GET", "/security/public-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
