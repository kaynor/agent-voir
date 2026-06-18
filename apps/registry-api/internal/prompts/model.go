package prompts

type Prompt struct {
	ID             string
	Name           string
	Version        string
	OwnerTeam      string
	Template       string
	RiskLevel      string
	ApprovedModels []string
}
