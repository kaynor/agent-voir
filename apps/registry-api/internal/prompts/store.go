package prompts

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrConflict = errors.New("prompt already registered")
	ErrNotFound = errors.New("prompt not found")
)

type Store interface {
	List() []Prompt
	Get(promptID, version string) (Prompt, bool)
	Create(req RegisterRequest) (Prompt, error)
	Update(promptID, version string, req UpdateRequest) (Prompt, error)
	Delete(promptID, version string) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	prompts map[string]Prompt
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{prompts: make(map[string]Prompt)}
}

func storeKey(promptID, version string) string {
	return fmt.Sprintf("%s:%s", promptID, version)
}

func (s *MemoryStore) List() []Prompt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		out = append(out, prompt)
	}
	return out
}

func (s *MemoryStore) Get(promptID, version string) (Prompt, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prompt, ok := s.prompts[storeKey(promptID, version)]
	return prompt, ok
}

func (s *MemoryStore) Create(req RegisterRequest) (Prompt, error) {
	req.ApplyDefaults()
	key := storeKey(req.PromptID, req.Version)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.prompts[key]; exists {
		return Prompt{}, ErrConflict
	}

	prompt := Prompt{
		ID:             uuid.NewString(),
		PromptID:       req.PromptID,
		Name:           req.Name,
		Version:        req.Version,
		OwnerTeam:      req.OwnerTeam,
		Template:       req.Template,
		RiskLevel:      req.RiskLevel,
		ApprovedModels: append([]string(nil), req.ApprovedModels...),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.prompts[key] = prompt
	return prompt, nil
}

func (s *MemoryStore) Update(promptID, version string, req UpdateRequest) (Prompt, error) {
	key := storeKey(promptID, version)

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.prompts[key]
	if !ok {
		return Prompt{}, ErrNotFound
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.OwnerTeam != "" {
		existing.OwnerTeam = req.OwnerTeam
	}
	if req.Template != "" {
		existing.Template = req.Template
	}
	if req.RiskLevel != "" {
		existing.RiskLevel = req.RiskLevel
	}
	if req.ApprovedModels != nil {
		existing.ApprovedModels = append([]string(nil), req.ApprovedModels...)
	}
	existing.UpdatedAt = time.Now().UTC()
	s.prompts[key] = existing
	return existing, nil
}

func (s *MemoryStore) Delete(promptID, version string) error {
	key := storeKey(promptID, version)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.prompts[key]; !ok {
		return ErrNotFound
	}
	delete(s.prompts, key)
	return nil
}
