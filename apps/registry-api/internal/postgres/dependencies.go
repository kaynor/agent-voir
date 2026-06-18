package postgres

import (
	"context"
	"fmt"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/dependencies"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DependenciesStore persists agent dependencies in PostgreSQL.
type DependenciesStore struct {
	pool *pgxpool.Pool
}

func NewDependenciesStore(pool *pgxpool.Pool) *DependenciesStore {
	return &DependenciesStore{pool: pool}
}

func (s *DependenciesStore) List(agentID, agentVersion string) []dependencies.Dependency {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id::text, agent_id, agent_version, dependency_type, dependency_name,
		       dependency_version, required, created_at
		FROM agent_dependencies
		WHERE agent_id = $1 AND agent_version = $2
		ORDER BY created_at ASC`,
		agentID, agentVersion,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]dependencies.Dependency, 0)
	for rows.Next() {
		var dep dependencies.Dependency
		var depType string
		if err := rows.Scan(
			&dep.ID, &dep.AgentID, &dep.AgentVersion, &depType, &dep.DependencyName,
			&dep.DependencyVersion, &dep.Required, &dep.CreatedAt,
		); err != nil {
			return nil
		}
		dep.DependencyType = dependencies.Type(depType)
		out = append(out, dep)
	}
	return out
}

func (s *DependenciesStore) Create(agentID, agentVersion string, req dependencies.CreateRequest) (dependencies.Dependency, error) {
	var dep dependencies.Dependency
	var depType string
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO agent_dependencies (
			agent_id, agent_version, dependency_type, dependency_name, dependency_version, required
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, agent_id, agent_version, dependency_type, dependency_name,
		          dependency_version, required, created_at`,
		agentID, agentVersion, string(req.DependencyType), req.DependencyName,
		req.DependencyVersion, req.RequiredValue(),
	).Scan(
		&dep.ID, &dep.AgentID, &dep.AgentVersion, &depType, &dep.DependencyName,
		&dep.DependencyVersion, &dep.Required, &dep.CreatedAt,
	)
	if err != nil {
		return dependencies.Dependency{}, err
	}
	dep.DependencyType = dependencies.Type(depType)
	return dep, nil
}

func (s *DependenciesStore) Delete(id string) error {
	tag, err := s.pool.Exec(context.Background(), `
		DELETE FROM agent_dependencies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return dependencies.ErrNotFound
	}
	return nil
}

func (s *DependenciesStore) Graph(agentID, agentVersion string) dependencies.Graph {
	deps := s.List(agentID, agentVersion)
	agentNodeID := fmt.Sprintf("agent:%s:%s", agentID, agentVersion)

	nodes := []dependencies.GraphNode{{
		ID:   agentNodeID,
		Type: dependencies.TypeAgent,
		Name: agentID,
	}}
	edges := make([]dependencies.GraphEdge, 0, len(deps))

	for _, dep := range deps {
		nodeID := fmt.Sprintf("%s:%s", dep.DependencyType, dep.DependencyName)
		nodes = append(nodes, dependencies.GraphNode{
			ID:   nodeID,
			Type: dep.DependencyType,
			Name: dep.DependencyName,
		})
		edges = append(edges, dependencies.GraphEdge{
			From:     agentNodeID,
			To:       nodeID,
			Required: dep.Required,
		})
	}

	return dependencies.Graph{
		AgentID:      agentID,
		AgentVersion: agentVersion,
		Nodes:        nodes,
		Edges:        edges,
	}
}
