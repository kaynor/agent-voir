package budget

import (
	"context"

	agentregistry "github.com/agentvoir/agentvoir/apps/gateway/internal/registry"
)

// RegistryAdapter adapts registry.Client to RegistryBudgetLoader.
type RegistryAdapter struct {
	Client *agentregistry.Client
}

func (a *RegistryAdapter) GetBudget(ctx context.Context, agentID, version string) (Limits, error) {
	if a == nil || a.Client == nil {
		return Limits{}, nil
	}
	limits, err := a.Client.GetBudget(ctx, agentID, version)
	if err != nil {
		return Limits{}, err
	}
	return Limits{
		MonthlyUSD:                    limits.MonthlyUSD,
		MaxPromptTokensPerRequest:     limits.MaxPromptTokensPerRequest,
		MaxCompletionTokensPerRequest: limits.MaxCompletionTokensPerRequest,
		RequestsPerMinute:             limits.RequestsPerMinute,
	}, nil
}
