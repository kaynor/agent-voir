#!/usr/bin/env bash
# Pack a self-contained onebox zip for GitHub Release (no monorepo checkout needed).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?usage: pack-onebox-bundle.sh <tag> [ghcr.io/owner/repo] [owner/repo]}"
IMAGE="${2:-ghcr.io/kaynor/agent-voir}"
GITHUB_REPO="${3:-kaynor/agent-voir}"

BUNDLE_SRC="${ROOT}/deployments/docker/onebox-bundle"
STAGING="$(mktemp -d)"
OUT_DIR="${ROOT}/dist"
ZIP_NAME="agentvoir-onebox-${TAG}.zip"

cleanup() { rm -rf "$STAGING"; }
trap cleanup EXIT

cp -a "${BUNDLE_SRC}/." "${STAGING}/"
mkdir -p "${STAGING}/policies/opa"
cp "${ROOT}/policies/opa/agentvoir.rego" "${STAGING}/policies/opa/"

patch_installer() {
  local dest="$1"
  sed \
    -e "s|__RELEASE_TAG__|${TAG}|g" \
    -e "s|__REPO__|${GITHUB_REPO}|g" \
    "${ROOT}/deployments/docker/run-agentvoir.sh" > "${dest}"
}

patch_installer "${STAGING}/run-agentvoir.sh"

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

patch_installer "${OUT_DIR}/run-agentvoir.sh"
chmod +x "${OUT_DIR}/run-agentvoir.sh"

if grep -q '__RELEASE_TAG__\|__REPO__' "${OUT_DIR}/run-agentvoir.sh"; then
  echo "ERROR: run-agentvoir.sh was not patched (placeholders remain)" >&2
  exit 1
fi

echo "Created ${OUT_DIR}/${ZIP_NAME}"
echo "Created ${OUT_DIR}/run-agentvoir.sh"
