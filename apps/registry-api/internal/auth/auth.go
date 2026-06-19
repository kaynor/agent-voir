package auth

import (
	"net/http"

	agentauth "github.com/agentvoir/agentvoir/packages/auth-go"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/httputil"
)

// NewMiddleware builds registry API authentication middleware from environment.
// Authentication is enabled when OIDC_ISSUER_URL or REGISTRY_API_KEY is set.
func NewMiddleware() func(http.Handler) http.Handler {
	cfg := agentauth.LoadConfigFromEnv("REGISTRY_API_KEY")
	authn := agentauth.NewAuthenticator(cfg)
	return agentauth.Middleware(authn, agentauth.MiddlewareOptions{
		Skip: agentauth.SkipHealthz,
		OnError: func(w http.ResponseWriter, _ *http.Request, err error) {
			message := "authentication required"
			switch err {
			case agentauth.ErrUnauthorized:
				message = "invalid or expired bearer token"
			case agentauth.ErrMissingBearer:
				message = "authorization bearer token required"
			}
			httputil.WriteError(w, http.StatusUnauthorized, message)
		},
	})
}
