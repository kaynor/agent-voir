package usage

import (
	"context"
	"sort"
	"sync"
)

// MemoryStore keeps usage events in process memory for local development and tests.
type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Insert(_ context.Context, event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *MemoryStore) List(_ context.Context, filter ListFilter) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	matches := make([]Event, 0, len(s.events))
	for _, event := range s.events {
		if filter.TenantID != "" && event.TenantID != filter.TenantID {
			continue
		}
		if filter.AgentID != "" && event.AgentID != filter.AgentID {
			continue
		}
		matches = append(matches, event)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].EventTime.After(matches[j].EventTime)
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}
