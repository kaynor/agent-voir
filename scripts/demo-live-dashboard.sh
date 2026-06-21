#!/usr/bin/env bash
# End-to-end verification: onebox + seed dummy live events + smoke API checks.
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
GATEWAY_URL="http://localhost:${GATEWAY_PORT}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"

echo "==> Waiting for gateway"
chmod +x "${ROOT}/scripts/wait-for-onebox.sh"
"${ROOT}/scripts/wait-for-onebox.sh"

echo "==> Seeding live dashboard dummy data"
chmod +x "${ROOT}/scripts/seed-live-events.sh"
COUNT=100 RESET=true "${ROOT}/scripts/seed-live-events.sh" 100

echo
echo "==> GET /v1/proxy-events/metrics"
curl -fsS "${GATEWAY_URL}/v1/proxy-events/metrics" | python3 -m json.tool

echo
echo "==> GET /v1/proxy-events (first trace for drilldown test)"
TRACE_ID=$(curl -fsS "${GATEWAY_URL}/v1/proxy-events?limit=1" | python3 -c "import sys,json; print(json.load(sys.stdin)['events'][0]['trace_id'])")
echo "trace_id=${TRACE_ID}"
curl -fsS "${GATEWAY_URL}/v1/traces/${TRACE_ID}" | python3 -m json.tool | head -30

echo
echo "PASS: Live dashboard backend ready"
echo
echo "Next steps:"
echo "  1. make run-web"
echo "  2. Open http://localhost:3000/live"
echo "  3. Optional: send real traffic — make quickstart (rows appear via gateway recorder)"
