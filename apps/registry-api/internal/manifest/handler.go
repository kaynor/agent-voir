package manifest

import (
	"errors"
	"io"
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

// Handler exposes manifest parse and register endpoints.
type Handler struct {
	stores Stores
}

func NewHandler(stores Stores) *Handler {
	return &Handler{stores: stores}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/agents/parse-manifest", h.parseManifest)
	mux.HandleFunc("POST /v1/agents/from-manifest", h.registerFromManifest)
}

func (h *Handler) parseManifest(w http.ResponseWriter, r *http.Request) {
	data, err := readManifestBody(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	doc, err := Parse(data)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) registerFromManifest(w http.ResponseWriter, r *http.Request) {
	data, err := readManifestBody(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := RegisterYAML(h.stores, data)
	if errors.Is(err, agents.ErrConflict) {
		httputil.WriteError(w, http.StatusConflict, "agent already registered for agent_id, version, and environment")
		return
	}
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidManifest) {
			status = http.StatusBadRequest
		}
		httputil.WriteError(w, status, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}

func readManifestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("manifest body is required")
	}
	return data, nil
}
