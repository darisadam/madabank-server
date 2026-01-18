package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

type RateLimitConfig struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
}

// Common rate limit configurations
var (
	// Auth endpoints - stricter limits
	AuthRateLimit = RateLimitConfig{
		Requests: 5,
		Window:   time.Minute,
	}

	// Transaction endpoints - moderate limits
	TransactionRateLimit = RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
	}

	// General API - generous limits
	GeneralRateLimit = RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
	}

	// Suspicious activity - very strict
	SuspiciousRateLimit = RateLimitConfig{
		Requests: 1,
		Window:   5 * time.Minute,
	}
)

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: redisClient,
	}
}

// CheckLimit checks if the request is within rate limits
func (rl *RateLimiter) CheckLimit(ctx context.Context, key string, config RateLimitConfig) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-config.Window)

	// Use Redis sorted set to track requests within the time window
	// Score is timestamp, member is unique request ID
	pipe := rl.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))

	// Count requests in the current window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration on the key
	pipe.Expire(ctx, key, config.Window+time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	count := countCmd.Val()

	// Allow if count is less than limit
	return count < int64(config.Requests), nil
}

// CheckLimitWithInfo checks rate limit and returns detailed info
func (rl *RateLimiter) CheckLimitWithInfo(ctx context.Context, key string, config RateLimitConfig) (*RateLimitInfo, error) {
	now := time.Now()
	windowStart := now.Add(-config.Window)

	pipe := rl.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))

	// Count current requests
	countCmd := pipe.ZCard(ctx, key)

	// Get oldest request in window
	oldestCmd := pipe.ZRange(ctx, key, 0, 0)

	// Add current request if within limit
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	pipe.Expire(ctx, key, config.Window+time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	count := countCmd.Val()
	allowed := count < int64(config.Requests)

	info := &RateLimitInfo{
		Limit:     config.Requests,
		Remaining: config.Requests - int(count),
		Reset:     now.Add(config.Window),
		Allowed:   allowed,
	}

	if info.Remaining < 0 {
		info.Remaining = 0
	}

	// Calculate retry after if blocked
	if !allowed && len(oldestCmd.Val()) > 0 {
		// Parse oldest timestamp
		_ = oldestCmd.Val()[0]
		// Set retry after based on when oldest request will expire
		info.RetryAfter = config.Window
	}

	return info, nil
}

// Block temporarily blocks a key (for suspicious activity)
func (rl *RateLimiter) Block(ctx context.Context, key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("blocked:%s", key)
	return rl.client.Set(ctx, blockKey, "1", duration).Err()
}

// IsBlocked checks if a key is blocked
func (rl *RateLimiter) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("blocked:%s", key)
	result, err := rl.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

type RateLimitInfo struct {
	Limit      int           `json:"limit"`
	Remaining  int           `json:"remaining"`
	Reset      time.Time     `json:"reset"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
	Allowed    bool          `json:"allowed"`
}
