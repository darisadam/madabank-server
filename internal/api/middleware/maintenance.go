package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/darisadam/madabank-server/internal/pkg/logger"
)

// MaintenanceMiddleware checks if the system is in maintenance mode
func MaintenanceMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Public health/metrics endpoints should always be accessible
		path := c.Request.URL.Path
		if path == "/health" || path == "/ready" || path == "/metrics" || path == "/version" || path == "/" {
			c.Next()
			return
		}

		// Check Redis for maintenance flag
		// We use a short timeout context to avoid blocking if Redis is slow
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		val, err := redisClient.Get(ctx, "system:maintenance").Result()
		if err != nil && err != redis.Nil {
			// Log error but continue (fail open) to avoid outage if Redis is down
			logger.Error("Failed to check maintenance mode", zap.Error(err))
			c.Next()
			return
		}

		if val == "true" {
			// Check for bypass token (optional, e.g. for admins testing)
			if c.GetHeader("X-Maintenance-Bypass") == "madabank-admin-bypass" {
				c.Next()
				return
			}

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service under maintenance",
				"message": "We are currently upgrading our systems. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
