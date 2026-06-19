package postgres

import (
	"context"
	"encoding/json"
	"fmt"

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

const agentSelectColumns = `
	id::text, agent_id, name, version, owner_team, cost_center, environment,
	framework, risk_level, lifecycle, cache_mode, cache_ttl_seconds, semantic_cache_allowed,
	policies, data_classes, created_at, updated_at`

func scanAgent(row interface {
	Scan(dest ...any) error
}) (agents.Agent, error) {
	var agent agents.Agent
	var policiesJSON []byte
	err := row.Scan(
		&agent.ID, &agent.AgentID, &agent.Name, &agent.Version, &agent.OwnerTeam,
		&agent.CostCenter, &agent.Environment, &agent.Framework, &agent.RiskLevel,
		&agent.Lifecycle, &agent.CacheMode, &agent.CacheTTLSeconds, &agent.SemanticCacheAllowed,
		&policiesJSON, &agent.DataClasses, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return agents.Agent{}, err
	}
	_ = json.Unmarshal(policiesJSON, &agent.Policies)
	if agent.Policies.AllowedProviders == nil {
		agent.Policies.AllowedProviders = []string{}
	}
	return agent, nil
}

func (s *AgentsStore) List(opts agents.ListOptions) agents.ListResult {
	where := ""
	args := []any{}
	if opts.Environment != "" {
		where = "WHERE environment = $1"
		args = append(args, opts.Environment)
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM agents " + where
	if err := s.pool.QueryRow(context.Background(), countSQL, args...).Scan(&total); err != nil {
		return agents.ListResult{Items: []agents.Agent{}, Limit: opts.Limit, Offset: opts.Offset}
	}

	sortColumn := "created_at"
	switch opts.SortBy {
	case "updated_at":
		sortColumn = "updated_at"
	case "agent_id":
		sortColumn = "agent_id"
	case "name":
		sortColumn = "name"
	}
	sortOrder := "DESC"
	if opts.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	query := fmt.Sprintf(`
		SELECT %s
		FROM agents
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, agentSelectColumns, where, sortColumn, sortOrder, limitArg, offsetArg)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return agents.ListResult{Items: []agents.Agent{}, Total: total, Limit: opts.Limit, Offset: opts.Offset}
	}
	defer rows.Close()

	out := make([]agents.Agent, 0)
	for rows.Next() {
		agent, err := scanAgent(rows)
		if err != nil {
			return agents.ListResult{Items: []agents.Agent{}, Total: total, Limit: opts.Limit, Offset: opts.Offset}
		}
		out = append(out, agent)
	}
	return agents.ListResult{
		Items:  out,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}
}

func (s *AgentsStore) Get(agentID, version, environment string) (agents.Agent, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT `+agentSelectColumns+`
		FROM agents
		WHERE agent_id = $1 AND version = $2 AND environment = $3`,
		agentID, version, environment,
	)
	agent, err := scanAgent(row)
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
	policiesJSON, err := json.Marshal(req.Policies)
	if err != nil {
		return agents.Agent{}, err
	}

	row := s.pool.QueryRow(context.Background(), `
		INSERT INTO agents (
			agent_id, name, version, owner_team, cost_center, environment,
			framework, risk_level, lifecycle, cache_mode, cache_ttl_seconds, semantic_cache_allowed,
			policies, data_classes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+agentSelectColumns,
		req.AgentID, req.Name, req.Version, req.OwnerTeam, req.CostCenter, req.Environment,
		req.Framework, req.RiskLevel, req.Lifecycle, req.CacheMode, req.CacheTTLSeconds, req.SemanticCacheAllowed,
		policiesJSON, req.DataClasses,
	)
	agent, err := scanAgent(row)
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
	if req.CacheMode != "" {
		existing.CacheMode = req.CacheMode
	}
	if req.CacheTTLSeconds > 0 {
		existing.CacheTTLSeconds = req.CacheTTLSeconds
	}
	if req.SemanticCacheAllowed != nil {
		existing.SemanticCacheAllowed = *req.SemanticCacheAllowed
	}
	if req.Policies != nil {
		existing.Policies = *req.Policies
	}
	if req.DataClasses != nil {
		existing.DataClasses = append([]string(nil), req.DataClasses...)
	}

	policiesJSON, err := json.Marshal(existing.Policies)
	if err != nil {
		return agents.Agent{}, err
	}

	row := s.pool.QueryRow(context.Background(), `
		UPDATE agents
		SET name = $4, owner_team = $5, cost_center = $6, framework = $7,
		    risk_level = $8, lifecycle = $9, cache_mode = $10, cache_ttl_seconds = $11,
		    semantic_cache_allowed = $12, policies = $13, data_classes = $14, updated_at = now()
		WHERE agent_id = $1 AND version = $2 AND environment = $3
		RETURNING `+agentSelectColumns,
		agentID, version, environment,
		existing.Name, existing.OwnerTeam, existing.CostCenter, existing.Framework,
		existing.RiskLevel, existing.Lifecycle, existing.CacheMode, existing.CacheTTLSeconds,
		existing.SemanticCacheAllowed, policiesJSON, existing.DataClasses,
	)
	updated, err := scanAgent(row)
	if isNoRows(err) {
		return agents.Agent{}, agents.ErrNotFound
	}
	if err != nil {
		return agents.Agent{}, err
	}
	return updated, nil
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
