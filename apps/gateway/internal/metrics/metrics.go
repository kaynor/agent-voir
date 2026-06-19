package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var (
	cacheHits        atomic.Uint64
	cacheMisses      atomic.Uint64
	cacheBypass      atomic.Uint64
	policyDenied     atomic.Uint64
	budgetExceeded   atomic.Uint64
)

func RecordCacheHit()        { cacheHits.Add(1) }
func RecordCacheMiss()       { cacheMisses.Add(1) }
func RecordCacheBypass()     { cacheBypass.Add(1) }
func RecordPolicyDenied()    { policyDenied.Add(1) }
func RecordBudgetExceeded() { budgetExceeded.Add(1) }

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP agentvoir_cache_hits_total Exact cache hits\n")
		fmt.Fprintf(w, "agentvoir_cache_hits_total %d\n", cacheHits.Load())
		fmt.Fprintf(w, "# HELP agentvoir_cache_misses_total Exact cache misses\n")
		fmt.Fprintf(w, "agentvoir_cache_misses_total %d\n", cacheMisses.Load())
		fmt.Fprintf(w, "# HELP agentvoir_cache_bypass_total Requests that bypassed cache\n")
		fmt.Fprintf(w, "agentvoir_cache_bypass_total %d\n", cacheBypass.Load())
		fmt.Fprintf(w, "# HELP agentvoir_policy_denied_total Requests denied by OPA policy\n")
		fmt.Fprintf(w, "agentvoir_policy_denied_total %d\n", policyDenied.Load())
		fmt.Fprintf(w, "# HELP agentvoir_budget_exceeded_total Requests blocked by budget limits\n")
		fmt.Fprintf(w, "agentvoir_budget_exceeded_total %d\n", budgetExceeded.Load())
	})
}
