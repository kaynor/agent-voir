package server

import (
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/budgets"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/dependencies"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/manifest"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/modelroutes"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/postgres"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/prompts"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Stores holds registry persistence backends.
type Stores struct {
	Agents       agents.Store
	Prompts      prompts.Store
	Dependencies dependencies.Store
	Budgets      budgets.Store
	ModelRoutes  modelroutes.Store
}

// NewMemoryStores creates in-memory stores for local development and tests.
func NewMemoryStores() *Stores {
	return &Stores{
		Agents:       agents.NewMemoryStore(),
		Prompts:      prompts.NewMemoryStore(),
		Dependencies: dependencies.NewMemoryStore(),
		Budgets:      budgets.NewMemoryStore(),
		ModelRoutes:  modelroutes.NewMemoryStore(),
	}
}

// NewPostgresStores creates PostgreSQL-backed stores.
func NewPostgresStores(pool *pgxpool.Pool) *Stores {
	return &Stores{
		Agents:       postgres.NewAgentsStore(pool),
		Prompts:      postgres.NewPromptsStore(pool),
		Dependencies: postgres.NewDependenciesStore(pool),
		Budgets:      postgres.NewBudgetsStore(pool),
		ModelRoutes:  postgres.NewModelRoutesStore(pool),
	}
}

// NewStores is an alias for NewMemoryStores.
func NewStores() *Stores {
	return NewMemoryStores()
}

// RegisterRoutes wires all registry API HTTP handlers.
func RegisterRoutes(mux *http.ServeMux, stores *Stores) {
	agents.NewHandler(stores.Agents).RegisterRoutes(mux)
	prompts.NewHandler(stores.Prompts).RegisterRoutes(mux)
	dependencies.NewHandler(stores.Dependencies).RegisterRoutes(mux)
	budgets.NewHandler(stores.Budgets).RegisterRoutes(mux)
	modelroutes.NewHandler(stores.ModelRoutes).RegisterRoutes(mux)
	manifest.NewHandler(manifest.Stores{
		Agents:       stores.Agents,
		Dependencies: stores.Dependencies,
		Budgets:      stores.Budgets,
		ModelRoutes:  stores.ModelRoutes,
	}).RegisterRoutes(mux)
}
