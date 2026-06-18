package dependencies

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("dependency not found")

type Store interface {
	List(agentID, agentVersion string) []Dependency
	Create(agentID, agentVersion string, req CreateRequest) (Dependency, error)
	Delete(id string) error
	Graph(agentID, agentVersion string) Graph
}

type MemoryStore struct {
	mu           sync.RWMutex
	dependencies map[string]Dependency
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{dependencies: make(map[string]Dependency)}
}

func (s *MemoryStore) List(agentID, agentVersion string) []Dependency {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Dependency, 0)
	for _, dep := range s.dependencies {
		if dep.AgentID == agentID && dep.AgentVersion == agentVersion {
			out = append(out, dep)
		}
	}
	return out
}

func (s *MemoryStore) Create(agentID, agentVersion string, req CreateRequest) (Dependency, error) {
	dep := Dependency{
		ID:                uuid.NewString(),
		AgentID:           agentID,
		AgentVersion:      agentVersion,
		DependencyType:    req.DependencyType,
		DependencyName:    req.DependencyName,
		DependencyVersion: req.DependencyVersion,
		Required:          req.RequiredValue(),
		CreatedAt:         time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.dependencies[dep.ID] = dep
	return dep, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.dependencies[id]; !ok {
		return ErrNotFound
	}
	delete(s.dependencies, id)
	return nil
}

func (s *MemoryStore) Graph(agentID, agentVersion string) Graph {
	deps := s.List(agentID, agentVersion)
	agentNodeID := fmt.Sprintf("agent:%s:%s", agentID, agentVersion)

	nodes := []GraphNode{{
		ID:   agentNodeID,
		Type: TypeAgent,
		Name: agentID,
	}}
	edges := make([]GraphEdge, 0, len(deps))

	for _, dep := range deps {
		nodeID := fmt.Sprintf("%s:%s", dep.DependencyType, dep.DependencyName)
		nodes = append(nodes, GraphNode{
			ID:   nodeID,
			Type: dep.DependencyType,
			Name: dep.DependencyName,
		})
		edges = append(edges, GraphEdge{
			From:     agentNodeID,
			To:       nodeID,
			Required: dep.Required,
		})
	}

	return Graph{
		AgentID:      agentID,
		AgentVersion: agentVersion,
		Nodes:        nodes,
		Edges:        edges,
	}
}
