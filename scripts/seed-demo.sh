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

REGISTRY_PORT="${AGENTVOIR_REGISTRY_PORT:-8081}"
REGISTRY_URL="http://localhost:${REGISTRY_PORT}"

echo "==> Seeding demo agents via manifest import"
for manifest in "${ROOT}/examples/agents/"*.yaml; do
  [[ -f "$manifest" ]] || continue
  echo "Registering $(basename "$manifest") ..."
  curl -fsS -X POST "${REGISTRY_URL}/v1/agents/from-manifest" \
    --data-binary "@${manifest}" \
    -H "Content-Type: application/x-yaml" || true
  echo
done

echo "==> Demo seed complete (conflicts are ignored if agents already exist)"
