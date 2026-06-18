package dependencies

import (
	"errors"
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
	mux.HandleFunc("GET /v1/agents/{agentID}/dependencies", h.listDependencies)
	mux.HandleFunc("POST /v1/agents/{agentID}/dependencies", h.createDependency)
	mux.HandleFunc("DELETE /v1/agents/{agentID}/dependencies/{dependencyID}", h.deleteDependency)
	mux.HandleFunc("GET /v1/agents/{agentID}/dependency-graph", h.getDependencyGraph)
}

func (h *Handler) listDependencies(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, h.store.List(agentID, version))
}

func (h *Handler) createDependency(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	var req CreateRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.Validate(); msg != "" {
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	dep, err := h.store.Create(agentID, version, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create dependency")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dep)
}

func (h *Handler) deleteDependency(w http.ResponseWriter, r *http.Request) {
	dependencyID := r.PathValue("dependencyID")
	if err := h.store.Delete(dependencyID); errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "dependency not found")
		return
	} else if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete dependency")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getDependencyGraph(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, h.store.Graph(agentID, version))
}
