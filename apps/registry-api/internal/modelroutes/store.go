package modelroutes

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Get(agentID, agentVersion string) (ModelRoute, bool)
	Upsert(agentID, agentVersion string, req UpsertRequest) (ModelRoute, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	routes map[string]ModelRoute
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{routes: make(map[string]ModelRoute)}
}

func storeKey(agentID, agentVersion string) string {
	return fmt.Sprintf("%s:%s", agentID, agentVersion)
}

func (s *MemoryStore) Get(agentID, agentVersion string) (ModelRoute, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	route, ok := s.routes[storeKey(agentID, agentVersion)]
	return route, ok
}

func (s *MemoryStore) Upsert(agentID, agentVersion string, req UpsertRequest) (ModelRoute, error) {
	req.ApplyDefaults()
	if msg := req.Validate(); msg != "" {
		return ModelRoute{}, fmt.Errorf("%s", msg)
	}

	key := storeKey(agentID, agentVersion)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.routes[key]
	if !ok {
		existing = ModelRoute{
			ID:           uuid.NewString(),
			AgentID:      agentID,
			AgentVersion: agentVersion,
			CreatedAt:    now,
		}
	}

	existing.PrimaryProvider = req.PrimaryProvider
	existing.PrimaryModel = req.PrimaryModel
	existing.FallbackProvider = req.FallbackProvider
	existing.FallbackModel = req.FallbackModel
	existing.RoutingPolicy = req.RoutingPolicy
	existing.UpdatedAt = now
	s.routes[key] = existing
	return existing, nil
}
