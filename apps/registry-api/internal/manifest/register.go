package manifest

import (
	"errors"
	"fmt"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/budgets"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/dependencies"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/modelroutes"
)

// RegistrationResult is the outcome of registering a manifest in the registry.
type RegistrationResult struct {
	Agent        agents.Agent               `json:"agent"`
	Dependencies []dependencies.Dependency  `json:"dependencies"`
	Budget       *budgets.Budget            `json:"budget,omitempty"`
	ModelRoute   *modelroutes.ModelRoute    `json:"model_route,omitempty"`
}

// Stores are the registry backends needed to apply a manifest.
type Stores struct {
	Agents       agents.Store
	Dependencies dependencies.Store
	Budgets      budgets.Store
	ModelRoutes  modelroutes.Store
}

// Register creates registry records from a parsed manifest document.
func Register(stores Stores, doc *Document) (RegistrationResult, error) {
	agentReq := agents.RegisterRequest{
		AgentID:              doc.Metadata.Name,
		Name:                 doc.Metadata.Name,
		Version:              doc.Metadata.Version,
		OwnerTeam:            doc.Spec.OwnerTeam,
		CostCenter:           doc.Spec.CostCenter,
		Environment:          doc.Spec.Environment,
		Framework:            doc.Spec.Framework,
		RiskLevel:            doc.Spec.RiskLevel,
		Lifecycle:            doc.Spec.Lifecycle,
		CacheMode:            doc.Spec.Cache.Mode,
		CacheTTLSeconds:      doc.Spec.Cache.TTLSeconds,
		SemanticCacheAllowed: doc.Spec.Cache.SemanticCacheAllowed,
		DataClasses:          doc.Spec.DataClasses,
	}

	agentReq.ApplyDefaults()
	if err := agents.ValidateLifecycle(agentReq.Lifecycle); err != nil {
		return RegistrationResult{}, err
	}

	agent, err := stores.Agents.Create(agentReq)
	if err != nil {
		return RegistrationResult{}, err
	}

	result := RegistrationResult{Agent: agent}
	agentID := doc.Metadata.Name
	version := doc.Metadata.Version

	for _, name := range doc.Spec.Dependencies.Tools {
		dep, err := stores.Dependencies.Create(agentID, version, dependencies.CreateRequest{
			DependencyType: dependencies.TypeTool,
			DependencyName: name,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register tool dependency %q: %w", name, err)
		}
		result.Dependencies = append(result.Dependencies, dep)
	}
	for _, name := range doc.Spec.Dependencies.APIs {
		dep, err := stores.Dependencies.Create(agentID, version, dependencies.CreateRequest{
			DependencyType: dependencies.TypeAPI,
			DependencyName: name,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register api dependency %q: %w", name, err)
		}
		result.Dependencies = append(result.Dependencies, dep)
	}
	for _, name := range doc.Spec.Dependencies.VectorStores {
		dep, err := stores.Dependencies.Create(agentID, version, dependencies.CreateRequest{
			DependencyType: dependencies.TypeVectorStore,
			DependencyName: name,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register vector store dependency %q: %w", name, err)
		}
		result.Dependencies = append(result.Dependencies, dep)
	}
	for _, name := range doc.Spec.Dependencies.Agents {
		dep, err := stores.Dependencies.Create(agentID, version, dependencies.CreateRequest{
			DependencyType: dependencies.TypeAgent,
			DependencyName: name,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register agent dependency %q: %w", name, err)
		}
		result.Dependencies = append(result.Dependencies, dep)
	}
	for _, name := range doc.Spec.Dependencies.MCPServers {
		dep, err := stores.Dependencies.Create(agentID, version, dependencies.CreateRequest{
			DependencyType: dependencies.TypeMCPServer,
			DependencyName: name,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register mcp server dependency %q: %w", name, err)
		}
		result.Dependencies = append(result.Dependencies, dep)
	}

	if doc.Spec.Budget.MonthlyUSD > 0 ||
		doc.Spec.Budget.MaxPromptTokensPerRequest > 0 ||
		doc.Spec.Budget.MaxCompletionTokensPerRequest > 0 {
		budget := stores.Budgets.Upsert(agentID, version, budgets.UpsertRequest{
			MonthlyUSD:                    doc.Spec.Budget.MonthlyUSD,
			MaxPromptTokensPerRequest:     doc.Spec.Budget.MaxPromptTokensPerRequest,
			MaxCompletionTokensPerRequest: doc.Spec.Budget.MaxCompletionTokensPerRequest,
		})
		result.Budget = &budget
	}

	if doc.Spec.Models.Primary.Provider != "" && doc.Spec.Models.Primary.Model != "" {
		route, err := stores.ModelRoutes.Upsert(agentID, version, modelroutes.UpsertRequest{
			PrimaryProvider:  doc.Spec.Models.Primary.Provider,
			PrimaryModel:     doc.Spec.Models.Primary.Model,
			FallbackProvider: doc.Spec.Models.Fallback.Provider,
			FallbackModel:    doc.Spec.Models.Fallback.Model,
		})
		if err != nil {
			return RegistrationResult{}, fmt.Errorf("register model route: %w", err)
		}
		result.ModelRoute = &route
	}

	return result, nil
}

// RegisterYAML parses YAML and registers the manifest.
func RegisterYAML(stores Stores, data []byte) (RegistrationResult, error) {
	doc, err := Parse(data)
	if err != nil {
		return RegistrationResult{}, err
	}
	result, err := Register(stores, doc)
	if errors.Is(err, agents.ErrConflict) {
		return RegistrationResult{}, agents.ErrConflict
	}
	return result, err
}
