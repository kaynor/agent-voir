package cache

import (
	"context"
	"fmt"
)

// NewStore returns Redis when addr is set, otherwise an in-memory store.
func NewStore(ctx context.Context, redisAddr string) (Store, error) {
	if redisAddr != "" {
		store, err := NewRedisStore(redisAddr)
		if err != nil {
			return nil, fmt.Errorf("create redis cache: %w", err)
		}
		return store, nil
	}
	return NewMemoryStore(), nil
}
