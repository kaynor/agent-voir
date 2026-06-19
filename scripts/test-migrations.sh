#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DSN="${POSTGRES_DSN:-postgres://agentvoir:agentvoir@localhost:5432/agentvoir?sslmode=disable}"
MIGRATIONS_DIR="${ROOT}/db/migrations/postgres"

export MIGRATIONS_DIR
export POSTGRES_DSN="$DSN"

echo "==> Applying migrations"
(cd "${ROOT}/apps/registry-api" && go run ./cmd/migrate)

echo "==> Rolling back migrations"
(cd "${ROOT}/apps/registry-api" && go run ./cmd/migrate-down)

echo "==> Re-applying migrations"
(cd "${ROOT}/apps/registry-api" && go run ./cmd/migrate)

echo "Migration up/down test passed."
