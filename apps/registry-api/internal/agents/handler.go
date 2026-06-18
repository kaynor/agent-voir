package agents

import (
	"errors"
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

// Handler serves agent registration HTTP endpoints.
type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agents", h.listAgents)
	mux.HandleFunc("POST /v1/agents", h.registerAgent)
	mux.HandleFunc("GET /v1/agents/{agentID}", h.getAgent)
	mux.HandleFunc("PUT /v1/agents/{agentID}", h.updateAgent)
	mux.HandleFunc("DELETE /v1/agents/{agentID}", h.deleteAgent)
}

func (h *Handler) listAgents(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) registerAgent(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.Validate(); msg != "" {
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	agent, err := h.store.Create(req)
	if errors.Is(err, ErrConflict) {
		httputil.WriteError(w, http.StatusConflict, "agent already registered for agent_id, version, and environment")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, agent)
}

func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}
	environment := r.URL.Query().Get("environment")
	if environment == "" {
		environment = "dev"
	}

	agent, found := h.store.Get(agentID, version, environment)
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "agent not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, agent)
}

func (h *Handler) updateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}
	environment := r.URL.Query().Get("environment")
	if environment == "" {
		environment = "dev"
	}

	var req UpdateRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	agent, err := h.store.Update(agentID, version, environment, req)
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update agent")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, agent)
}

func (h *Handler) deleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}
	environment := r.URL.Query().Get("environment")
	if environment == "" {
		environment = "dev"
	}

	if err := h.store.Delete(agentID, version, environment); errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "agent not found")
		return
	} else if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete agent")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
