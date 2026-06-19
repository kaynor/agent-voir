#!/usr/bin/env bash
# Start AgentVoir onebox from the release bundle (Docker only — no repo clone).
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${DIR}/docker-compose.yml"

if [[ -f "${DIR}/.env.defaults" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${DIR}/.env.defaults"
  set +a
fi

if [[ -f "${DIR}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${DIR}/.env"
  set +a
fi

if [[ -f "${DIR}/.image" ]]; then
  AGENTVOIR_IMAGE="$(tr -d '[:space:]' < "${DIR}/.image")"
fi
export AGENTVOIR_IMAGE="${AGENTVOIR_IMAGE:-ghcr.io/kaynor/agent-voir}"
if [[ -z "${AGENTVOIR_VERSION:-}" && -f "${DIR}/.version" ]]; then
  AGENTVOIR_VERSION="$(tr -d '[:space:]' < "${DIR}/.version")"
fi
export AGENTVOIR_VERSION="${AGENTVOIR_VERSION:-latest}"

GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
USAGE_PORT="${AGENTVOIR_USAGE_PORT:-8082}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"

if ! docker info >/dev/null 2>&1; then
  echo "ERROR: Docker is not running. Start Docker Desktop or the Docker daemon." >&2
  exit 1
fi

echo "Pulling images (AgentVoir ${AGENTVOIR_IMAGE}:${AGENTVOIR_VERSION})..."
docker compose -f "${COMPOSE_FILE}" pull

echo "Starting AgentVoir onebox..."
docker compose -f "${COMPOSE_FILE}" up -d

echo ""
echo "AgentVoir onebox is starting."
echo "  Gateway          http://localhost:${GATEWAY_PORT}"
echo "  Registry API     http://localhost:${REGISTRY_PORT}"
echo "  Token accounting http://localhost:${USAGE_PORT}"
echo ""
echo "API key: ${API_KEY}"
echo ""
echo "Wait ~30s, then:  ./onebox-smoke.sh"
echo "Stop:             docker compose -f ${COMPOSE_FILE} down"
echo "Stop + wipe data: docker compose -f ${COMPOSE_FILE} down -v"
