package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRateLimiterTest(t *testing.T) (*RateLimiter, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRateLimiter(client)
	return rl, mr
}

func TestCheckLimit_WithinLimit(t *testing.T) {
	rl, mr := setupRateLimiterTest(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:user:123"
	config := RateLimitConfig{
		Requests: 5,
		Window:   time.Minute,
	}

	// First request should be allowed
	allowed, err := rl.CheckLimit(ctx, key, config)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Additional requests up to limit should be allowed
	for i := 0; i < 3; i++ {
		allowed, err = rl.CheckLimit(ctx, key, config)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}
}

func TestCheckLimit_ExceedsLimit(t *testing.T) {
	rl, mr := setupRateLimiterTest(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:exceed:456"
	config := RateLimitConfig{
		Requests: 3,
		Window:   time.Minute,
	}

	// Make requests up to limit
	for i := 0; i < 3; i++ {
		allowed, err := rl.CheckLimit(ctx, key, config)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	// 4th request should be blocked
	allowed, err := rl.CheckLimit(ctx, key, config)
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckLimitWithInfo_Returns_Correct_Info(t *testing.T) {
	rl, mr := setupRateLimiterTest(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:info:789"
	config := RateLimitConfig{
		Requests: 5,
		Window:   time.Minute,
	}

	info, err := rl.CheckLimitWithInfo(ctx, key, config)
	assert.NoError(t, err)
	assert.True(t, info.Allowed)
	assert.Equal(t, 5, info.Limit)
	assert.Equal(t, 5, info.Remaining) // First request, so remaining = limit

	// Make more requests
	_, err = rl.CheckLimitWithInfo(ctx, key, config)
	assert.NoError(t, err)
	_, err = rl.CheckLimitWithInfo(ctx, key, config)
	assert.NoError(t, err)

	info, err = rl.CheckLimitWithInfo(ctx, key, config)
	assert.NoError(t, err)
	assert.True(t, info.Allowed)
	assert.Equal(t, 2, info.Remaining) // After 3 requests, 2 remaining
}

func TestBlock_And_IsBlocked(t *testing.T) {
	rl, mr := setupRateLimiterTest(t)
	defer mr.Close()

	ctx := context.Background()
	key := "suspicious:ip:1.2.3.4"

	// Initially not blocked
	blocked, err := rl.IsBlocked(ctx, key)
	assert.NoError(t, err)
	assert.False(t, blocked)

	// Block the key
	err = rl.Block(ctx, key, 5*time.Minute)
	assert.NoError(t, err)

	// Now should be blocked
	blocked, err = rl.IsBlocked(ctx, key)
	assert.NoError(t, err)
	assert.True(t, blocked)
}

func TestBlock_Expires(t *testing.T) {
	rl, mr := setupRateLimiterTest(t)
	defer mr.Close()

	ctx := context.Background()
	key := "temp:block:user"

	// Block for 1 second
	err := rl.Block(ctx, key, 1*time.Second)
	assert.NoError(t, err)

	// Immediately blocked
	blocked, _ := rl.IsBlocked(ctx, key)
	assert.True(t, blocked)

	// Fast-forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Should no longer be blocked
	blocked, err = rl.IsBlocked(ctx, key)
	assert.NoError(t, err)
	assert.False(t, blocked)
}

func TestPredefinedConfigs(t *testing.T) {
	// Test that predefined configs are sensible
	assert.Equal(t, 5, AuthRateLimit.Requests)
	assert.Equal(t, time.Minute, AuthRateLimit.Window)

	assert.Equal(t, 10, TransactionRateLimit.Requests)
	assert.Equal(t, time.Minute, TransactionRateLimit.Window)

	assert.Equal(t, 100, GeneralRateLimit.Requests)
	assert.Equal(t, time.Minute, GeneralRateLimit.Window)

	assert.Equal(t, 1, SuspiciousRateLimit.Requests)
	assert.Equal(t, 5*time.Minute, SuspiciousRateLimit.Window)
}
