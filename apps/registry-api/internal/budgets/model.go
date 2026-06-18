package budgets

type Budget struct {
	AgentID                    string
	MonthlyUSD                 float64
	MaxPromptTokensPerRequest  int64
	MaxOutputTokensPerRequest  int64
	RequestsPerMinute          int64
}
