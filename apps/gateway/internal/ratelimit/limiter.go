package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter enforces fixed-window request counts using Redis.
type Limiter struct {
	client *redis.Client
	prefix string
}

func NewLimiter(addr string) (*Limiter, error) {
	if addr == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Limiter{client: client, prefix: "agentvoir:ratelimit:"}, nil
}

// Allow increments the counter for key and reports whether limit is exceeded.
// limit is requests allowed per 60-second window.
func (l *Limiter) Allow(ctx context.Context, key string, limit int64) (allowed bool, retryAfterSec int) {
	if l == nil || limit <= 0 {
		return true, 0
	}
	window := time.Now().UTC().Unix() / 60
	redisKey := fmt.Sprintf("%s%s:%d", l.prefix, key, window)

	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return true, 0
	}
	if count == 1 {
		_ = l.client.Expire(ctx, redisKey, 60*time.Second).Err()
	}
	if count > limit {
		return false, secondsUntilNextWindow()
	}
	return true, 0
}

func secondsUntilNextWindow() int {
	now := time.Now().UTC()
	next := now.Truncate(time.Minute).Add(time.Minute)
	secs := int(next.Sub(now).Seconds())
	if secs < 1 {
		return 1
	}
	return secs
}

func AgentKey(tenantID, agentID string) string {
	return tenantID + ":" + agentID
}

func RetryAfterHeader(seconds int) string {
	return strconv.Itoa(seconds)
}
