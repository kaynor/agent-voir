package postgres

import (
	"context"
	"fmt"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/modelroutes"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ModelRoutesStore persists model routes in PostgreSQL.
type ModelRoutesStore struct {
	pool *pgxpool.Pool
}

func NewModelRoutesStore(pool *pgxpool.Pool) *ModelRoutesStore {
	return &ModelRoutesStore{pool: pool}
}

func (s *ModelRoutesStore) Get(agentID, agentVersion string) (modelroutes.ModelRoute, bool) {
	var route modelroutes.ModelRoute
	err := s.pool.QueryRow(context.Background(), `
		SELECT id::text, agent_id, agent_version, primary_provider, primary_model,
		       fallback_provider, fallback_model, routing_policy, created_at, updated_at
		FROM model_routes
		WHERE agent_id = $1 AND agent_version = $2`,
		agentID, agentVersion,
	).Scan(
		&route.ID, &route.AgentID, &route.AgentVersion, &route.PrimaryProvider, &route.PrimaryModel,
		&route.FallbackProvider, &route.FallbackModel, &route.RoutingPolicy,
		&route.CreatedAt, &route.UpdatedAt,
	)
	if isNoRows(err) {
		return modelroutes.ModelRoute{}, false
	}
	if err != nil {
		return modelroutes.ModelRoute{}, false
	}
	return route, true
}

func (s *ModelRoutesStore) Upsert(agentID, agentVersion string, req modelroutes.UpsertRequest) (modelroutes.ModelRoute, error) {
	req.ApplyDefaults()
	if msg := req.Validate(); msg != "" {
		return modelroutes.ModelRoute{}, fmt.Errorf("%s", msg)
	}

	var route modelroutes.ModelRoute
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO model_routes (
			agent_id, agent_version, primary_provider, primary_model,
			fallback_provider, fallback_model, routing_policy
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (agent_id, agent_version) DO UPDATE SET
			primary_provider = EXCLUDED.primary_provider,
			primary_model = EXCLUDED.primary_model,
			fallback_provider = EXCLUDED.fallback_provider,
			fallback_model = EXCLUDED.fallback_model,
			routing_policy = EXCLUDED.routing_policy,
			updated_at = now()
		RETURNING id::text, agent_id, agent_version, primary_provider, primary_model,
		          fallback_provider, fallback_model, routing_policy, created_at, updated_at`,
		agentID, agentVersion, req.PrimaryProvider, req.PrimaryModel,
		req.FallbackProvider, req.FallbackModel, req.RoutingPolicy,
	).Scan(
		&route.ID, &route.AgentID, &route.AgentVersion, &route.PrimaryProvider, &route.PrimaryModel,
		&route.FallbackProvider, &route.FallbackModel, &route.RoutingPolicy,
		&route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return modelroutes.ModelRoute{}, err
	}
	return route, nil
}
