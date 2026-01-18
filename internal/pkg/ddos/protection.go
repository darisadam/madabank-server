package ddos

import (
	"context"
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DDoSProtection struct {
	redis *redis.Client
}

func NewDDoSProtection(redisClient *redis.Client) *DDoSProtection {
	return &DDoSProtection{
		redis: redisClient,
	}
}

// TrackRequest tracks incoming requests per IP
func (d *DDoSProtection) TrackRequest(ctx context.Context, ip string) error {
	key := fmt.Sprintf("ddos:requests:%s", ip)

	pipe := d.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Minute)

	_, err := pipe.Exec(ctx)
	return err
}

// CheckThreshold checks if IP has exceeded DDoS threshold
func (d *DDoSProtection) CheckThreshold(ctx context.Context, ip string, threshold int64) (bool, error) {
	key := fmt.Sprintf("ddos:requests:%s", ip)

	count, err := d.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return count > threshold, nil
}

// MonitorGlobalTraffic monitors overall system traffic
func (d *DDoSProtection) MonitorGlobalTraffic(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.analyzeTraffic(ctx)
		}
	}
}

func (d *DDoSProtection) analyzeTraffic(ctx context.Context) {
	// Get all IP request counts
	pattern := "ddos:requests:*"
	iter := d.redis.Scan(ctx, 0, pattern, 0).Iterator()

	suspiciousIPs := []string{}

	for iter.Next(ctx) {
		key := iter.Val()
		count, err := d.redis.Get(ctx, key).Int64()
		if err != nil {
			continue
		}

		// Flag IPs with > 1000 requests/minute
		if count > 1000 {
			ip := key[len("ddos:requests:"):]
			suspiciousIPs = append(suspiciousIPs, ip)

			logger.Warn("Potential DDoS attack detected",
				zap.String("ip", ip),
				zap.Int64("requests_per_minute", count),
			)
		}
	}

	if err := iter.Err(); err != nil {
		logger.Error("Failed to scan traffic", zap.Error(err))
	}

	// If many IPs are attacking, enable global rate limiting
	if len(suspiciousIPs) > 10 {
		logger.Error("Large-scale DDoS attack detected",
			zap.Int("suspicious_ips", len(suspiciousIPs)),
		)

		// TODO: Enable emergency mode
		// TODO: Send alert to ops team
	}
}
