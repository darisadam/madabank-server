package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/darisadam/madabank-server/internal/pkg/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRateLimitTest(t *testing.T) (*miniredis.Miniredis, *ratelimit.RateLimiter) {
	logger.Init("test")
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	limiter := ratelimit.NewRateLimiter(redisClient)
	return mr, limiter
}

// ==================== RateLimitMiddleware Tests ====================

func TestRateLimitMiddleware_AllowsUnderLimit(t *testing.T) {
	mr, limiter := setupRateLimitTest(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	req, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimitMiddleware_SetsHeaders(t *testing.T) {
	mr, limiter := setupRateLimitTest(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should have rate limit headers
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

// ==================== UserRateLimitMiddleware Tests ====================

func TestUserRateLimitMiddleware_NoUserID(t *testing.T) {
	mr, limiter := setupRateLimitTest(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(UserRateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request without user_id should pass through
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== SuspiciousActivityMiddleware Tests ====================

func TestSuspiciousActivityMiddleware_PassesNormalRequest(t *testing.T) {
	mr, limiter := setupRateLimitTest(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SuspiciousActivityMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== getRateLimitConfig Tests ====================

func TestGetRateLimitConfig_AuthEndpoints(t *testing.T) {
	config := getRateLimitConfig("/api/v1/auth/login")
	assert.Equal(t, ratelimit.AuthRateLimit, config)

	config = getRateLimitConfig("/api/v1/auth/register")
	assert.Equal(t, ratelimit.AuthRateLimit, config)
}

func TestGetRateLimitConfig_TransactionEndpoints(t *testing.T) {
	config := getRateLimitConfig("/api/v1/transactions/transfer")
	assert.Equal(t, ratelimit.TransactionRateLimit, config)

	config = getRateLimitConfig("/api/v1/transactions/deposit")
	assert.Equal(t, ratelimit.TransactionRateLimit, config)

	config = getRateLimitConfig("/api/v1/transactions/withdraw")
	assert.Equal(t, ratelimit.TransactionRateLimit, config)
}

func TestGetRateLimitConfig_GeneralEndpoints(t *testing.T) {
	config := getRateLimitConfig("/api/v1/accounts")
	assert.Equal(t, ratelimit.GeneralRateLimit, config)

	config = getRateLimitConfig("/api/v1/users/profile")
	assert.Equal(t, ratelimit.GeneralRateLimit, config)
}
