package auth

// Identity is the authenticated caller resolved from a JWT or static API key.
type Identity struct {
	Subject   string   `json:"sub"`
	Email     string   `json:"email,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	TenantID  string   `json:"tenant_id,omitempty"`
	AuthMethod string  `json:"auth_method"` // "oidc" or "api_key"
}
