#!/usr/bin/env bash
# Seed GitHub issues from the Phase 0+ backlog (requires gh CLI).
# Safe to rerun — skips issues that already exist (by title prefix).
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "Install GitHub CLI: https://cli.github.com/" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"${SCRIPT_DIR}/bootstrap-github-labels.sh"

issue_exists() {
  local title="$1"
  gh issue list --search "in:title \"${title}\"" --json title --jq '.[].title' | grep -Fxq "$title"
}

create_issue() {
  local title="$1"
  local labels="$2"
  local body="$3"
  if issue_exists "$title"; then
    echo "SKIP (exists): $title"
    return 0
  fi
  gh issue create --title "$title" --label "$labels" --body "$body"
  echo "CREATED: $title"
}

create_issue \
  "[gateway] Wire OPA policy checks before upstream model calls" \
  "gateway,security,phase-2,help-wanted" \
  "## Goal
Call OPA from the gateway on each chat completion and enforce allow/deny.

## Acceptance criteria
- [ ] Gateway loads agent policy context from registry
- [ ] OPA query uses policies/opa/agentvoir.rego
- [ ] Denied requests return 403 with structured error
- [ ] Usage events record policy denials

## Suggested files
- apps/gateway/internal/gateway/handler.go
- policies/opa/"

create_issue \
  "[registry-api] Add pagination to GET /v1/agents" \
  "registry-api,phase-1,good-first-issue" \
  "## Goal
Support limit/offset or cursor pagination on agent list.

## Acceptance criteria
- [ ] Query params documented in OpenAPI
- [ ] Tests for pagination boundaries"

create_issue \
  "[sdk] Publish Python SDK to PyPI" \
  "sdk,phase-1,help-wanted" \
  "## Goal
Package and publish packages/sdk-python for pip install.

## Acceptance criteria
- [ ] CI publish workflow on release tag
- [ ] README install instructions updated"

create_issue \
  "[docs] Expand API examples for gateway and usage APIs" \
  "docs,phase-0,good-first-issue" \
  "## Goal
Document authentication, registration, gateway, usage, and policy examples in docs/api/examples.md.

## Acceptance criteria
- [ ] curl examples for each API surface
- [ ] SDK snippets linked from examples"

create_issue \
  "[security] Sign container images with cosign on release" \
  "security,docker,phase-1" \
  "## Goal
Add Sigstore/cosign signing to release-images workflow.

## Acceptance criteria
- [ ] Images signed on GitHub Release
- [ ] INSTALL.md documents verification steps"

create_issue \
  "[observability] Expose Prometheus /metrics on gateway" \
  "observability,gateway,phase-2,help-wanted" \
  "## Goal
Implement metrics listed in apps/gateway/internal/observability/metrics.go.

## Acceptance criteria
- [ ] /metrics endpoint
- [ ] Request count, latency, cache hit/miss counters
- [ ] Grafana dashboard panel updated"

echo "Backlog seed complete."
