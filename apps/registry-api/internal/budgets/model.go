package budgets

import "time"

// Budget defines spend and rate limits for an agent version.
type Budget struct {
	ID                              string    `json:"id"`
	AgentID                         string    `json:"agent_id"`
	AgentVersion                    string    `json:"agent_version"`
	MonthlyUSD                      float64   `json:"monthly_usd,omitempty"`
	MaxPromptTokensPerRequest       int64     `json:"max_prompt_tokens_per_request,omitempty"`
	MaxCompletionTokensPerRequest   int64     `json:"max_completion_tokens_per_request,omitempty"`
	RequestsPerMinute               int64     `json:"requests_per_minute,omitempty"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}

// UpsertRequest is the JSON body for PUT /v1/agents/{agentID}/budget.
type UpsertRequest struct {
	MonthlyUSD                    float64 `json:"monthly_usd"`
	MaxPromptTokensPerRequest     int64   `json:"max_prompt_tokens_per_request"`
	MaxCompletionTokensPerRequest int64   `json:"max_completion_tokens_per_request"`
	RequestsPerMinute             int64   `json:"requests_per_minute"`
}
