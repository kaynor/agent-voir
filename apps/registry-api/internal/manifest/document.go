package manifest

// Document is the top-level AgentVoir agent manifest (YAML).
type Document struct {
	APIVersion string   `yaml:"apiVersion" json:"api_version"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata identifies the agent manifest.
type Metadata struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

// Spec holds agent configuration from the manifest.
type Spec struct {
	OwnerTeam   string       `yaml:"ownerTeam" json:"owner_team"`
	CostCenter  string       `yaml:"costCenter" json:"cost_center"`
	Environment string       `yaml:"environment" json:"environment"`
	Framework   string       `yaml:"framework" json:"framework"`
	RiskLevel   string       `yaml:"riskLevel" json:"risk_level"`
	DataClasses []string     `yaml:"dataClasses" json:"data_classes"`
	Lifecycle   string       `yaml:"lifecycle" json:"lifecycle"`
	Models      Models       `yaml:"models" json:"models"`
	Cache       Cache        `yaml:"cache" json:"cache"`
	Budget      Budget       `yaml:"budget" json:"budget"`
	Policies    Policies     `yaml:"policies" json:"policies"`
	Dependencies Dependencies `yaml:"dependencies" json:"dependencies"`
}

type Policies struct {
	PIIAllowed       bool     `yaml:"piiAllowed" json:"pii_allowed"`
	RequireAuditLog  bool     `yaml:"requireAuditLog" json:"require_audit_log"`
	AllowedProviders []string `yaml:"allowedProviders" json:"allowed_providers"`
}

type Cache struct {
	Mode                 string `yaml:"mode" json:"mode"`
	TTLSeconds           int64  `yaml:"ttlSeconds" json:"ttl_seconds"`
	SemanticCacheAllowed bool   `yaml:"semanticCacheAllowed" json:"semantic_cache_allowed"`
}

type Models struct {
	Primary  ModelRef `yaml:"primary" json:"primary"`
	Fallback ModelRef `yaml:"fallback" json:"fallback"`
}

type ModelRef struct {
	Provider string `yaml:"provider" json:"provider"`
	Model    string `yaml:"model" json:"model"`
}

type Budget struct {
	MonthlyUSD                    float64 `yaml:"monthlyUsd" json:"monthly_usd"`
	MaxPromptTokensPerRequest     int64   `yaml:"maxPromptTokensPerRequest" json:"max_prompt_tokens_per_request"`
	MaxCompletionTokensPerRequest int64   `yaml:"maxCompletionTokensPerRequest" json:"max_completion_tokens_per_request"`
	RequestsPerMinute             int64   `yaml:"requestsPerMinute" json:"requests_per_minute"`
}

type Dependencies struct {
	Tools        []string `yaml:"tools" json:"tools"`
	APIs         []string `yaml:"apis" json:"apis"`
	VectorStores []string `yaml:"vectorStores" json:"vector_stores"`
	Agents       []string `yaml:"agents" json:"agents"`
	MCPServers   []string `yaml:"mcpServers" json:"mcp_servers"`
}
