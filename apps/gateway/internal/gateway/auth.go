package gateway

import (
	"net/http"
	"strings"
)

func authorize(r *http.Request, apiKey string) bool {
	if apiKey == "" {
		return true
	}
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	return strings.TrimPrefix(header, "Bearer ") == apiKey
}
