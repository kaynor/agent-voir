#!/usr/bin/env bash
set -euo pipefail

cp -n .env.example .env || true
make dev-up

echo "AgentVoir local dependencies are starting."
echo "Registry API: make run-api"
echo "Gateway:      make run-gateway"
echo "Web console:  make run-web"
