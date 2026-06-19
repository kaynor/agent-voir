package cache

import (
	"net/http"
	"slices"
	"strings"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/openai"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/registry"
)

var sensitiveDataClasses = []string{"customer_pii", "phi", "pci", "secrets"}

// ShouldBypass reports whether the request should skip exact cache lookup/write.
func ShouldBypass(r *http.Request, agent registry.AgentConfig, req openai.ChatCompletionRequest) bool {
	if req.Stream {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Cache-Bypass"), "true") {
		return true
	}
	if agent.CacheMode == "off" {
		return true
	}
	if req.Temperature != nil && *req.Temperature > 0 {
		return true
	}
	for _, class := range agent.DataClasses {
		if slices.Contains(sensitiveDataClasses, class) && !agent.SemanticCacheAllowed {
			return true
		}
	}
	return false
}

func CacheReadEnabled(agent registry.AgentConfig, globalMode string) bool {
	mode := agent.CacheMode
	if mode == "" {
		mode = globalMode
	}
	switch mode {
	case "off", "write_only":
		return false
	default:
		return true
	}
}

func CacheWriteEnabled(agent registry.AgentConfig, globalMode string) bool {
	mode := agent.CacheMode
	if mode == "" {
		mode = globalMode
	}
	return mode != "off"
}
