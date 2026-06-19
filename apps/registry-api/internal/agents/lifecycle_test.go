package agents_test

import (
	"errors"
	"testing"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
)

func TestValidateLifecycleTransition(t *testing.T) {
	t.Run("draft to production blocked", func(t *testing.T) {
		err := agents.ValidateLifecycleTransition("draft", "production")
		if !errors.Is(err, agents.ErrInvalidTransition) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("staging to production allowed", func(t *testing.T) {
		if err := agents.ValidateLifecycleTransition("staging", "production"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})

	t.Run("review to production allowed", func(t *testing.T) {
		if err := agents.ValidateLifecycleTransition("review", "production"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}
