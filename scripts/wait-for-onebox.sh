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
TIMEOUT_SECONDS="${ONEBOX_WAIT_TIMEOUT:-120}"
INTERVAL_SECONDS="${ONEBOX_WAIT_INTERVAL:-2}"

wait_for() {
  local name="$1"
  local url="$2"
  local deadline=$((SECONDS + TIMEOUT_SECONDS))

  echo "Waiting for ${name} at ${url} ..."
  until curl -fsS "$url" >/dev/null 2>&1; do
    if (( SECONDS >= deadline )); then
      echo "Timed out waiting for ${name}" >&2
      exit 1
    fi
    sleep "$INTERVAL_SECONDS"
  done
  echo "  ${name} is ready"
}

wait_for "registry-api" "http://localhost:${REGISTRY_PORT}/healthz"
wait_for "token-accounting" "http://localhost:${USAGE_PORT}/healthz"
wait_for "gateway" "http://localhost:${GATEWAY_PORT}/healthz"

echo "Onebox stack is healthy."
