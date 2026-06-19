package agents

import "time"

// Agent is the persisted registry record for an enterprise agent.
type Agent struct {
	ID                   string        `json:"id"`
	AgentID              string        `json:"agent_id"`
	Name                 string        `json:"name"`
	Version              string        `json:"version"`
	OwnerTeam            string        `json:"owner_team"`
	CostCenter           string        `json:"cost_center,omitempty"`
	Environment          string        `json:"environment"`
	Framework            string        `json:"framework,omitempty"`
	RiskLevel            string        `json:"risk_level"`
	Lifecycle            string        `json:"lifecycle"`
	CacheMode            string        `json:"cache_mode,omitempty"`
	CacheTTLSeconds      int64         `json:"cache_ttl_seconds,omitempty"`
	SemanticCacheAllowed bool          `json:"semantic_cache_allowed"`
	Policies             AgentPolicies `json:"policies"`
	DataClasses          []string      `json:"data_classes"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

// RegisterRequest is the JSON body for POST /v1/agents.
type RegisterRequest struct {
	AgentID     string   `json:"agent_id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	OwnerTeam   string   `json:"owner_team"`
	CostCenter  string   `json:"cost_center"`
	Environment string   `json:"environment"`
	Framework   string   `json:"framework"`
	RiskLevel   string   `json:"risk_level"`
	Lifecycle            string   `json:"lifecycle"`
	CacheMode            string   `json:"cache_mode"`
	CacheTTLSeconds      int64    `json:"cache_ttl_seconds"`
	SemanticCacheAllowed bool          `json:"semantic_cache_allowed"`
	Policies             AgentPolicies `json:"policies"`
	DataClasses          []string      `json:"data_classes"`
}

// ApplyDefaults fills in registry defaults for optional fields.
func (r *RegisterRequest) ApplyDefaults() {
	if r.Environment == "" {
		r.Environment = "dev"
	}
	if r.Lifecycle == "" {
		r.Lifecycle = "draft"
	}
	if r.RiskLevel == "" {
		r.RiskLevel = "low"
	}
	if r.CacheMode == "" {
		r.CacheMode = "exact_only"
	}
	if r.CacheTTLSeconds == 0 {
		r.CacheTTLSeconds = 86400
	}
	if r.DataClasses == nil {
		r.DataClasses = []string{}
	}
	if r.Policies.AllowedProviders == nil {
		r.Policies.AllowedProviders = []string{}
	}
}

// Validate returns a human-readable error when required fields are missing.
func (r *RegisterRequest) Validate() string {
	switch {
	case r.AgentID == "":
		return "agent_id is required"
	case r.Name == "":
		return "name is required"
	case r.Version == "":
		return "version is required"
	case r.OwnerTeam == "":
		return "owner_team is required"
	default:
		return ""
	}
}

// UpdateRequest is the JSON body for PUT /v1/agents/{agentID}.
type UpdateRequest struct {
	Name        string   `json:"name"`
	OwnerTeam   string   `json:"owner_team"`
	CostCenter  string   `json:"cost_center"`
	Framework   string   `json:"framework"`
	RiskLevel   string   `json:"risk_level"`
	Lifecycle            string   `json:"lifecycle"`
	CacheMode            string   `json:"cache_mode"`
	CacheTTLSeconds      int64    `json:"cache_ttl_seconds"`
	SemanticCacheAllowed *bool         `json:"semantic_cache_allowed"`
	Policies             *AgentPolicies `json:"policies"`
	DataClasses          []string      `json:"data_classes"`
}
