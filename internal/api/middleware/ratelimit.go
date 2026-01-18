package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/darisadam/madabank-server/internal/pkg/metrics"
	"github.com/darisadam/madabank-server/internal/pkg/ratelimit"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimitMiddleware applies rate limiting based on IP address and endpoint
func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Get client IP
		clientIP := c.ClientIP()

		// Determine rate limit config based on endpoint
		config := getRateLimitConfig(c.FullPath())

		logger.Info("Rate Limit Check",
			zap.String("path", c.FullPath()),
			zap.String("ip", clientIP),
			zap.Int("limit", config.Requests),
		)

		// Create rate limit key
		key := fmt.Sprintf("ratelimit:%s:%s", clientIP, c.FullPath())

		// Check if IP is blocked
		blocked, err := limiter.IsBlocked(ctx, clientIP)
		if err != nil {
			logger.Error("Failed to check block status", zap.Error(err))
		}

		if blocked {
			metrics.RecordAuthAttempt(false)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Your IP has been temporarily blocked.",
				"retry_after": "5 minutes",
			})
			c.Abort()
			return
		}

		// Check rate limit
		info, err := limiter.CheckLimitWithInfo(ctx, key, config)
		if err != nil {
			logger.Error("Rate limit check failed", zap.Error(err))
			// Continue on error (fail open)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset.Unix()))

		if !info.Allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(info.RetryAfter.Seconds())))

			logger.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.FullPath()),
				zap.Int("limit", info.Limit),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"limit":       info.Limit,
				"retry_after": fmt.Sprintf("%d seconds", int(info.RetryAfter.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimitMiddleware applies rate limiting per authenticated user
func UserRateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Determine rate limit config
		config := getRateLimitConfig(c.FullPath())

		// Create user-specific rate limit key
		key := fmt.Sprintf("ratelimit:user:%s:%s", userID, c.FullPath())

		// Check rate limit
		info, err := limiter.CheckLimitWithInfo(ctx, key, config)
		if err != nil {
			logger.Error("User rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-User-Limit", fmt.Sprintf("%d", info.Limit))
		c.Header("X-RateLimit-User-Remaining", fmt.Sprintf("%d", info.Remaining))

		if !info.Allowed {
			logger.Warn("User rate limit exceeded",
				zap.Any("user_id", userID),
				zap.String("path", c.FullPath()),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "User rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRateLimitConfig returns appropriate rate limit based on endpoint
func getRateLimitConfig(path string) ratelimit.RateLimitConfig {
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/register":
		return ratelimit.AuthRateLimit
	case "/api/v1/transactions/transfer", "/api/v1/transactions/deposit", "/api/v1/transactions/withdraw":
		return ratelimit.TransactionRateLimit
	default:
		return ratelimit.GeneralRateLimit
	}
}

// SuspiciousActivityMiddleware detects and blocks suspicious patterns
func SuspiciousActivityMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		clientIP := c.ClientIP()

		// Check for suspicious patterns after request completes
		c.Next()

		// Detect suspicious activity based on status code
		status := c.Writer.Status()

		if status == http.StatusUnauthorized && c.FullPath() == "/api/v1/auth/login" {
			// Failed login attempt
			key := fmt.Sprintf("suspicious:login:%s", clientIP)

			allowed, err := limiter.CheckLimit(ctx, key, ratelimit.RateLimitConfig{
				Requests: 5,
				Window:   15 * time.Minute,
			})

			if err != nil {
				logger.Error("Suspicious activity check failed", zap.Error(err))
				return
			}

			if !allowed {
				// Block IP for 1 hour after 5 failed attempts
				logger.Warn("Blocking IP due to multiple failed login attempts",
					zap.String("ip", clientIP),
				)

				err := limiter.Block(ctx, clientIP, time.Hour)
				if err != nil {
					logger.Error("Failed to block IP", zap.Error(err))
				}

				// TODO: Send alert to security team
			}
		}
	}
}
