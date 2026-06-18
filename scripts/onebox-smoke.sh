#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/deployments/docker/.env.onebox"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
USAGE_PORT="${AGENTVOIR_USAGE_PORT:-8082}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"

echo "==> registry /healthz"
curl -fsS "http://localhost:${REGISTRY_PORT}/healthz"
echo

echo "==> usage /healthz"
curl -fsS "http://localhost:${USAGE_PORT}/healthz"
echo

echo "==> gateway /healthz"
curl -fsS "http://localhost:${GATEWAY_PORT}/healthz"
echo

echo "==> gateway /v1/models"
curl -fsS "http://localhost:${GATEWAY_PORT}/v1/models" \
  -H "Authorization: Bearer ${API_KEY}" | head -c 200
echo

echo "Onebox smoke checks passed."
