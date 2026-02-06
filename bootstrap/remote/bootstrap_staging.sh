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
  "30_install_network_stack.sh"
  "40_install_platform_dependencies.sh"
  "42_apply_network_policy_baseline.sh"
  "55_setup_internal_registry_and_build_image.sh"
  "45_configure_github_repo_ci.sh"
  "50_install_runner_arc.sh"
  "60_deploy_codex_k8s.sh"
  "65_harden_network_firewall.sh"
  "70_finalize_and_report.sh"
)

for step in "${steps[@]}"; do
  log "Run step: $step"
  bash "${ROOT_DIR}/${step}"
done

log "Remote bootstrap done"
