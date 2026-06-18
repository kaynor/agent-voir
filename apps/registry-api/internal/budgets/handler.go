package budgets

import (
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agents/{agentID}/budget", h.getBudget)
	mux.HandleFunc("PUT /v1/agents/{agentID}/budget", h.upsertBudget)
}

func (h *Handler) getBudget(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	budget, found := h.store.Get(agentID, version)
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "budget not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, budget)
}

func (h *Handler) upsertBudget(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	var req UpsertRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	budget := h.store.Upsert(agentID, version, req)
	httputil.WriteJSON(w, http.StatusOK, budget)
}
