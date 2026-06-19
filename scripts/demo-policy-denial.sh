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
AGENT_ID="customer-support-agent"
CHAT_BODY="${ROOT}/examples/demo/sample-chat-request.json"

echo "==> Policy denial demo"
echo "    Agent lifecycle: draft (registered as staging environment)"
echo "    Sending request with x-agent-environment: production (should be denied by OPA)"
echo

HTTP_CODE=$(curl -sS -o /tmp/policy-denial-response.json -w "%{http_code}" \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: ${AGENT_ID}" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: production" \
  -H "X-Cache-Bypass: true" \
  --data-binary "@${CHAT_BODY}")

echo "HTTP status: ${HTTP_CODE}"
cat /tmp/policy-denial-response.json
echo

if [[ "${HTTP_CODE}" == "403" ]]; then
  echo "PASS: gateway denied request via policy (403)"
  exit 0
fi

echo "FAIL: expected HTTP 403 policy denial, got ${HTTP_CODE}" >&2
exit 1
