#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

log "Skip: network stack provisioning moved to Go runtime deploy prerequisites mode"
log "Use: go run ./services/internal/control-plane/cmd/runtime-deploy --prerequisites-only ..."
