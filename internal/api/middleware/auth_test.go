package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darisadam/madabank-server/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== AuthMiddleware Tests ====================

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header required")
}

func TestAuthMiddleware_InvalidHeaderFormat_NoBearerPrefix(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_InvalidHeaderFormat_WrongPrefix(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	// Generate a valid token
	userID := uuid.New()
	token, _, err := jwtService.GenerateToken(userID, "test@example.com", "user")
	assert.NoError(t, err)

	var capturedUserID uuid.UUID
	var capturedEmail string
	var capturedRole string

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		capturedUserID = c.MustGet("user_id").(uuid.UUID)
		capturedEmail = c.MustGet("email").(string)
		capturedRole = c.MustGet("role").(string)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, "test@example.com", capturedEmail)
	assert.Equal(t, "user", capturedRole)
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	jwtService := jwt.NewJWTService("test-secret", 1)
	router := setupTestRouter()

	router.Use(AuthMiddleware(jwtService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
