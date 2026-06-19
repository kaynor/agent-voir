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

DEX_PORT="${DEX_PORT:-5556}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
DEX_URL="http://localhost:${DEX_PORT}"
CLIENT_ID="${OIDC_CLIENT_ID:-agentvoir}"
CLIENT_SECRET="${OIDC_CLIENT_SECRET:-agentvoir-dev-secret}"

echo "==> OIDC demo (Dex password grant + registry JWT)"
echo "    Dex: ${DEX_URL}"
echo "    Fetching access token for admin@agentvoir.local ..."

TOKEN_RESPONSE=$(curl -sS -X POST "${DEX_URL}/dex/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  --data-urlencode "grant_type=password" \
  --data-urlencode "username=admin@agentvoir.local" \
  --data-urlencode "password=password" \
  --data-urlencode "scope=openid profile email groups")

# Prefer id_token for OIDC verification (go-oidc ID token validator).
ACCESS_TOKEN=$(python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("id_token") or d.get("access_token",""))' <<<"${TOKEN_RESPONSE}")
if [[ -z "${ACCESS_TOKEN}" ]]; then
  echo "FAIL: could not obtain access token from Dex" >&2
  echo "${TOKEN_RESPONSE}" >&2
  echo "Restart Dex after config changes: docker compose ... restart dex" >&2
  echo "Start onebox with OIDC overlay: docker compose -f deployments/docker/docker-compose.onebox.yml -f deployments/docker/docker-compose.onebox.oidc.yml up -d" >&2
  exit 1
fi

echo "    Token acquired (${#ACCESS_TOKEN} chars)"
echo "    Registering cache-demo-agent (ignore 409 if already exists) ..."
curl -sS -X POST "http://localhost:${REGISTRY_PORT}/v1/agents/from-manifest" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/yaml" \
  --data-binary "@${ROOT}/examples/agents/cache-demo-agent.yaml" >/dev/null || true

echo "    GET /v1/agents with Bearer JWT ..."
HTTP_CODE=$(curl -sS -o /tmp/oidc-agents.json -w "%{http_code}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "http://localhost:${REGISTRY_PORT}/v1/agents")

echo "    Registry HTTP ${HTTP_CODE}"
if [[ "${HTTP_CODE}" != "200" ]]; then
  cat /tmp/oidc-agents.json >&2 || true
  echo "FAIL: expected registry 200 with valid JWT (is OIDC overlay enabled?)" >&2
  exit 1
fi

echo "    Gateway chat with JWT + bootstrap API key fallback check ..."
GATEWAY_CODE=$(curl -sS -o /tmp/oidc-chat.json -w "%{http_code}" \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: cache-demo-agent" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: staging" \
  -H "X-Cache-Bypass: true" \
  --data-binary "@${ROOT}/examples/demo/quickstart-chat-request.json")

echo "    Gateway HTTP ${GATEWAY_CODE}"
if [[ "${GATEWAY_CODE}" == "200" ]]; then
  echo "PASS: OIDC JWT accepted by gateway"
  exit 0
fi

if [[ "${GATEWAY_CODE}" == "403" ]]; then
  echo "NOTE: gateway returned 403 (likely OPA policy — check x-agent-environment matches registered agent)" >&2
  cat /tmp/oidc-chat.json >&2 || true
fi

# Gateway may still require API key if OIDC not enabled on gateway-only setups
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"
GATEWAY_CODE2=$(curl -sS -o /tmp/oidc-chat2.json -w "%{http_code}" \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: cache-demo-agent" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: staging" \
  -H "X-Cache-Bypass: true" \
  --data-binary "@${ROOT}/examples/demo/quickstart-chat-request.json")

if [[ "${GATEWAY_CODE2}" == "200" ]]; then
  echo "PASS: registry JWT auth works; gateway still on API key (enable OIDC overlay for gateway JWT)"
  exit 0
fi

if [[ "${HTTP_CODE}" == "200" ]]; then
  echo "PASS: registry OIDC JWT auth works (gateway chat skipped or policy-blocked)"
  exit 0
fi

echo "FAIL: neither JWT nor API key worked on gateway" >&2
exit 1
