#!/usr/bin/env bash
# Create standard GitHub labels for AgentVoir (requires gh CLI + repo admin).
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "Install GitHub CLI: https://cli.github.com/" >&2
  exit 1
fi

create_label() {
  local name="$1"
  local color="$2"
  local description="$3"
  gh label create "$name" --color "$color" --description "$description" --force
}

echo "Creating labels..."
create_label "good-first-issue" "7057ff" "Approachable for new contributors"
create_label "help-wanted" "008672" "Maintainer welcomes community help"
create_label "ai-suggested" "fbca04" "Suggested by scout/AI workflow; needs human approval"
create_label "gateway" "1d76db" "LLM gateway / proxy"
create_label "registry-api" "5319e7" "Agent registry API"
create_label "sdk" "bfd4f2" "Python, TypeScript, or Go SDK"
create_label "docs" "0075ca" "Documentation"
create_label "security" "d93f0b" "Security or supply chain"
create_label "observability" "006b75" "Metrics, traces, dashboards"
create_label "docker" "f9d0c4" "Docker, Compose, GHCR"
create_label "phase-0" "ededed" "Developer experience / trust"
create_label "phase-1" "c2e0c6" "Registry and exact cache"
create_label "phase-2" "fef2c0" "Enterprise controls"

echo "Done. List labels: gh label list"
