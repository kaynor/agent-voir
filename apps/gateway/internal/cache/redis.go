package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultRedisKeyPrefix = "agentvoir:cache:exact:"

// RedisStore persists exact cache entries in Redis.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore connects to Redis and verifies connectivity with PING.
func NewRedisStore(addr string) (*RedisStore, error) {
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

	return &RedisStore{
		client: client,
		prefix: defaultRedisKeyPrefix,
	}, nil
}

func (s *RedisStore) redisKey(key string) string {
	return s.prefix + key
}

func (s *RedisStore) Get(ctx context.Context, key string) (*Entry, error) {
	value, err := s.client.Get(ctx, s.redisKey(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &Entry{
		Key:         key,
		Value:       value,
		CacheStatus: "hit",
	}, nil
}

func (s *RedisStore) Set(ctx context.Context, entry Entry) error {
	redisKey := s.redisKey(entry.Key)
	if entry.TTLSeconds > 0 {
		return s.client.Set(ctx, redisKey, entry.Value, time.Duration(entry.TTLSeconds)*time.Second).Err()
	}
	return s.client.Set(ctx, redisKey, entry.Value, 0).Err()
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, s.redisKey(key)).Err()
}

// Close releases the Redis client connection pool.
func (s *RedisStore) Close() error {
	return s.client.Close()
}
