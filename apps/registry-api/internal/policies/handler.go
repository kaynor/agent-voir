package policies

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

// Handler exposes policy simulation against OPA.
type Handler struct {
	opaURL string
	client *http.Client
}

func NewHandler(opaURL string) *Handler {
	return &Handler{
		opaURL: strings.TrimRight(opaURL, "/"),
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	if h == nil || h.opaURL == "" {
		return
	}
	mux.HandleFunc("POST /v1/policies/simulate", h.simulate)
}

type SimulateRequest struct {
	Agent       map[string]any `json:"agent"`
	Request     map[string]any `json:"request"`
	Environment string         `json:"environment"`
}

type SimulateResponse struct {
	Allowed bool     `json:"allowed"`
	Reason  string   `json:"reason,omitempty"`
	Deny    []string `json:"deny,omitempty"`
}

func (h *Handler) simulate(w http.ResponseWriter, r *http.Request) {
	var req SimulateRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Environment == "" {
		req.Environment = "staging"
	}

	payload := map[string]any{
		"input": map[string]any{
			"agent":       req.Agent,
			"request":     req.Request,
			"environment": req.Environment,
		},
	}
	body, _ := json.Marshal(payload)

	allowReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.opaURL+"/v1/data/agentvoir/authz/allow", bytes.NewReader(body))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "policy simulation failed")
		return
	}
	allowReq.Header.Set("Content-Type", "application/json")

	allowResp, err := h.client.Do(allowReq)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "opa unavailable")
		return
	}
	defer allowResp.Body.Close()

	var allowResult struct {
		Result bool `json:"result"`
	}
	_ = json.NewDecoder(allowResp.Body).Decode(&allowResult)

	denyReasons := h.fetchDenyReasons(r.Context(), body)
	resp := SimulateResponse{Allowed: allowResult.Result}
	if !allowResult.Result {
		if len(denyReasons) > 0 {
			resp.Deny = denyReasons
			resp.Reason = denyReasons[0]
		} else {
			resp.Reason = "request denied by policy"
		}
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) fetchDenyReasons(ctx context.Context, body []byte) []string {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.opaURL+"/v1/data/agentvoir/authz/deny", bytes.NewReader(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Result []string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	return result.Result
}
