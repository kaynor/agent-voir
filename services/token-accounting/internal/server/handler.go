package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/agentvoir/agentvoir/services/token-accounting/internal/httputil"
	"github.com/agentvoir/agentvoir/services/token-accounting/internal/usage"
)

// Handler serves usage event ingestion HTTP endpoints.
type Handler struct {
	store usage.Store
}

func NewHandler(store usage.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/usage-events", h.ingestEvent)
	mux.HandleFunc("GET /v1/usage-events", h.listEvents)
	mux.HandleFunc("GET /v1/usage-events/summary", h.summarizeEvents)
}

func (h *Handler) ingestEvent(w http.ResponseWriter, r *http.Request) {
	var req usage.IngestRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	event, msg := req.Normalize()
	if msg != "" {
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	if err := h.store.Insert(r.Context(), event); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to ingest usage event")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, event)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	events, err := h.store.List(r.Context(), usage.ListFilter{
		TenantID: r.URL.Query().Get("tenant_id"),
		AgentID:  r.URL.Query().Get("agent_id"),
		Limit:    limit,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list usage events")
		return
	}
	if events == nil {
		events = []usage.Event{}
	}

	httputil.WriteJSON(w, http.StatusOK, events)
}

func (h *Handler) summarizeEvents(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	switch period {
	case "monthly":
		since = time.Now().UTC().AddDate(0, -1, 0)
	case "daily":
	default:
		httputil.WriteError(w, http.StatusBadRequest, "period must be daily or monthly")
		return
	}

	rollup, err := h.store.Summary(r.Context(), usage.SummaryFilter{
		Period:   period,
		AgentID:  r.URL.Query().Get("agent_id"),
		TenantID: r.URL.Query().Get("tenant_id"),
		Since:    since,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to summarize usage events")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rollup)
}

// OpenStore connects to ClickHouse when configured, otherwise uses memory storage.
func OpenStore(ctx context.Context, clickhouseDSN string) (usage.Store, error) {
	if clickhouseDSN == "" {
		return usage.NewMemoryStore(), nil
	}
	return usage.NewClickHouseStore(clickhouseDSN)
}
