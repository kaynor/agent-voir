package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var (
	cacheHits   atomic.Uint64
	cacheMisses atomic.Uint64
	cacheBypass atomic.Uint64
)

func RecordCacheHit()   { cacheHits.Add(1) }
func RecordCacheMiss()  { cacheMisses.Add(1) }
func RecordCacheBypass(){ cacheBypass.Add(1) }

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP agentvoir_cache_hits_total Exact cache hits\n")
		fmt.Fprintf(w, "agentvoir_cache_hits_total %d\n", cacheHits.Load())
		fmt.Fprintf(w, "# HELP agentvoir_cache_misses_total Exact cache misses\n")
		fmt.Fprintf(w, "agentvoir_cache_misses_total %d\n", cacheMisses.Load())
		fmt.Fprintf(w, "# HELP agentvoir_cache_bypass_total Requests that bypassed cache\n")
		fmt.Fprintf(w, "agentvoir_cache_bypass_total %d\n", cacheBypass.Load())
	})
}
