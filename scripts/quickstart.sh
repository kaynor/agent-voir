#!/usr/bin/env bash
# AgentVoir quickstart — end-to-end demo in one command.
#
# Starts onebox (unless already running), registers the demo agent, exercises
# the gateway cache (miss then hit), and prints recent usage events.
#
# Usage:
#   ./scripts/quickstart.sh           # start onebox + run demo
#   ./scripts/quickstart.sh --no-start  # assume onebox is already up

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/deployments/docker/.env.onebox"
COMPOSE_FILE="$ROOT/deployments/docker/docker-compose.onebox.yml"
MANIFEST="$ROOT/examples/agents/customer-support-agent.yaml"
CHAT_BODY="$ROOT/examples/demo/sample-chat-request.json"

START_STACK=1
if [[ "${1:-}" == "--no-start" ]]; then
  START_STACK=0
fi

cp -n "$ROOT/deployments/docker/.env.onebox.example" "$ENV_FILE" || true
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
USAGE_PORT="${AGENTVOIR_USAGE_PORT:-8082}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"
AGENT_ID="customer-support-agent"

banner() {
  echo ""
  echo "==> $1"
  echo ""
}

wait_for_url() {
  local url="$1"
  local label="$2"
  local i
  for i in $(seq 1 60); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "    $label ready"
      return 0
    fi
    sleep 2
  done
  echo "ERROR: $label did not become ready at $url" >&2
  exit 1
}

if [[ "$START_STACK" -eq 1 ]]; then
  banner "Starting AgentVoir onebox"
  "$ROOT/scripts/onebox.sh"
fi

banner "Waiting for services"
wait_for_url "http://localhost:${REGISTRY_PORT}/healthz" "Registry API"
wait_for_url "http://localhost:${USAGE_PORT}/healthz" "Token accounting"
wait_for_url "http://localhost:${GATEWAY_PORT}/healthz" "Gateway"

banner "Registering demo agent from manifest"
register_status="$(curl -sS -o /tmp/agentvoir-register.json -w "%{http_code}" \
  -X POST "http://localhost:${REGISTRY_PORT}/v1/agents/from-manifest" \
  -H "Content-Type: application/yaml" \
  --data-binary "@${MANIFEST}")"
case "$register_status" in
  201)
    echo "    Registered ${AGENT_ID} (201 Created)"
    head -c 300 /tmp/agentvoir-register.json
    echo ""
    ;;
  409)
    echo "    ${AGENT_ID} already registered (409 Conflict — OK for reruns)"
    ;;
  *)
    echo "ERROR: agent registration returned HTTP ${register_status}" >&2
    cat /tmp/agentvoir-register.json >&2
    exit 1
    ;;
esac

banner "Gateway chat completion — first request (expect cache miss)"
curl -sS -D /tmp/agentvoir-headers-1.txt -o /tmp/agentvoir-chat-1.json \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: ${AGENT_ID}" \
  -H "x-tenant-id: acme" \
  -H "x-user-id: quickstart-demo" \
  --data-binary "@${CHAT_BODY}"
cache1="$(grep -i '^x-cache-status:' /tmp/agentvoir-headers-1.txt | awk '{print $2}' | tr -d '\r' || true)"
echo "    x-cache-status: ${cache1:-unknown}"
head -c 200 /tmp/agentvoir-chat-1.json
echo "..."

banner "Gateway chat completion — second request (expect cache hit)"
curl -sS -D /tmp/agentvoir-headers-2.txt -o /tmp/agentvoir-chat-2.json \
  -X POST "http://localhost:${GATEWAY_PORT}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: ${AGENT_ID}" \
  -H "x-tenant-id: acme" \
  -H "x-user-id: quickstart-demo" \
  --data-binary "@${CHAT_BODY}"
cache2="$(grep -i '^x-cache-status:' /tmp/agentvoir-headers-2.txt | awk '{print $2}' | tr -d '\r' || true)"
echo "    x-cache-status: ${cache2:-unknown}"
head -c 200 /tmp/agentvoir-chat-2.json
echo "..."

if [[ "$cache1" == "miss" && "$cache2" == "hit" ]]; then
  echo "    Cache behavior: OK (miss → hit)"
elif [[ "$cache1" == "hit" && "$cache2" == "hit" ]]; then
  echo "    Cache behavior: prior run left cache warm (hit → hit)"
else
  echo "    Cache behavior: first=${cache1:-?} second=${cache2:-?} (check CACHE_MODE in onebox)"
fi

banner "Recent usage events for ${AGENT_ID}"
curl -sS "http://localhost:${USAGE_PORT}/v1/usage-events?agent_id=${AGENT_ID}&limit=5" \
  | python3 -m json.tool 2>/dev/null || curl -sS "http://localhost:${USAGE_PORT}/v1/usage-events?agent_id=${AGENT_ID}&limit=5"

banner "Quickstart complete"
cat <<EOF

Gateway:  http://localhost:${GATEWAY_PORT}
Registry: http://localhost:${REGISTRY_PORT}
Usage:    http://localhost:${USAGE_PORT}

OpenAI client:
  export OPENAI_BASE_URL=http://localhost:${GATEWAY_PORT}/v1
  export OPENAI_API_KEY=${API_KEY}

Demo walkthrough: docs/demo/README.md
Stop stack: docker compose --env-file ${ENV_FILE} -f ${COMPOSE_FILE} down
EOF
