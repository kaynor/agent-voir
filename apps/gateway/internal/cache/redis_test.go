package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/alicebob/miniredis/v2"
)

func TestRedisStoreExactCache(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	store, err := cache.NewRedisStore(mr.Addr())
	if err != nil {
		t.Fatalf("new redis store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	entry := cache.Entry{
		Key:        "abc123",
		Value:      []byte(`{"object":"chat.completion"}`),
		TTLSeconds: 60,
		AgentID:    "demo-agent",
	}

	if err := store.Set(ctx, entry); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := store.Get(ctx, entry.Key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if string(got.Value) != string(entry.Value) {
		t.Fatalf("value = %q", got.Value)
	}

	if err := store.Delete(ctx, entry.Key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	miss, err := store.Get(ctx, entry.Key)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if miss != nil {
		t.Fatal("expected cache miss after delete")
	}
}

func TestRedisStoreExpiresEntries(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	store, err := cache.NewRedisStore(mr.Addr())
	if err != nil {
		t.Fatalf("new redis store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	entry := cache.Entry{
		Key:        "ttl-key",
		Value:      []byte("value"),
		TTLSeconds: 1,
	}
	if err := store.Set(ctx, entry); err != nil {
		t.Fatalf("set: %v", err)
	}

	mr.FastForward(2 * time.Second)

	got, err := store.Get(ctx, entry.Key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatal("expected expired cache entry to miss")
	}
}
