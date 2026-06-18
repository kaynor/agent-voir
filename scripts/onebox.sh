#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/deployments/docker/.env.onebox"
COMPOSE_FILE="$ROOT/deployments/docker/docker-compose.onebox.yml"

cp -n "$ROOT/deployments/docker/.env.onebox.example" "$ENV_FILE" || true

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

GATEWAY_PORT="${AGENTVOIR_GATEWAY_PORT:-8080}"
REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
USAGE_PORT="${AGENTVOIR_USAGE_PORT:-8082}"
API_KEY="${GATEWAY_API_KEY:-agentvoir-onebox-key}"

echo "Pulling AgentVoir onebox images..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull

echo "Starting AgentVoir onebox..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d

echo ""
echo "AgentVoir onebox is starting. Only these ports are exposed on your machine:"
echo "  Gateway          http://localhost:${GATEWAY_PORT}"
echo "  Registry API     http://localhost:${REGISTRY_PORT}"
echo "  Token accounting http://localhost:${USAGE_PORT}"
echo ""
echo "Postgres, Redis, ClickHouse, and OPA run inside Docker only — no host port conflicts."
echo ""
echo "Smoke test:"
echo "  ./scripts/onebox-smoke.sh"
echo ""
echo "OpenAI-compatible client:"
echo "  export OPENAI_BASE_URL=http://localhost:${GATEWAY_PORT}/v1"
echo "  export OPENAI_API_KEY=${API_KEY}"
echo ""
echo "Stop:  docker compose --env-file $ENV_FILE -f $COMPOSE_FILE down"
echo "Logs:  docker compose --env-file $ENV_FILE -f $COMPOSE_FILE logs -f"
