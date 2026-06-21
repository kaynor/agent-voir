package middleware

import (
	"encoding/json"
	"net/http"

	agentauth "github.com/agentvoir/agentvoir/packages/auth-go"
)

// Auth builds gateway authentication middleware (OIDC JWT + GATEWAY_API_KEY hybrid).
// When OIDC is not configured, only GATEWAY_API_KEY is enforced (existing behavior).
func Auth(cfg agentauth.Config) func(http.Handler) http.Handler {
	authn := agentauth.NewAuthenticator(cfg)
	return agentauth.Middleware(authn, agentauth.MiddlewareOptions{
		Skip: skipAuth,
		OnError: func(w http.ResponseWriter, _ *http.Request, err error) {
			message := "invalid API key"
			if err == agentauth.ErrMissingBearer {
				message = "authorization bearer token required"
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": message,
					"type":    "invalid_request_error",
					"code":    "invalid_api_key",
				},
			})
		},
	})
}

// LoadAuthConfig reads gateway auth settings from the environment.
func LoadAuthConfig() agentauth.Config {
	return agentauth.LoadConfigFromEnv("GATEWAY_API_KEY")
}
