package server

import (
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/budgets"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/dependencies"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/modelroutes"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/prompts"
)

// Stores holds in-memory registry stores for local development.
type Stores struct {
	Agents       *agents.MemoryStore
	Prompts      *prompts.MemoryStore
	Dependencies *dependencies.MemoryStore
	Budgets      *budgets.MemoryStore
	ModelRoutes  *modelroutes.MemoryStore
}

// NewStores creates empty in-memory stores.
func NewStores() *Stores {
	return &Stores{
		Agents:       agents.NewMemoryStore(),
		Prompts:      prompts.NewMemoryStore(),
		Dependencies: dependencies.NewMemoryStore(),
		Budgets:      budgets.NewMemoryStore(),
		ModelRoutes:  modelroutes.NewMemoryStore(),
	}
}

// RegisterRoutes wires all registry API HTTP handlers.
func RegisterRoutes(mux *http.ServeMux, stores *Stores) {
	agents.NewHandler(stores.Agents).RegisterRoutes(mux)
	prompts.NewHandler(stores.Prompts).RegisterRoutes(mux)
	dependencies.NewHandler(stores.Dependencies).RegisterRoutes(mux)
	budgets.NewHandler(stores.Budgets).RegisterRoutes(mux)
	modelroutes.NewHandler(stores.ModelRoutes).RegisterRoutes(mux)
}
