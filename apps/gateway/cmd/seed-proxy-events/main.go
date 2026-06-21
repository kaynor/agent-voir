// Command seed-proxy-events POSTs dummy Live Proxy Flow rows to a running gateway.
//
// Usage:
//
//	go run ./cmd/seed-proxy-events -url http://localhost:8080 -count 80 -reset
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", envOr("GATEWAY_URL", "http://localhost:8080"), "gateway base URL")
	key := flag.String("key", envOr("GATEWAY_API_KEY", "agentvoir-onebox-key"), "gateway API key")
	count := flag.Int("count", 80, "approximate number of traces to generate")
	reset := flag.Bool("reset", true, "clear existing in-memory proxy events first")
	flag.Parse()

	body, _ := json.Marshal(map[string]any{
		"count":  *count,
		"reset":  *reset,
		"stream": false,
	})
	req, err := http.NewRequest(http.MethodPost, *url+"/v1/proxy-events/seed", bytes.NewReader(body))
	if err != nil {
		exitErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*key)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		exitErr(err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		exitErr(fmt.Errorf("seed failed (%d): %s", resp.StatusCode, string(payload)))
	}
	fmt.Println(string(payload))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
