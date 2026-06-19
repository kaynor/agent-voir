package auth

import (
	"net/http"
)

// MiddlewareOptions configures HTTP authentication middleware.
type MiddlewareOptions struct {
	// Skip returns true for routes that should not require authentication.
	Skip func(*http.Request) bool
	// OnError writes an HTTP error response. Defaults to plain 401 JSON.
	OnError func(w http.ResponseWriter, r *http.Request, err error)
}

// Middleware protects routes using the configured authenticator.
// When auth is disabled (Config.Enabled() == false), requests pass through unchanged.
func Middleware(authn *Authenticator, opts MiddlewareOptions) func(http.Handler) http.Handler {
	if authn == nil || !authn.cfg.Enabled() {
		return func(next http.Handler) http.Handler { return next }
	}

	onError := opts.OnError
	if onError == nil {
		onError = defaultUnauthorized
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.Skip != nil && opts.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}

			id, err := authn.Authenticate(r.Context(), r.Header.Get("Authorization"))
			if err != nil {
				onError(w, r, err)
				return
			}

			r = r.WithContext(ContextWithIdentity(r.Context(), id))
			applyIdentityHeaders(r, id)
			next.ServeHTTP(w, r)
		})
	}
}

func applyIdentityHeaders(r *http.Request, id Identity) {
	if id.Subject != "" && id.AuthMethod == "oidc" {
		r.Header.Set("x-user-id", id.Subject)
	}
	if id.TenantID != "" {
		r.Header.Set("x-tenant-id", id.TenantID)
	}
}

func defaultUnauthorized(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	message := "authentication required"
	if err == ErrUnauthorized {
		message = "invalid or expired bearer token"
	} else if err == ErrMissingBearer {
		message = "authorization bearer token required"
	}
	_, _ = w.Write([]byte(`{"error":"unauthorized","message":"` + message + `"}`))
}

// SkipHealthz skips authentication for GET /healthz.
func SkipHealthz(r *http.Request) bool {
	return r.Method == http.MethodGet && r.URL.Path == "/healthz"
}
