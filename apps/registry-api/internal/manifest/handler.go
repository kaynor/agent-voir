package manifest

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

const maxManifestFetchBytes = 1 << 20

// Handler exposes manifest parse and register endpoints.
type Handler struct {
	stores     Stores
	httpClient *http.Client
}

func NewHandler(stores Stores) *Handler {
	return &Handler{
		stores: stores,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/agents/parse-manifest", h.parseManifest)
	mux.HandleFunc("POST /v1/agents/from-manifest", h.registerFromManifest)
	mux.HandleFunc("POST /v1/agents/from-manifest-url", h.registerFromManifestURL)
}

func (h *Handler) parseManifest(w http.ResponseWriter, r *http.Request) {
	data, err := readManifestBody(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	doc, err := Parse(data)
	if err != nil {
		writeManifestError(w, err)
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
	if err != nil {
		writeRegisterError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}

type manifestURLRequest struct {
	URL string `json:"url"`
}

func (h *Handler) registerFromManifestURL(w http.ResponseWriter, r *http.Request) {
	var req manifestURLRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "url is required")
		return
	}

	data, err := h.fetchManifest(req.URL)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := RegisterYAML(h.stores, data)
	if err != nil {
		writeRegisterError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) fetchManifest(rawURL string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, errors.New("url must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("url host is required")
	}

	resp, err := h.httpClient.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("failed to fetch manifest URL")
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxManifestFetchBytes))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("manifest URL returned empty body")
	}
	return data, nil
}

func readManifestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxManifestFetchBytes))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("manifest body is required")
	}
	return data, nil
}

func writeManifestError(w http.ResponseWriter, err error) {
	var validation ValidationErrors
	if errors.As(err, &validation) {
		httputil.WriteValidationErrors(w, http.StatusBadRequest, ErrInvalidManifest.Error(), validation.Issues)
		return
	}
	httputil.WriteError(w, http.StatusBadRequest, err.Error())
}

func writeRegisterError(w http.ResponseWriter, err error) {
	if errors.Is(err, agents.ErrConflict) {
		httputil.WriteError(w, http.StatusConflict, "agent already registered for agent_id, version, and environment")
		return
	}
	var validation ValidationErrors
	if errors.As(err, &validation) {
		httputil.WriteValidationErrors(w, http.StatusBadRequest, ErrInvalidManifest.Error(), validation.Issues)
		return
	}
	status := http.StatusInternalServerError
	if errors.Is(err, ErrInvalidManifest) {
		status = http.StatusBadRequest
	}
	httputil.WriteError(w, status, err.Error())
}
