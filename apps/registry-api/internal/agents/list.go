package agents

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// ListOptions controls pagination, sorting, and filtering for GET /v1/agents.
type ListOptions struct {
	Limit       int
	Offset      int
	SortBy      string
	SortOrder   string
	Environment string
}

// ListResult is the paginated response envelope for agent listing.
type ListResult struct {
	Items  []Agent `json:"items"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// ParseListOptions reads list query parameters from an HTTP request.
func ParseListOptions(r *http.Request) (ListOptions, string) {
	opts := ListOptions{
		Limit:     defaultListLimit,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := parsePositiveInt(raw)
		if err != nil || parsed > maxListLimit {
			return ListOptions{}, "limit must be a positive integer up to 200"
		}
		opts.Limit = parsed
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return ListOptions{}, "offset must be a non-negative integer"
		}
		opts.Offset = parsed
	}

	opts.SortBy = strings.TrimSpace(r.URL.Query().Get("sort"))
	if opts.SortBy == "" {
		opts.SortBy = "created_at"
	}
	if !slices.Contains([]string{"created_at", "updated_at", "agent_id", "name"}, opts.SortBy) {
		return ListOptions{}, "sort must be one of: created_at, updated_at, agent_id, name"
	}

	opts.SortOrder = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("order")))
	if opts.SortOrder == "" {
		opts.SortOrder = "desc"
	}
	if opts.SortOrder != "asc" && opts.SortOrder != "desc" {
		return ListOptions{}, "order must be asc or desc"
	}

	opts.Environment = strings.TrimSpace(r.URL.Query().Get("environment"))
	return opts, ""
}

func parsePositiveInt(raw string) (int, error) {
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, err
	}
	return parsed, nil
}
