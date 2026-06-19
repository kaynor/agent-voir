package accounting

import (
	"context"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/budget"
)

// SpendAdapter implements budget.SpendLoader.
type SpendAdapter struct {
	client *Client
}

func NewSpendAdapter(client *Client) *SpendAdapter {
	return &SpendAdapter{client: client}
}

func (a *SpendAdapter) GetMonthlySpend(ctx context.Context, tenantID, agentID string) (budget.SpendSummary, error) {
	if a == nil || a.client == nil {
		return budget.SpendSummary{}, nil
	}
	summary, err := a.client.GetMonthlySummary(ctx, tenantID, agentID)
	if err != nil {
		return budget.SpendSummary{}, err
	}
	return budget.SpendSummary{CostUSD: summary.CostUSD}, nil
}
