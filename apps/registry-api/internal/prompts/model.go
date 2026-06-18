package prompts

import "time"

// Prompt is a versioned prompt template in the registry.
type Prompt struct {
	ID             string    `json:"id"`
	PromptID       string    `json:"prompt_id"`
	Name           string    `json:"name"`
	Version        string    `json:"version"`
	OwnerTeam      string    `json:"owner_team"`
	Template       string    `json:"template"`
	RiskLevel      string    `json:"risk_level"`
	ApprovedModels []string  `json:"approved_models"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RegisterRequest is the JSON body for POST /v1/prompts.
type RegisterRequest struct {
	PromptID       string   `json:"prompt_id"`
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	OwnerTeam      string   `json:"owner_team"`
	Template       string   `json:"template"`
	RiskLevel      string   `json:"risk_level"`
	ApprovedModels []string `json:"approved_models"`
}

func (r *RegisterRequest) ApplyDefaults() {
	if r.RiskLevel == "" {
		r.RiskLevel = "low"
	}
	if r.ApprovedModels == nil {
		r.ApprovedModels = []string{}
	}
}

func (r *RegisterRequest) Validate() string {
	switch {
	case r.PromptID == "":
		return "prompt_id is required"
	case r.Name == "":
		return "name is required"
	case r.Version == "":
		return "version is required"
	case r.OwnerTeam == "":
		return "owner_team is required"
	case r.Template == "":
		return "template is required"
	default:
		return ""
	}
}

// UpdateRequest is the JSON body for PUT /v1/prompts/{promptID}.
type UpdateRequest struct {
	Name           string   `json:"name"`
	OwnerTeam      string   `json:"owner_team"`
	Template       string   `json:"template"`
	RiskLevel      string   `json:"risk_level"`
	ApprovedModels []string `json:"approved_models"`
}
