package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Authenticator validates Bearer tokens (OIDC JWT or static API key).
type Authenticator struct {
	cfg      Config
	verifier *oidc.IDTokenVerifier
	initOnce sync.Once
	initErr  error
}

// NewAuthenticator builds an authenticator from config. OIDC provider discovery
// is lazy on first JWT validation.
func NewAuthenticator(cfg Config) *Authenticator {
	return &Authenticator{cfg: cfg}
}

// Authenticate resolves the caller from Authorization: Bearer <token>.
func (a *Authenticator) Authenticate(ctx context.Context, authorizationHeader string) (Identity, error) {
	token, ok := bearerToken(authorizationHeader)
	if !ok {
		return Identity{}, ErrMissingBearer
	}

	for _, key := range a.cfg.StaticAPIKeys {
		if token == key {
			return Identity{
				Subject:    "bootstrap-api-key",
				AuthMethod: "api_key",
			}, nil
		}
	}

	if a.cfg.IssuerURL == "" {
		return Identity{}, ErrUnauthorized
	}

	return a.authenticateOIDC(ctx, token)
}

func (a *Authenticator) authenticateOIDC(ctx context.Context, token string) (Identity, error) {
	verifier, err := a.oidcVerifier(ctx)
	if err != nil {
		return Identity{}, fmt.Errorf("oidc verifier: %w", err)
	}

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return Identity{}, ErrUnauthorized
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return Identity{}, ErrUnauthorized
	}

	id := Identity{
		Subject:    idToken.Subject,
		AuthMethod: "oidc",
	}
	if email, _ := claimString(claims, "email"); email != "" {
		id.Email = email
	}
	if groups := claimStringSlice(claims, a.cfg.GroupsClaim); len(groups) > 0 {
		id.Groups = groups
	}
	if a.cfg.TenantClaim != "" {
		if tenant, _ := claimString(claims, a.cfg.TenantClaim); tenant != "" {
			id.TenantID = tenant
		}
	}
	return id, nil
}

func (a *Authenticator) oidcVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	a.initOnce.Do(func() {
		provider, err := oidc.NewProvider(ctx, a.cfg.IssuerURL)
		if err != nil {
			a.initErr = err
			return
		}
		verifierConfig := &oidc.Config{}
		switch {
		case a.cfg.Audience != "":
			verifierConfig.ClientID = a.cfg.Audience
		case a.cfg.ClientID != "":
			verifierConfig.ClientID = a.cfg.ClientID
		}
		a.verifier = provider.Verifier(verifierConfig)
	})
	return a.verifier, a.initErr
}

func bearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	return token, token != ""
}

func claimString(claims map[string]any, key string) (string, bool) {
	raw, ok := claims[key]
	if !ok || raw == nil {
		return "", false
	}
	switch v := raw.(type) {
	case string:
		return v, v != ""
	default:
		return fmt.Sprint(v), true
	}
}

func claimStringSlice(claims map[string]any, key string) []string {
	raw, ok := claims[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
