package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HTTPRecorder posts usage events to the token-accounting service.
type HTTPRecorder struct {
	url    string
	client *http.Client
}

func NewHTTPRecorder(url string) *HTTPRecorder {
	return &HTTPRecorder{
		url: url,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func NewRecorder(url string) Recorder {
	if url == "" {
		return NopRecorder{}
	}
	return NewHTTPRecorder(url)
}

func (r *HTTPRecorder) Record(event Event) {
	go r.post(event)
}

func (r *HTTPRecorder) post(event Event) {
	if event.EventTime.IsZero() {
		event.EventTime = time.Now().UTC()
	}
	if event.StatusCode == 0 {
		event.StatusCode = 200
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("usage event marshal failed: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url+"/v1/usage-events", bytes.NewReader(payload))
	if err != nil {
		log.Printf("usage event request failed: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("usage event ingest failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		log.Printf("usage event ingest returned %d", resp.StatusCode)
	}
}
