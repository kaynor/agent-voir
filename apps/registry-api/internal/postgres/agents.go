package postgres

import (
	"context"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentsStore persists agents in PostgreSQL.
type AgentsStore struct {
	pool *pgxpool.Pool
}

func NewAgentsStore(pool *pgxpool.Pool) *AgentsStore {
	return &AgentsStore{pool: pool}
}

func (s *AgentsStore) List() []agents.Agent {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id::text, agent_id, name, version, owner_team, cost_center, environment,
		       framework, risk_level, lifecycle, data_classes, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]agents.Agent, 0)
	for rows.Next() {
		var agent agents.Agent
		if err := rows.Scan(
			&agent.ID, &agent.AgentID, &agent.Name, &agent.Version, &agent.OwnerTeam,
			&agent.CostCenter, &agent.Environment, &agent.Framework, &agent.RiskLevel,
			&agent.Lifecycle, &agent.DataClasses, &agent.CreatedAt, &agent.UpdatedAt,
		); err != nil {
			return nil
		}
		out = append(out, agent)
	}
	return out
}

func (s *AgentsStore) Get(agentID, version, environment string) (agents.Agent, bool) {
	var agent agents.Agent
	err := s.pool.QueryRow(context.Background(), `
		SELECT id::text, agent_id, name, version, owner_team, cost_center, environment,
		       framework, risk_level, lifecycle, data_classes, created_at, updated_at
		FROM agents
		WHERE agent_id = $1 AND version = $2 AND environment = $3`,
		agentID, version, environment,
	).Scan(
		&agent.ID, &agent.AgentID, &agent.Name, &agent.Version, &agent.OwnerTeam,
		&agent.CostCenter, &agent.Environment, &agent.Framework, &agent.RiskLevel,
		&agent.Lifecycle, &agent.DataClasses, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if isNoRows(err) {
		return agents.Agent{}, false
	}
	if err != nil {
		return agents.Agent{}, false
	}
	return agent, true
}

func (s *AgentsStore) Create(req agents.RegisterRequest) (agents.Agent, error) {
	req.ApplyDefaults()

	var agent agents.Agent
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO agents (
			agent_id, name, version, owner_team, cost_center, environment,
			framework, risk_level, lifecycle, data_classes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text, agent_id, name, version, owner_team, cost_center, environment,
		          framework, risk_level, lifecycle, data_classes, created_at, updated_at`,
		req.AgentID, req.Name, req.Version, req.OwnerTeam, req.CostCenter, req.Environment,
		req.Framework, req.RiskLevel, req.Lifecycle, req.DataClasses,
	).Scan(
		&agent.ID, &agent.AgentID, &agent.Name, &agent.Version, &agent.OwnerTeam,
		&agent.CostCenter, &agent.Environment, &agent.Framework, &agent.RiskLevel,
		&agent.Lifecycle, &agent.DataClasses, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return agents.Agent{}, agents.ErrConflict
	}
	if err != nil {
		return agents.Agent{}, err
	}
	return agent, nil
}

func (s *AgentsStore) Update(agentID, version, environment string, req agents.UpdateRequest) (agents.Agent, error) {
	if environment == "" {
		environment = "dev"
	}

	existing, ok := s.Get(agentID, version, environment)
	if !ok {
		return agents.Agent{}, agents.ErrNotFound
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

	err := s.pool.QueryRow(context.Background(), `
		UPDATE agents
		SET name = $4, owner_team = $5, cost_center = $6, framework = $7,
		    risk_level = $8, lifecycle = $9, data_classes = $10, updated_at = now()
		WHERE agent_id = $1 AND version = $2 AND environment = $3
		RETURNING id::text, agent_id, name, version, owner_team, cost_center, environment,
		          framework, risk_level, lifecycle, data_classes, created_at, updated_at`,
		agentID, version, environment,
		existing.Name, existing.OwnerTeam, existing.CostCenter, existing.Framework,
		existing.RiskLevel, existing.Lifecycle, existing.DataClasses,
	).Scan(
		&existing.ID, &existing.AgentID, &existing.Name, &existing.Version, &existing.OwnerTeam,
		&existing.CostCenter, &existing.Environment, &existing.Framework, &existing.RiskLevel,
		&existing.Lifecycle, &existing.DataClasses, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if isNoRows(err) {
		return agents.Agent{}, agents.ErrNotFound
	}
	if err != nil {
		return agents.Agent{}, err
	}
	return existing, nil
}

func (s *AgentsStore) Delete(agentID, version, environment string) error {
	if environment == "" {
		environment = "dev"
	}

	tag, err := s.pool.Exec(context.Background(), `
		DELETE FROM agents
		WHERE agent_id = $1 AND version = $2 AND environment = $3`,
		agentID, version, environment,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return agents.ErrNotFound
	}
	return nil
}
