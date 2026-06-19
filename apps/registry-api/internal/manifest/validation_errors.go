package manifest

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidManifest = errors.New("invalid agent manifest")

// ValidationIssue describes a single manifest validation problem.
type ValidationIssue struct {
	Field   string `json:"field"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// ValidationErrors groups multiple manifest validation issues.
type ValidationErrors struct {
	Issues []ValidationIssue `json:"issues"`
}

func (e ValidationErrors) Error() string {
	if len(e.Issues) == 0 {
		return "invalid agent manifest"
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		if issue.Line > 0 {
			parts = append(parts, fmt.Sprintf("line %d: %s (%s)", issue.Line, issue.Message, issue.Field))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", issue.Message, issue.Field))
	}
	return fmt.Sprintf("%s: %s", ErrInvalidManifest.Error(), strings.Join(parts, "; "))
}

func (e ValidationErrors) HasIssues() bool {
	return len(e.Issues) > 0
}

func issue(field, message string, line int) ValidationIssue {
	return ValidationIssue{Field: field, Message: message, Line: line}
}
