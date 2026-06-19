package budget

import (
	"context"
	"fmt"
	"strings"
)

// Limits are per-agent spend and token caps from the registry.
type Limits struct {
	MonthlyUSD                  float64
	MaxPromptTokensPerRequest   int64
	MaxCompletionTokensPerRequest int64
}

// SpendSummary is monthly spend for an agent.
type SpendSummary struct {
	CostUSD float64
}

// RegistryBudgetLoader loads budget limits from the registry API.
type RegistryBudgetLoader interface {
	GetBudget(ctx context.Context, agentID, version string) (Limits, error)
}

// SpendLoader loads monthly spend from token accounting.
type SpendLoader interface {
	GetMonthlySpend(ctx context.Context, tenantID, agentID string) (SpendSummary, error)
}

// Checker enforces per-agent budgets before upstream calls.
type Checker struct {
	registry RegistryBudgetLoader
	spend    SpendLoader
}

func NewChecker(registry RegistryBudgetLoader, spend SpendLoader) *Checker {
	return &Checker{registry: registry, spend: spend}
}

type Violation struct {
	Code    string
	Message string
}

func (c *Checker) Check(
	ctx context.Context,
	tenantID, agentID, version string,
	estimatedPromptTokens int64,
) *Violation {
	if c == nil || c.registry == nil {
		return nil
	}

	limits, err := c.registry.GetBudget(ctx, agentID, version)
	if err != nil || limits == (Limits{}) {
		return nil
	}

	if limits.MaxPromptTokensPerRequest > 0 && estimatedPromptTokens > limits.MaxPromptTokensPerRequest {
		return &Violation{
			Code:    "budget_exceeded",
			Message: fmt.Sprintf("prompt exceeds max_prompt_tokens_per_request (%d)", limits.MaxPromptTokensPerRequest),
		}
	}

	if limits.MonthlyUSD <= 0 || c.spend == nil {
		return nil
	}

	summary, err := c.spend.GetMonthlySpend(ctx, tenantID, agentID)
	if err != nil {
		return nil
	}
	if summary.CostUSD >= limits.MonthlyUSD {
		return &Violation{
			Code:    "budget_exceeded",
			Message: fmt.Sprintf("monthly budget exceeded ($%.4f / $%.4f)", summary.CostUSD, limits.MonthlyUSD),
		}
	}
	return nil
}

func EstimatePromptTokens(messages []string) int64 {
	var chars int64
	for _, msg := range messages {
		chars += int64(len(msg))
	}
	tokens := chars / 4
	if tokens < 1 {
		return 1
	}
	return tokens
}

func NormalizeTenant(tenantID string) string {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "default"
	}
	return tenantID
}
