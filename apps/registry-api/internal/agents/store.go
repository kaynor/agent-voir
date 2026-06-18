package agents

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrConflict = errors.New("agent already registered")
var ErrNotFound = errors.New("agent not found")

// Store persists registered agents.
type Store interface {
	List() []Agent
	Get(agentID, version, environment string) (Agent, bool)
	Create(req RegisterRequest) (Agent, error)
	Update(agentID, version, environment string, req UpdateRequest) (Agent, error)
	Delete(agentID, version, environment string) error
}

// MemoryStore is an in-process agent registry for local development and tests.
type MemoryStore struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{agents: make(map[string]Agent)}
}

func storeKey(agentID, version, environment string) string {
	return fmt.Sprintf("%s:%s:%s", agentID, version, environment)
}

func (s *MemoryStore) List() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		out = append(out, agent)
	}
	return out
}

func (s *MemoryStore) Get(agentID, version, environment string) (Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, ok := s.agents[storeKey(agentID, version, environment)]
	return agent, ok
}

func (s *MemoryStore) Create(req RegisterRequest) (Agent, error) {
	req.ApplyDefaults()

	key := storeKey(req.AgentID, req.Version, req.Environment)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[key]; exists {
		return Agent{}, ErrConflict
	}

	agent := Agent{
		ID:          uuid.NewString(),
		AgentID:     req.AgentID,
		Name:        req.Name,
		Version:     req.Version,
		OwnerTeam:   req.OwnerTeam,
		CostCenter:  req.CostCenter,
		Environment: req.Environment,
		Framework:   req.Framework,
		RiskLevel:   req.RiskLevel,
		Lifecycle:   req.Lifecycle,
		DataClasses: append([]string(nil), req.DataClasses...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.agents[key] = agent
	return agent, nil
}

func (s *MemoryStore) Update(agentID, version, environment string, req UpdateRequest) (Agent, error) {
	if environment == "" {
		environment = "dev"
	}

	key := storeKey(agentID, version, environment)

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.agents[key]
	if !ok {
		return Agent{}, ErrNotFound
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.OwnerTeam != "" {
		existing.OwnerTeam = req.OwnerTeam
	}
	if req.CostCenter != "" {
		existing.CostCenter = req.CostCenter
	}
	if req.Framework != "" {
		existing.Framework = req.Framework
	}
	if req.RiskLevel != "" {
		existing.RiskLevel = req.RiskLevel
	}
	if req.Lifecycle != "" {
		existing.Lifecycle = req.Lifecycle
	}
	if req.DataClasses != nil {
		existing.DataClasses = append([]string(nil), req.DataClasses...)
	}
	existing.UpdatedAt = time.Now().UTC()
	s.agents[key] = existing
	return existing, nil
}

func (s *MemoryStore) Delete(agentID, version, environment string) error {
	if environment == "" {
		environment = "dev"
	}

	key := storeKey(agentID, version, environment)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.agents[key]; !ok {
		return ErrNotFound
	}
	delete(s.agents, key)
	return nil
}
