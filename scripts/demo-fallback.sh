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

GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"
AGENT_ID="fallback-demo-agent"
MANIFEST="${ROOT}/examples/agents/fallback-demo-agent.yaml"
CHAT_BODY="${ROOT}/examples/demo/quickstart-chat-request.json"

echo "==> Provider fallback demo"
echo "    Registering ${AGENT_ID} (primary=unavailable, fallback=mock)"
curl -fsS -X POST "http://localhost:${REGISTRY_PORT}/v1/agents/from-manifest" \
  -H "Content-Type: application/yaml" \
  --data-binary "@${MANIFEST}" >/dev/null || true

HTTP_CODE=$(curl -sS -D /tmp/fallback-headers.txt -o /tmp/fallback-response.json -w "%{http_code}" \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: ${AGENT_ID}" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: staging" \
  -H "X-Cache-Bypass: true" \
  --data-binary "@${CHAT_BODY}")

echo "HTTP status: ${HTTP_CODE}"
PROVIDER=$(grep -i '^x-model-provider:' /tmp/fallback-headers.txt | awk '{print $2}' | tr -d '\r')
FALLBACK=$(grep -i '^x-routing-fallback:' /tmp/fallback-headers.txt | awk '{print $2}' | tr -d '\r')
echo "x-model-provider: ${PROVIDER:-n/a}"
echo "x-routing-fallback: ${FALLBACK:-n/a}"
cat /tmp/fallback-response.json
echo

if [[ "${HTTP_CODE}" == "200" && "${PROVIDER}" == "mock" && "${FALLBACK}" == "true" ]]; then
  echo "PASS: gateway used fallback provider"
  exit 0
fi

echo "FAIL: expected HTTP 200 with mock fallback provider" >&2
exit 1
