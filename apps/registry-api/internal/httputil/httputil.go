package httputil

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func WriteValidationErrors(w http.ResponseWriter, status int, message string, issues any) {
	WriteJSON(w, status, map[string]any{
		"error":  message,
		"issues": issues,
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func RequiredQuery(r *http.Request, key string) (string, bool) {
	value := r.URL.Query().Get(key)
	return value, value != ""
}
