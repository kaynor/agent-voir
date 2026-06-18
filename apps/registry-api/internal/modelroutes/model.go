package modelroutes

import "time"

// ModelRoute defines primary and fallback model routing for an agent version.
type ModelRoute struct {
	ID               string    `json:"id"`
	AgentID          string    `json:"agent_id"`
	AgentVersion     string    `json:"agent_version"`
	PrimaryProvider  string    `json:"primary_provider"`
	PrimaryModel     string    `json:"primary_model"`
	FallbackProvider string    `json:"fallback_provider,omitempty"`
	FallbackModel    string    `json:"fallback_model,omitempty"`
	RoutingPolicy    string    `json:"routing_policy"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UpsertRequest is the JSON body for PUT /v1/agents/{agentID}/model-route.
type UpsertRequest struct {
	PrimaryProvider  string `json:"primary_provider"`
	PrimaryModel     string `json:"primary_model"`
	FallbackProvider string `json:"fallback_provider"`
	FallbackModel    string `json:"fallback_model"`
	RoutingPolicy    string `json:"routing_policy"`
}

func (r *UpsertRequest) ApplyDefaults() {
	if r.RoutingPolicy == "" {
		r.RoutingPolicy = "primary_then_fallback"
	}
}

func (r *UpsertRequest) Validate() string {
	switch {
	case r.PrimaryProvider == "":
		return "primary_provider is required"
	case r.PrimaryModel == "":
		return "primary_model is required"
	default:
		return ""
	}
}
