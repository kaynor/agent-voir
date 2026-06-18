package dependencies

import "time"

type Type string

const (
	TypeTool        Type = "tool"
	TypeAPI         Type = "api"
	TypeModel       Type = "model"
	TypeVectorStore Type = "vector_store"
	TypeAgent       Type = "agent"
	TypeMCPServer   Type = "mcp_server"
)

// Dependency links an agent version to an external resource or another agent.
type Dependency struct {
	ID              string    `json:"id"`
	AgentID         string    `json:"agent_id"`
	AgentVersion    string    `json:"agent_version"`
	DependencyType  Type      `json:"dependency_type"`
	DependencyName  string    `json:"dependency_name"`
	DependencyVersion string  `json:"dependency_version,omitempty"`
	Required        bool      `json:"required"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateRequest is the JSON body for POST /v1/agents/{agentID}/dependencies.
type CreateRequest struct {
	DependencyType    Type   `json:"dependency_type"`
	DependencyName    string `json:"dependency_name"`
	DependencyVersion string `json:"dependency_version"`
	Required          *bool  `json:"required"`
}

func (r *CreateRequest) Validate() string {
	switch {
	case r.DependencyType == "":
		return "dependency_type is required"
	case r.DependencyName == "":
		return "dependency_name is required"
	default:
		return ""
	}
}

func (r *CreateRequest) RequiredValue() bool {
	if r.Required == nil {
		return true
	}
	return *r.Required
}

// GraphNode is a node in the dependency graph response.
type GraphNode struct {
	ID   string `json:"id"`
	Type Type   `json:"type"`
	Name string `json:"name"`
}

// GraphEdge connects an agent to a dependency node.
type GraphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Required bool   `json:"required"`
}

// Graph is the dependency graph for an agent version.
type Graph struct {
	AgentID      string      `json:"agent_id"`
	AgentVersion string      `json:"agent_version"`
	Nodes        []GraphNode `json:"nodes"`
	Edges        []GraphEdge `json:"edges"`
}
