#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT}/deployments/docker/.env.onebox"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
AGENT_ID="budget-demo-agent"

echo "==> Budget status demo"
echo "    GET /v1/agents/${AGENT_ID}/budget/status"
curl -fsS "http://localhost:${REGISTRY_PORT}/v1/agents/${AGENT_ID}/budget/status?version=0.1.0" | tee /tmp/budget-status.json
echo

LIMIT=$(python3 -c 'import json; print(json.load(open("/tmp/budget-status.json")).get("monthly_usd_limit", 0))' 2>/dev/null || echo "0")
if [[ "${LIMIT}" != "0" && "${LIMIT}" != "0.0" ]]; then
  echo "PASS: budget status returned utilization fields"
  exit 0
fi

echo "NOTE: register budget-demo-agent first (./scripts/demo-budget-block.sh) if limit is 0"
echo "PASS: budget status endpoint reachable"
exit 0
