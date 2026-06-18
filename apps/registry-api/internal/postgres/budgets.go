package postgres

import (
	"context"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/budgets"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BudgetsStore persists agent budgets in PostgreSQL.
type BudgetsStore struct {
	pool *pgxpool.Pool
}

func NewBudgetsStore(pool *pgxpool.Pool) *BudgetsStore {
	return &BudgetsStore{pool: pool}
}

func (s *BudgetsStore) Get(agentID, agentVersion string) (budgets.Budget, bool) {
	var budget budgets.Budget
	err := s.pool.QueryRow(context.Background(), `
		SELECT id::text, agent_id, agent_version, monthly_usd, max_prompt_tokens_per_request,
		       max_completion_tokens_per_request, requests_per_minute, created_at, updated_at
		FROM budgets
		WHERE agent_id = $1 AND agent_version = $2`,
		agentID, agentVersion,
	).Scan(
		&budget.ID, &budget.AgentID, &budget.AgentVersion, &budget.MonthlyUSD,
		&budget.MaxPromptTokensPerRequest, &budget.MaxCompletionTokensPerRequest,
		&budget.RequestsPerMinute, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if isNoRows(err) {
		return budgets.Budget{}, false
	}
	if err != nil {
		return budgets.Budget{}, false
	}
	return budget, true
}

func (s *BudgetsStore) Upsert(agentID, agentVersion string, req budgets.UpsertRequest) budgets.Budget {
	var budget budgets.Budget
	_ = s.pool.QueryRow(context.Background(), `
		INSERT INTO budgets (
			agent_id, agent_version, monthly_usd, max_prompt_tokens_per_request,
			max_completion_tokens_per_request, requests_per_minute
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (agent_id, agent_version) DO UPDATE SET
			monthly_usd = EXCLUDED.monthly_usd,
			max_prompt_tokens_per_request = EXCLUDED.max_prompt_tokens_per_request,
			max_completion_tokens_per_request = EXCLUDED.max_completion_tokens_per_request,
			requests_per_minute = EXCLUDED.requests_per_minute,
			updated_at = now()
		RETURNING id::text, agent_id, agent_version, monthly_usd, max_prompt_tokens_per_request,
		          max_completion_tokens_per_request, requests_per_minute, created_at, updated_at`,
		agentID, agentVersion, req.MonthlyUSD, req.MaxPromptTokensPerRequest,
		req.MaxCompletionTokensPerRequest, req.RequestsPerMinute,
	).Scan(
		&budget.ID, &budget.AgentID, &budget.AgentVersion, &budget.MonthlyUSD,
		&budget.MaxPromptTokensPerRequest, &budget.MaxCompletionTokensPerRequest,
		&budget.RequestsPerMinute, &budget.CreatedAt, &budget.UpdatedAt,
	)
	return budget
}
