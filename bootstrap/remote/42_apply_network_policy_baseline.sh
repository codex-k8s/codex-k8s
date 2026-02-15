#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

REPO_DIR="$(repo_dir)"
CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_NETWORK_POLICY_BASELINE="${CODEXK8S_NETWORK_POLICY_BASELINE:-true}"
CODEXK8S_PLATFORM_MCP_PORT="${CODEXK8S_PLATFORM_MCP_PORT:-8081}"
CODEXK8S_K8S_API_CIDR="${CODEXK8S_K8S_API_CIDR:-0.0.0.0/0}"
CODEXK8S_K8S_API_PORT="${CODEXK8S_K8S_API_PORT:-6443}"

if [ "$CODEXK8S_NETWORK_POLICY_BASELINE" != "true" ]; then
  log "Skip network policy baseline: CODEXK8S_NETWORK_POLICY_BASELINE=${CODEXK8S_NETWORK_POLICY_BASELINE}"
  exit 0
fi

log "Apply network policy baseline for platform namespace ${CODEXK8S_STAGING_NAMESPACE}"
export CODEXK8S_STAGING_NAMESPACE \
  CODEXK8S_NETWORK_POLICY_BASELINE \
  CODEXK8S_PLATFORM_MCP_PORT \
  CODEXK8S_K8S_API_CIDR \
  CODEXK8S_K8S_API_PORT
bash "${REPO_DIR}/deploy/scripts/apply_network_policy_baseline.sh"
