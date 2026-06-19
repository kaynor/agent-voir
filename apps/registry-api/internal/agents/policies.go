package agents

import "encoding/json"

// AgentPolicies holds governance settings loaded from manifest or API updates.
type AgentPolicies struct {
	AllowedProviders []string `json:"allowed_providers"`
	PIIAllowed       bool     `json:"pii_allowed"`
	RequireAuditLog  bool     `json:"require_audit_log"`
}

// OPAFormat returns policies in the shape expected by OPA Rego policies.
func (p AgentPolicies) OPAFormat() map[string]any {
	return map[string]any{
		"allowedProviders": p.AllowedProviders,
		"piiAllowed":       p.PIIAllowed,
		"requireAuditLog":  p.RequireAuditLog,
	}
}

func decodePolicies(raw json.RawMessage) AgentPolicies {
	if len(raw) == 0 {
		return AgentPolicies{}
	}
	var policies AgentPolicies
	_ = json.Unmarshal(raw, &policies)
	return policies
}
