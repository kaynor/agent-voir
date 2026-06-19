#!/usr/bin/env bash
# Pack a self-contained onebox zip for GitHub Release (no monorepo checkout needed).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?usage: pack-onebox-bundle.sh <tag> [ghcr.io/owner/repo]}"
IMAGE="${2:-ghcr.io/kaynor/agent-voir}"

BUNDLE_SRC="${ROOT}/deployments/docker/onebox-bundle"
STAGING="$(mktemp -d)"
OUT_DIR="${ROOT}/dist"
ZIP_NAME="agentvoir-onebox-${TAG}.zip"

cleanup() { rm -rf "$STAGING"; }
trap cleanup EXIT

cp -a "${BUNDLE_SRC}/." "${STAGING}/"
mkdir -p "${STAGING}/policies/opa"
cp "${ROOT}/policies/opa/agentvoir.rego" "${STAGING}/policies/opa/"
cp "${ROOT}/deployments/docker/run-agentvoir.sh" "${STAGING}/run-agentvoir.sh"

echo "${TAG}" > "${STAGING}/.version"
echo "${IMAGE}" > "${STAGING}/.image"

# Default image for this release (onebox.sh reads .image if present)
cat > "${STAGING}/.env.defaults" <<EOF
AGENTVOIR_IMAGE=${IMAGE}
AGENTVOIR_VERSION=${TAG}
EOF

chmod +x "${STAGING}/onebox.sh" "${STAGING}/onebox-smoke.sh" "${STAGING}/run-agentvoir.sh"

mkdir -p "${OUT_DIR}"
rm -f "${OUT_DIR}/${ZIP_NAME}"
(
  cd "${STAGING}"
  zip -qr "${OUT_DIR}/${ZIP_NAME}" .
)

cp "${ROOT}/deployments/docker/run-agentvoir.sh" "${OUT_DIR}/run-agentvoir.sh"
chmod +x "${OUT_DIR}/run-agentvoir.sh"

echo "Created ${OUT_DIR}/${ZIP_NAME}"
echo "Created ${OUT_DIR}/run-agentvoir.sh"
