#!/usr/bin/env bash
set -euo pipefail

cp -n .env.example .env || true
cp -n deployments/docker/.env.onebox.example deployments/docker/.env.onebox || true
./scripts/onebox.sh

echo ""
echo "For isolated try-out (recommended): ./scripts/onebox.sh  (already started above)"
echo "For developer infra only:           make dev-up"
