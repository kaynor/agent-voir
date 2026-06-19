#!/usr/bin/env bash
# Download the AgentVoir onebox bundle from a GitHub Release and start the stack.
# No git clone required.
#
# Usage (use the release tag in the URL and pass it to bash):
#   curl -fsSL https://github.com/kaynor/agent-voir/releases/download/v0.2.7/run-agentvoir.sh | bash -s v0.2.7
#   ./run-agentvoir.sh
#   ./run-agentvoir.sh --smoke

set -euo pipefail

# Substituted when packed for each GitHub Release (__RELEASE_TAG__ / __REPO__).
DEFAULT_RELEASE_TAG="__RELEASE_TAG__"
DEFAULT_REPO="__REPO__"

REPO="${AGENTVOIR_REPO:-${DEFAULT_REPO}}"
VERSION="${AGENTVOIR_VERSION:-}"
INSTALL_DIR="${AGENTVOIR_INSTALL_DIR:-${HOME}/.agentvoir/onebox}"
RUN_SMOKE=0

if [[ "$REPO" == "__REPO__" ]]; then
  REPO="kaynor/agent-voir"
fi

# curl ... | bash -s v0.2.7  — required when the script is piped (not executed as a file).
if [[ -z "$VERSION" && $# -gt 0 && "$1" != --* ]]; then
  VERSION="$1"
  shift
fi

for arg in "$@"; do
  case "$arg" in
    --smoke) RUN_SMOKE=1 ;;
    -h|--help)
      cat <<EOF
AgentVoir onebox installer (Docker only)

  curl -fsSL https://github.com/${REPO}/releases/download/<tag>/run-agentvoir.sh | bash -s <tag>

  AGENTVOIR_VERSION=v0.2.7 bash -s        Only when piping: env must apply to bash, not curl
  AGENTVOIR_INSTALL_DIR=~/av $0           Install directory
  $0 --smoke                              Start stack and run health checks

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

if [[ -z "$VERSION" && "$DEFAULT_RELEASE_TAG" != "__RELEASE_TAG__" ]]; then
  VERSION="$DEFAULT_RELEASE_TAG"
fi

if [[ -z "$VERSION" ]]; then
  echo "ERROR: Pass the release tag when piping: curl ... | bash -s vX.Y.Z" >&2
  echo "  Example: curl -fsSL https://github.com/${REPO}/releases/download/v0.2.7/run-agentvoir.sh | bash -s v0.2.7" >&2
  exit 1
fi

BUNDLE_NAME="agentvoir-onebox-${VERSION}.zip"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BUNDLE_NAME}"

mkdir -p "${INSTALL_DIR}"
TMP_ZIP="$(mktemp)"
trap 'rm -f "$TMP_ZIP"' EXIT

echo "AgentVoir release: ${VERSION}"
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
