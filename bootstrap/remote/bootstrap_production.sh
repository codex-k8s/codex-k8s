#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${1:-${ROOT_DIR}/../bootstrap.env}"

# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"

require_root
load_env_file "$ENV_FILE"
export BOOTSTRAP_ENV_FILE="$ENV_FILE"

steps=(
  "00_prepare_host.sh"
  "10_create_operator_user.sh"
  "20_install_k3s.sh"
  "40_install_platform_dependencies.sh"
  "65_harden_network_firewall.sh"
  "70_finalize_and_report.sh"
)

for step in "${steps[@]}"; do
  log "Run step: $step"
  bash "${ROOT_DIR}/${step}"
done

log "Remote bootstrap done"
