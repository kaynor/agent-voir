package budgets

import (
	"math"
	"net/http"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/usageclient"
)

type Handler struct {
	store          Store
	usageClient    *usageclient.Client
}

func NewHandler(store Store, usageBaseURL string) *Handler {
	return &Handler{
		store:       store,
		usageClient: usageclient.NewClient(usageBaseURL),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agents/{agentID}/budget", h.getBudget)
	mux.HandleFunc("GET /v1/agents/{agentID}/budget/status", h.getBudgetStatus)
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

type BudgetStatus struct {
	AgentID             string  `json:"agent_id"`
	AgentVersion        string  `json:"agent_version"`
	MonthlyUSDLimit     float64 `json:"monthly_usd_limit"`
	MonthlyUSDUsed      float64 `json:"monthly_usd_used"`
	MonthlyUSDRemaining float64 `json:"monthly_usd_remaining"`
	UtilizationPercent  float64 `json:"utilization_percent"`
	RequestsPerMinute   int64   `json:"requests_per_minute,omitempty"`
}

func (h *Handler) getBudgetStatus(w http.ResponseWriter, r *http.Request) {
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

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	used := 0.0
	if h.usageClient != nil {
		summary, err := h.usageClient.GetMonthlySummary(r.Context(), tenantID, agentID)
		if err == nil {
			used = summary.CostUSD
		}
	}

	remaining := math.Max(0, budget.MonthlyUSD-used)
	utilization := 0.0
	if budget.MonthlyUSD > 0 {
		utilization = (used / budget.MonthlyUSD) * 100
	}

	httputil.WriteJSON(w, http.StatusOK, BudgetStatus{
		AgentID:             agentID,
		AgentVersion:        version,
		MonthlyUSDLimit:     budget.MonthlyUSD,
		MonthlyUSDUsed:      used,
		MonthlyUSDRemaining: remaining,
		UtilizationPercent:  utilization,
		RequestsPerMinute:   budget.RequestsPerMinute,
	})
}
