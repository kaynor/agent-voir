#!/bin/sh
set -eu

wait_for() {
  url=$1
  name=$2
  for _ in $(seq 1 60); do
    if wget -q -O - "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "timeout waiting for ${name} at ${url}" >&2
  return 1
}

shutdown() {
  kill "$REGISTRY_PID" "$ACCOUNTING_PID" "$GATEWAY_PID" 2>/dev/null || true
  wait "$REGISTRY_PID" "$ACCOUNTING_PID" "$GATEWAY_PID" 2>/dev/null || true
}

trap shutdown TERM INT

agentvoir-registry-api &
REGISTRY_PID=$!

agentvoir-token-accounting &
ACCOUNTING_PID=$!

wait_for "http://127.0.0.1:8081/healthz" "registry-api"
wait_for "http://127.0.0.1:8082/healthz" "token-accounting"

agentvoir-gateway &
GATEWAY_PID=$!

wait_for "http://127.0.0.1:8080/healthz" "gateway"

while kill -0 "$REGISTRY_PID" 2>/dev/null \
  && kill -0 "$ACCOUNTING_PID" 2>/dev/null \
  && kill -0 "$GATEWAY_PID" 2>/dev/null; do
  sleep 2
done

echo "an AgentVoir process exited; shutting down" >&2
shutdown
exit 1
