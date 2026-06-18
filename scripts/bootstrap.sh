#!/usr/bin/env bash
set -euo pipefail

cp -n .env.example .env || true
cp -n deployments/docker/.env.onebox.example deployments/docker/.env.onebox || true
make onebox-up

echo ""
echo "For isolated try-out (recommended): make onebox-up  (already started above)"
echo "For developer infra only:           make dev-up"
