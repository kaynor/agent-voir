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

echo "==> Policy simulation demo"
echo "    Simulating draft agent in production (should deny)"
curl -fsS -X POST "http://localhost:${REGISTRY_PORT}/v1/policies/simulate" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "production",
    "agent": {
      "lifecycle": "draft",
      "policies": { "allowedProviders": ["mock"], "piiAllowed": false, "requireAuditLog": false },
      "cache": { "mode": "off", "semanticCacheAllowed": false },
      "dataClasses": []
    },
    "request": { "provider": "mock", "contains_pii": false, "contains_secret": false }
  }' | tee /tmp/policy-simulate.json
echo

ALLOWED=$(python3 -c 'import json; print(json.load(open("/tmp/policy-simulate.json"))["allowed"])' 2>/dev/null || echo "false")
if [[ "${ALLOWED}" == "False" || "${ALLOWED}" == "false" ]]; then
  echo "PASS: policy simulation denied draft agent in production"
  exit 0
fi

echo "FAIL: expected allowed=false" >&2
exit 1
