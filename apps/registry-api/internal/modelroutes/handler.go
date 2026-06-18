package modelroutes

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
	mux.HandleFunc("GET /v1/agents/{agentID}/model-route", h.getModelRoute)
	mux.HandleFunc("PUT /v1/agents/{agentID}/model-route", h.upsertModelRoute)
}

func (h *Handler) getModelRoute(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	route, found := h.store.Get(agentID, version)
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "model route not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, route)
}

func (h *Handler) upsertModelRoute(w http.ResponseWriter, r *http.Request) {
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

	route, err := h.store.Upsert(agentID, version, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, route)
}
