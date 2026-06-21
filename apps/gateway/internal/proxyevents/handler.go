package proxyevents

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Handler serves proxy-events REST endpoints for the operations dashboard.
type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/proxy-events", h.listEvents)
	mux.HandleFunc("GET /v1/proxy-events/metrics", h.getMetrics)
	mux.HandleFunc("POST /v1/proxy-events/seed", h.seedEvents)
	mux.HandleFunc("GET /v1/traces/{traceID}", h.getTrace)
	mux.HandleFunc("GET /ws/proxy-events", h.wsProxyEvents)
}

type listResponse struct {
	Window        string  `json:"window"`
	Limit         int     `json:"limit"`
	MatchedCount  int     `json:"matched_count"`
	ReturnedCount int     `json:"returned_count"`
	Metrics       MetricsSnapshot `json:"metrics"`
	Events        []Event `json:"events"`
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	filter := parseListFilter(r)
	since := filter.Since
	if since.IsZero() {
		since = time.Now().UTC().Add(-5 * time.Minute)
		filter.Since = since
	}
	events, matched := h.store.List(filter)
	writeJSON(w, http.StatusOK, listResponse{
		Window:        "last_5m",
		Limit:         filter.Limit,
		MatchedCount:  matched,
		ReturnedCount: len(events),
		Metrics:       h.store.Metrics(since),
		Events:        events,
	})
}

func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	since := time.Now().UTC().Add(-5 * time.Minute)
	switch window {
	case "1m":
		since = time.Now().UTC().Add(-1 * time.Minute)
	case "15m":
		since = time.Now().UTC().Add(-15 * time.Minute)
	case "1h":
		since = time.Now().UTC().Add(-1 * time.Hour)
	}
	writeJSON(w, http.StatusOK, h.store.Metrics(since))
}

type seedRequest struct {
	Count  int  `json:"count"`
	Reset  bool `json:"reset"`
	Stream bool `json:"stream"`
}

type seedResponse struct {
	Inserted int `json:"inserted"`
	Total    int `json:"total"`
	Message  string `json:"message"`
}

func (h *Handler) seedEvents(w http.ResponseWriter, r *http.Request) {
	var req seedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if req.Count <= 0 {
		req.Count = 60
	}
	if req.Reset {
		h.store.Reset()
	}
	events := GenerateSeed(SeedOptions{Count: req.Count})
	inserted, err := h.store.InsertBatch(events)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, seedResponse{
		Inserted: len(inserted),
		Total:    h.store.Count(),
		Message:  "Dummy proxy events loaded for Live Proxy Flow dashboard",
	})
}

func (h *Handler) getTrace(w http.ResponseWriter, r *http.Request) {
	traceID := strings.TrimPrefix(r.URL.Path, "/v1/traces/")
	traceID = strings.Trim(traceID, "/")
	if traceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "trace id required"})
		return
	}
	detail, err := h.store.GetTrace(traceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if detail == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trace not found"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func parseListFilter(r *http.Request) ListFilter {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	filter := ListFilter{
		Limit:        limit,
		AgentID:      q.Get("agent"),
		ResponseType: ResponseType(q.Get("response_type")),
		Tag:          q.Get("tag"),
		TraceID:      q.Get("trace_id"),
	}
	if statusMin, err := strconv.Atoi(q.Get("status_min")); err == nil {
		filter.StatusMin = statusMin
	}
	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.Since = t
		}
	}
	if window := q.Get("window"); window != "" && filter.Since.IsZero() {
		filter.Since = time.Now().UTC().Add(parseWindow(window))
	}
	return filter
}

func parseWindow(window string) time.Duration {
	switch window {
	case "1m":
		return -1 * time.Minute
	case "15m":
		return -15 * time.Minute
	case "1h":
		return -1 * time.Hour
	default:
		return -5 * time.Minute
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
