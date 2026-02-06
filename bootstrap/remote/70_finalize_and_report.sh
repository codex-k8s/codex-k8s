#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

STAGING_NAMESPACE="${STAGING_NAMESPACE:-codex-k8s-ai-staging}"
RUNNER_NAMESPACE="${RUNNER_NAMESPACE:-actions-runner-staging}"
RUNNER_SCALE_SET_NAME="${RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"

log "Staging namespace: ${STAGING_NAMESPACE}"
kubectl get pods -n "$STAGING_NAMESPACE" || true

log "Bootstrap finished. Recommended checks:"
log "  kubectl get pods -n ${STAGING_NAMESPACE}"
log "  kubectl get pods -n ${RUNNER_NAMESPACE}"
log "  helm list -n ${RUNNER_NAMESPACE} | grep ${RUNNER_SCALE_SET_NAME}"
log "  git push origin main  # should trigger staging deploy workflow"
