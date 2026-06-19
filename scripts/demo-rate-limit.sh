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
AGENT_ID="rate-limit-demo-agent"
MANIFEST="${ROOT}/examples/agents/rate-limit-demo-agent.yaml"
CHAT_BODY="${ROOT}/examples/demo/quickstart-chat-request.json"

echo "==> Rate limit demo"
echo "    Registering ${AGENT_ID} (5 requests/minute)"
curl -fsS -X POST "http://localhost:${REGISTRY_PORT}/v1/agents/from-manifest" \
  -H "Content-Type: application/yaml" \
  --data-binary "@${MANIFEST}" >/dev/null || true

echo "    Sending requests until rate limited..."
LIMIT_HIT=0
for i in $(seq 1 8); do
  CODE=$(curl -sS -o /tmp/rate-limit-response.json -w "%{http_code}" \
    -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
    -H "Authorization: Bearer ${API_KEY}" \
    -H "Content-Type: application/json" \
    -H "x-agent-id: ${AGENT_ID}" \
    -H "x-agent-version: 0.1.0" \
    -H "x-agent-environment: staging" \
    -H "X-Cache-Bypass: true" \
    --data-binary "@${CHAT_BODY}")
  echo "    request ${i}: HTTP ${CODE}"
  if [[ "${CODE}" == "429" ]]; then
    LIMIT_HIT=1
    RETRY=$(curl -sSI -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
      -H "Authorization: Bearer ${API_KEY}" \
      -H "Content-Type: application/json" \
      -H "x-agent-id: ${AGENT_ID}" \
      -H "x-agent-version: 0.1.0" \
      -H "x-agent-environment: staging" \
      -H "X-Cache-Bypass: true" \
      --data-binary "@${CHAT_BODY}" 2>/dev/null | awk -F': ' 'tolower($1)=="retry-after"{print $2}' | tr -d '\r') || true
    echo "    Retry-After: ${RETRY:-n/a}"
    cat /tmp/rate-limit-response.json
    echo
    break
  fi
done

if [[ "${LIMIT_HIT}" -eq 1 ]]; then
  echo "PASS: gateway rate limited requests (429)"
  exit 0
fi

echo "FAIL: expected HTTP 429 rate limit, did not hit limit in 8 requests" >&2
exit 1
