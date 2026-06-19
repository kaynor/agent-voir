package agents

import (
	"errors"
	"fmt"
	"slices"
)

var ErrInvalidLifecycle = errors.New("invalid lifecycle")
var ErrInvalidTransition = errors.New("invalid lifecycle transition")

var validLifecycles = []string{"draft", "review", "staging", "production", "deprecated", "retired"}

var allowedTransitions = map[string][]string{
	"draft":      {"review", "staging", "deprecated"},
	"review":     {"draft", "staging", "production", "deprecated"},
	"staging":    {"draft", "review", "production", "deprecated"},
	"production": {"deprecated", "retired"},
	"deprecated": {"retired"},
	"retired":    {},
}

// ValidateLifecycle returns an error when lifecycle is not a supported stage.
func ValidateLifecycle(lifecycle string) error {
	if lifecycle == "" {
		return fmt.Errorf("%w: lifecycle is required", ErrInvalidLifecycle)
	}
	if !slices.Contains(validLifecycles, lifecycle) {
		return fmt.Errorf("%w: %q is not supported (allowed: draft, review, staging, production, deprecated, retired)", ErrInvalidLifecycle, lifecycle)
	}
	return nil
}

// ValidateLifecycleTransition ensures from -> to is allowed.
// Moving to production requires the agent to be in review or staging first.
func ValidateLifecycleTransition(from, to string) error {
	if from == to {
		return nil
	}
	if err := ValidateLifecycle(from); err != nil {
		return err
	}
	if err := ValidateLifecycle(to); err != nil {
		return err
	}
	if to == "production" && from != "review" && from != "staging" {
		return fmt.Errorf("%w: production requires review or staging (current: %s)", ErrInvalidTransition, from)
	}
	allowed := allowedTransitions[from]
	if !slices.Contains(allowed, to) {
		return fmt.Errorf("%w: cannot move from %s to %s", ErrInvalidTransition, from, to)
	}
	return nil
}
