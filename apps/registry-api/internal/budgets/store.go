package budgets

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Get(agentID, agentVersion string) (Budget, bool)
	Upsert(agentID, agentVersion string, req UpsertRequest) Budget
}

type MemoryStore struct {
	mu      sync.RWMutex
	budgets map[string]Budget
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{budgets: make(map[string]Budget)}
}

func storeKey(agentID, agentVersion string) string {
	return fmt.Sprintf("%s:%s", agentID, agentVersion)
}

func (s *MemoryStore) Get(agentID, agentVersion string) (Budget, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	budget, ok := s.budgets[storeKey(agentID, agentVersion)]
	return budget, ok
}

func (s *MemoryStore) Upsert(agentID, agentVersion string, req UpsertRequest) Budget {
	key := storeKey(agentID, agentVersion)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.budgets[key]
	if !ok {
		existing = Budget{
			ID:           uuid.NewString(),
			AgentID:      agentID,
			AgentVersion: agentVersion,
			CreatedAt:    now,
		}
	}

	existing.MonthlyUSD = req.MonthlyUSD
	existing.MaxPromptTokensPerRequest = req.MaxPromptTokensPerRequest
	existing.MaxCompletionTokensPerRequest = req.MaxCompletionTokensPerRequest
	existing.RequestsPerMinute = req.RequestsPerMinute
	existing.UpdatedAt = now
	s.budgets[key] = existing
	return existing
}
