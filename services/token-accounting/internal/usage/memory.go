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

func (s *MemoryStore) Summary(_ context.Context, filter SummaryFilter) (SummaryRollup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rollup := SummaryRollup{
		Period:   filter.Period,
		AgentID:  filter.AgentID,
		TenantID: filter.TenantID,
	}
	var hits int
	for _, event := range s.events {
		if event.EventTime.Before(filter.Since) {
			continue
		}
		if filter.TenantID != "" && event.TenantID != filter.TenantID {
			continue
		}
		if filter.AgentID != "" && event.AgentID != filter.AgentID {
			continue
		}
		rollup.EventCount++
		rollup.PromptTokens += event.PromptTokens
		rollup.CompletionTokens += event.CompletionTokens
		rollup.CostUSD += event.CostUSD
		if event.CacheStatus == "hit" {
			hits++
		}
	}
	if rollup.EventCount > 0 {
		rollup.CacheHitRate = float64(hits) / float64(rollup.EventCount)
	}
	return rollup, nil
}
