package prompts

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
	mux.HandleFunc("GET /v1/prompts", h.listPrompts)
	mux.HandleFunc("POST /v1/prompts", h.registerPrompt)
	mux.HandleFunc("GET /v1/prompts/{promptID}", h.getPrompt)
	mux.HandleFunc("PUT /v1/prompts/{promptID}", h.updatePrompt)
	mux.HandleFunc("DELETE /v1/prompts/{promptID}", h.deletePrompt)
}

func (h *Handler) listPrompts(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) registerPrompt(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.Validate(); msg != "" {
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	prompt, err := h.store.Create(req)
	if errors.Is(err, ErrConflict) {
		httputil.WriteError(w, http.StatusConflict, "prompt already registered for prompt_id and version")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to register prompt")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, prompt)
}

func (h *Handler) getPrompt(w http.ResponseWriter, r *http.Request) {
	promptID := r.PathValue("promptID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	prompt, found := h.store.Get(promptID, version)
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "prompt not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, prompt)
}

func (h *Handler) updatePrompt(w http.ResponseWriter, r *http.Request) {
	promptID := r.PathValue("promptID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	var req UpdateRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	prompt, err := h.store.Update(promptID, version, req)
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "prompt not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update prompt")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, prompt)
}

func (h *Handler) deletePrompt(w http.ResponseWriter, r *http.Request) {
	promptID := r.PathValue("promptID")
	version, ok := httputil.RequiredQuery(r, "version")
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "version query parameter is required")
		return
	}

	if err := h.store.Delete(promptID, version); errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "prompt not found")
		return
	} else if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete prompt")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
