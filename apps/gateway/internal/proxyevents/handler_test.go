package proxyevents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListAndTrace(t *testing.T) {
	store := NewMemoryStore(1000)
	h := NewHandler(store)
	_, _ = store.InsertBatch(GenerateSeed(SeedOptions{Count: 9}))

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/proxy-events?limit=10", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}
	var list listResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Events) == 0 {
		t.Fatal("expected events")
	}
	traceID := list.Events[0].TraceID

	req = httptest.NewRequest(http.MethodGet, "/v1/traces/"+traceID, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("trace status = %d", rec.Code)
	}
	var detail TraceDetail
	if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
		t.Fatal(err)
	}
	if detail.TraceID != traceID {
		t.Fatalf("trace id mismatch: %s", detail.TraceID)
	}
}

func TestSeedEndpoint(t *testing.T) {
	store := NewMemoryStore(1000)
	h := NewHandler(store)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"count":12,"reset":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/proxy-events/seed", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("seed status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.Count() == 0 {
		t.Fatal("expected seeded events")
	}
}

func TestMetrics(t *testing.T) {
	store := NewMemoryStore(100)
	_, _ = store.Insert(Event{
		EventTime:    time.Now().UTC(),
		TraceID:      "trace_test",
		SpanID:       "span_1",
		AgentID:      "a1",
		StatusCode:   200,
		ResponseType: ResponseFinalAnswer,
		TokensIn:     100,
		TokensOut:    50,
		DurationMS:   500,
		CostUSD:      0.01,
		Terminal:     true,
	})
	metrics := store.Metrics(time.Now().UTC().Add(-5 * time.Minute))
	if metrics.RequestsTotal != 1 {
		t.Fatalf("requests = %d", metrics.RequestsTotal)
	}
}
