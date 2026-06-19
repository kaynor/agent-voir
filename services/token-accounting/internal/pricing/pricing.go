package pricing

// USD per 1M tokens (input, output).
var modelRates = map[string][2]float64{
	"gpt-4.1-mini": {0.40, 1.60},
	"gpt-4.1":      {2.00, 8.00},
	"gpt-4o-mini":  {0.15, 0.60},
	"gpt-4o":       {2.50, 10.00},
}

// ComputeCostUSD estimates request cost from token counts.
func ComputeCostUSD(model string, promptTokens, completionTokens uint64) float64 {
	rates, ok := modelRates[model]
	if !ok {
		return 0
	}
	inputCost := float64(promptTokens) / 1_000_000 * rates[0]
	outputCost := float64(completionTokens) / 1_000_000 * rates[1]
	return inputCost + outputCost
}
