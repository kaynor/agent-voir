package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/openai"
)

// ExactKey builds a deterministic cache key for an agent-scoped chat request.
func ExactKey(agentID, agentVersion string, req openai.ChatCompletionRequest) string {
	payload := struct {
		AgentID      string                    `json:"agent_id"`
		AgentVersion string                    `json:"agent_version"`
		Request      openai.ChatCompletionRequest `json:"request"`
	}{
		AgentID:      agentID,
		AgentVersion: agentVersion,
		Request:      req,
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
