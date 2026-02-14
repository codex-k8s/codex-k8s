#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SERVICES_PATH="${CODEXK8S_SERVICES_CONFIG_PATH:-${ROOT_DIR}/services.yaml}"
TARGET_ENVIRONMENT="${CODEXK8S_DEPLOY_ENVIRONMENT:-ai-staging}"
BOOTSTRAP_BIN="${CODEXK8S_BOOTSTRAP_BIN:-${ROOT_DIR}/.bin/codex-bootstrap}"

ensure_bootstrap_cli() {
  if [ -x "${BOOTSTRAP_BIN}" ]; then
    return 0
  fi
  if ! command -v go >/dev/null 2>&1; then
    echo "Missing Go toolchain. Install Go >=1.25 or set CODEXK8S_BOOTSTRAP_BIN to a prebuilt codex-bootstrap binary." >&2
    return 1
  fi
  mkdir -p "$(dirname "${BOOTSTRAP_BIN}")"
  (
    cd "${ROOT_DIR}"
    go build -o "${BOOTSTRAP_BIN}" ./bin/codex-bootstrap
  )
}

ensure_bootstrap_cli

args=(
  reconcile
  --services "${SERVICES_PATH}"
  --environment "${TARGET_ENVIRONMENT}"
  --no-prompt
)
if [ -n "${CODEXK8S_KUBECONFIG:-}" ]; then
  args+=(--kubeconfig "${CODEXK8S_KUBECONFIG}")
fi

"${BOOTSTRAP_BIN}" "${args[@]}"
