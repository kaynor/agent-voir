#!/usr/bin/env bash
# Load dummy proxy events into the gateway for the Live Proxy Flow dashboard.
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
GATEWAY_URL="${GATEWAY_URL:-http://localhost:${GATEWAY_PORT}}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"
COUNT="${1:-80}"
RESET="${RESET:-true}"

echo "==> Seeding ${COUNT} dummy proxy-event traces at ${GATEWAY_URL}"
echo "    (set COUNT=120 RESET=false to append without clearing)"

payload=$(printf '{"count":%s,"reset":%s}' "$COUNT" "$RESET")

response=$(curl -fsS -X POST "${GATEWAY_URL}/v1/proxy-events/seed" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$payload")

echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"

echo
echo "==> Sample rows"
curl -fsS "${GATEWAY_URL}/v1/proxy-events?limit=5" | python3 -m json.tool 2>/dev/null | head -40

echo
echo "==> Open Live Proxy Flow: http://localhost:3000/live"
echo "    (run 'make run-web' in another terminal if the console is not up)"
