package dependencies

type DependencyType string

const (
	DependencyTool        DependencyType = "tool"
	DependencyAPI         DependencyType = "api"
	DependencyModel       DependencyType = "model"
	DependencyVectorStore DependencyType = "vector_store"
	DependencyAgent       DependencyType = "agent"
	DependencyMCPServer   DependencyType = "mcp_server"
)

type Dependency struct {
	AgentID string
	Type    DependencyType
	Name    string
	Version string
	Required bool
}
