package auth

import (
	"os"
	"strings"
)

// Config holds OIDC and static API key authentication settings.
type Config struct {
	// IssuerURL enables OIDC JWT validation when set (OIDC_ISSUER_URL).
	IssuerURL string
	// Audience is the expected JWT aud claim (OIDC_AUDIENCE). Optional.
	Audience string
	// ClientID is used as audience when Audience is empty (OIDC_CLIENT_ID).
	ClientID string
	// GroupsClaim is the JWT claim for group membership (default: groups).
	GroupsClaim string
	// TenantClaim is an optional JWT claim mapped to tenant ID.
	TenantClaim string
	// StaticAPIKeys are bootstrap keys accepted alongside JWTs.
	StaticAPIKeys []string
}

// Enabled reports whether any authentication mechanism is configured.
func (c Config) Enabled() bool {
	return c.IssuerURL != "" || len(c.StaticAPIKeys) > 0
}

// LoadConfigFromEnv reads standard AgentVoir OIDC environment variables.
// staticKeyEnv is the service-specific bootstrap key env var name
// (e.g. GATEWAY_API_KEY or REGISTRY_API_KEY).
func LoadConfigFromEnv(staticKeyEnv string) Config {
	cfg := Config{
		IssuerURL:   strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL")),
		Audience:    strings.TrimSpace(os.Getenv("OIDC_AUDIENCE")),
		ClientID:    strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID")),
		GroupsClaim: envDefault("OIDC_GROUPS_CLAIM", "groups"),
		TenantClaim: strings.TrimSpace(os.Getenv("OIDC_TENANT_CLAIM")),
	}
	if key := strings.TrimSpace(os.Getenv(staticKeyEnv)); key != "" {
		cfg.StaticAPIKeys = []string{key}
	}
	return cfg
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
