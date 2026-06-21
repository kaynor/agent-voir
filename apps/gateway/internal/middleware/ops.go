package middleware

import (
	"net/http"
	"strings"

	agentauth "github.com/agentvoir/agentvoir/packages/auth-go"
)

// SkipOpsDashboard allows unauthenticated read access to live proxy-events APIs
// so the browser console can connect without embedding GATEWAY_API_KEY client-side.
// POST /v1/proxy-events/seed still requires auth.
func SkipOpsDashboard(r *http.Request) bool {
	path := r.URL.Path
	if r.Method == http.MethodGet && (path == "/v1/proxy-events" || path == "/v1/proxy-events/metrics") {
		return true
	}
	if path == "/ws/proxy-events" {
		return true
	}
	if r.Method == http.MethodGet && strings.HasPrefix(path, "/v1/traces/") {
		return true
	}
	if path == "/metrics" {
		return true
	}
	return false
}

func skipAuth(r *http.Request) bool {
	return agentauth.SkipHealthz(r) || SkipOpsDashboard(r)
}
