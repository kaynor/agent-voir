package postgres

import (
	"context"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/prompts"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PromptsStore persists prompts in PostgreSQL.
type PromptsStore struct {
	pool *pgxpool.Pool
}

func NewPromptsStore(pool *pgxpool.Pool) *PromptsStore {
	return &PromptsStore{pool: pool}
}

func (s *PromptsStore) List() []prompts.Prompt {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id::text, prompt_id, name, version, owner_team, template,
		       risk_level, approved_models, created_at, updated_at
		FROM prompts
		ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]prompts.Prompt, 0)
	for rows.Next() {
		var prompt prompts.Prompt
		if err := rows.Scan(
			&prompt.ID, &prompt.PromptID, &prompt.Name, &prompt.Version, &prompt.OwnerTeam,
			&prompt.Template, &prompt.RiskLevel, &prompt.ApprovedModels, &prompt.CreatedAt, &prompt.UpdatedAt,
		); err != nil {
			return nil
		}
		out = append(out, prompt)
	}
	return out
}

func (s *PromptsStore) Get(promptID, version string) (prompts.Prompt, bool) {
	var prompt prompts.Prompt
	err := s.pool.QueryRow(context.Background(), `
		SELECT id::text, prompt_id, name, version, owner_team, template,
		       risk_level, approved_models, created_at, updated_at
		FROM prompts
		WHERE prompt_id = $1 AND version = $2`,
		promptID, version,
	).Scan(
		&prompt.ID, &prompt.PromptID, &prompt.Name, &prompt.Version, &prompt.OwnerTeam,
		&prompt.Template, &prompt.RiskLevel, &prompt.ApprovedModels, &prompt.CreatedAt, &prompt.UpdatedAt,
	)
	if isNoRows(err) {
		return prompts.Prompt{}, false
	}
	if err != nil {
		return prompts.Prompt{}, false
	}
	return prompt, true
}

func (s *PromptsStore) Create(req prompts.RegisterRequest) (prompts.Prompt, error) {
	req.ApplyDefaults()

	var prompt prompts.Prompt
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO prompts (
			prompt_id, name, version, owner_team, template, risk_level, approved_models
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, prompt_id, name, version, owner_team, template,
		          risk_level, approved_models, created_at, updated_at`,
		req.PromptID, req.Name, req.Version, req.OwnerTeam, req.Template, req.RiskLevel, req.ApprovedModels,
	).Scan(
		&prompt.ID, &prompt.PromptID, &prompt.Name, &prompt.Version, &prompt.OwnerTeam,
		&prompt.Template, &prompt.RiskLevel, &prompt.ApprovedModels, &prompt.CreatedAt, &prompt.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return prompts.Prompt{}, prompts.ErrConflict
	}
	if err != nil {
		return prompts.Prompt{}, err
	}
	return prompt, nil
}

func (s *PromptsStore) Update(promptID, version string, req prompts.UpdateRequest) (prompts.Prompt, error) {
	existing, ok := s.Get(promptID, version)
	if !ok {
		return prompts.Prompt{}, prompts.ErrNotFound
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

	err := s.pool.QueryRow(context.Background(), `
		UPDATE prompts
		SET name = $3, owner_team = $4, template = $5, risk_level = $6,
		    approved_models = $7, updated_at = now()
		WHERE prompt_id = $1 AND version = $2
		RETURNING id::text, prompt_id, name, version, owner_team, template,
		          risk_level, approved_models, created_at, updated_at`,
		promptID, version,
		existing.Name, existing.OwnerTeam, existing.Template, existing.RiskLevel, existing.ApprovedModels,
	).Scan(
		&existing.ID, &existing.PromptID, &existing.Name, &existing.Version, &existing.OwnerTeam,
		&existing.Template, &existing.RiskLevel, &existing.ApprovedModels, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if isNoRows(err) {
		return prompts.Prompt{}, prompts.ErrNotFound
	}
	if err != nil {
		return prompts.Prompt{}, err
	}
	return existing, nil
}

func (s *PromptsStore) Delete(promptID, version string) error {
	tag, err := s.pool.Exec(context.Background(), `
		DELETE FROM prompts WHERE prompt_id = $1 AND version = $2`,
		promptID, version,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return prompts.ErrNotFound
	}
	return nil
}
