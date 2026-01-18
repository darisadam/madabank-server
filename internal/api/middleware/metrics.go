package middleware

import (
	"time"

	"github.com/darisadam/madabank-server/internal/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// MetricsMiddleware records HTTP metrics for each request
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method

		// Fix status conversion (rune to string is likely wrong in previous code,
		// but RecordHTTPRequest takes string(rune(status)) which gives a symbol, not number.
		// Wait, user provided code: HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
		// string(rune(200)) is definitely not "200". It's a character.
		// I should fix this bug while I'm at it?
		// The user provided the code. I restore it as is, or fix it?
		// User provided: string(rune(status)) in metrics.go.
		// Let's check metrics.go content I just wrote.
		// It has: HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
		// This is bug. I will fix it in metrics.go via Replace or Write.
		// But first, let's restore middleware.

		metrics.RecordHTTPRequest(method, path, status, duration)
	}
}
