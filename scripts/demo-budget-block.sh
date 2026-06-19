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
USAGE_PORT="${AGENTVOIR_USAGE_PORT:-8082}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"
MANIFEST="${ROOT}/examples/agents/budget-demo-agent.yaml"
CHAT_BODY="${ROOT}/examples/demo/sample-chat-request.json"
AGENT_ID="budget-demo-agent"

echo "==> Budget block demo"
echo "    Registering agent with monthlyUsd: 0.001"
curl -fsS -X POST "http://localhost:${REGISTRY_PORT}/v1/agents/from-manifest" \
  -H "Content-Type: application/x-yaml" \
  --data-binary "@${MANIFEST}" || true
echo

echo "==> Simulating prior spend via usage ingestion"
curl -fsS -X POST "http://localhost:${USAGE_PORT}/v1/usage-events" \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"${AGENT_ID}\",
    \"agent_version\": \"0.1.0\",
    \"tenant_id\": \"default\",
    \"provider\": \"openai\",
    \"model\": \"gpt-4.1-mini\",
    \"cache_status\": \"miss\",
    \"prompt_tokens\": 100,
    \"completion_tokens\": 50,
    \"cost_usd\": 0.001,
    \"latency_ms\": 100,
    \"status_code\": 200
  }"
echo

echo "==> Gateway request (should be blocked — monthly budget exceeded)"
HTTP_CODE=$(curl -sS -o /tmp/budget-block-response.json -w "%{http_code}" \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: ${AGENT_ID}" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: staging" \
  -H "X-Cache-Bypass: true" \
  --data-binary "@${CHAT_BODY}")

echo "HTTP status: ${HTTP_CODE}"
cat /tmp/budget-block-response.json
echo

if [[ "${HTTP_CODE}" == "429" ]]; then
  echo "PASS: gateway blocked request due to budget (429)"
  exit 0
fi

echo "FAIL: expected HTTP 429 budget block, got ${HTTP_CODE}" >&2
exit 1
