#!/usr/bin/env bash
# Download the AgentVoir onebox bundle from a GitHub Release and start the stack.
# No git clone required.
#
# Usage:
#   curl -fsSL https://github.com/kaynor/agent-voir/releases/latest/download/run-agentvoir.sh | bash
#   AGENTVOIR_VERSION=v0.2.4 curl -fsSL .../run-agentvoir.sh | bash
#   ./run-agentvoir.sh                    # if you already downloaded this script
#   ./run-agentvoir.sh --smoke            # start and run health checks

set -euo pipefail

REPO="${AGENTVOIR_REPO:-kaynor/agent-voir}"
VERSION="${AGENTVOIR_VERSION:-}"
INSTALL_DIR="${AGENTVOIR_INSTALL_DIR:-${HOME}/.agentvoir/onebox}"
RUN_SMOKE=0

for arg in "$@"; do
  case "$arg" in
    --smoke) RUN_SMOKE=1 ;;
    -h|--help)
      cat <<EOF
AgentVoir onebox installer (Docker only)

  AGENTVOIR_VERSION=v0.2.4 $0     Pin release tag (default: latest GitHub release)
  AGENTVOIR_INSTALL_DIR=~/av $0   Where to unpack the bundle
  $0 --smoke                      Start stack and run health checks

Requires: docker, curl, unzip
EOF
      exit 0
      ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: docker is not installed." >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "ERROR: Docker daemon is not running." >&2
  exit 1
fi

resolve_latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n1
}

if [[ -z "$VERSION" ]]; then
  echo "Resolving latest release for ${REPO}..."
  VERSION="$(resolve_latest_version)"
  if [[ -z "$VERSION" ]]; then
    echo "ERROR: Could not resolve latest release. Set AGENTVOIR_VERSION=vX.Y.Z" >&2
    exit 1
  fi
fi

BUNDLE_NAME="agentvoir-onebox-${VERSION}.zip"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BUNDLE_NAME}"

mkdir -p "${INSTALL_DIR}"
TMP_ZIP="$(mktemp)"
trap 'rm -f "$TMP_ZIP"' EXIT

echo "Downloading ${DOWNLOAD_URL} ..."
if ! curl -fsSL -o "$TMP_ZIP" "$DOWNLOAD_URL"; then
  echo "ERROR: Download failed. Check AGENTVOIR_VERSION=${VERSION} and that the release includes ${BUNDLE_NAME}" >&2
  exit 1
fi

echo "Unpacking to ${INSTALL_DIR} ..."
rm -rf "${INSTALL_DIR:?}/"*
unzip -q -o "$TMP_ZIP" -d "${INSTALL_DIR}"

if [[ ! -x "${INSTALL_DIR}/onebox.sh" ]]; then
  chmod +x "${INSTALL_DIR}/onebox.sh" "${INSTALL_DIR}/onebox-smoke.sh" 2>/dev/null || true
fi

export AGENTVOIR_VERSION="${VERSION}"
"${INSTALL_DIR}/onebox.sh"

if [[ "$RUN_SMOKE" -eq 1 ]]; then
  echo "Waiting for services..."
  sleep 30
  "${INSTALL_DIR}/onebox-smoke.sh"
fi

echo ""
echo "Installed to: ${INSTALL_DIR}"
echo "Re-run:         ${INSTALL_DIR}/onebox.sh"
echo "Smoke test:     ${INSTALL_DIR}/onebox-smoke.sh"
