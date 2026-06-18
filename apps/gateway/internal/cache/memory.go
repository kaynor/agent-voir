package cache

import (
	"context"
	"sync"
	"time"
)

type memoryEntry struct {
	entry   Entry
	expires time.Time
}

// MemoryStore is an in-process exact cache for local development.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]memoryEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: make(map[string]memoryEntry)}
}

func (s *MemoryStore) Get(_ context.Context, key string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.entries[key]
	if !ok {
		return nil, nil
	}
	if !item.expires.IsZero() && time.Now().After(item.expires) {
		return nil, nil
	}
	copy := item.entry
	return &copy, nil
}

func (s *MemoryStore) Set(_ context.Context, entry Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	expires := time.Time{}
	if entry.TTLSeconds > 0 {
		expires = time.Now().Add(time.Duration(entry.TTLSeconds) * time.Second)
	}
	s.entries[entry.Key] = memoryEntry{entry: entry, expires: expires}
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
	return nil
}
