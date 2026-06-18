package router

type ModelRoute struct {
	AgentID          string
	PrimaryProvider  string
	PrimaryModel     string
	FallbackProvider string
	FallbackModel    string
	Policy           string
}
