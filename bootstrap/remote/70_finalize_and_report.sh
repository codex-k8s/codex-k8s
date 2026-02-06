#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_RUNNER_NAMESPACE="${CODEXK8S_RUNNER_NAMESPACE:-actions-runner-staging}"
CODEXK8S_RUNNER_SCALE_SET_NAME="${CODEXK8S_RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"

log "Staging namespace: ${CODEXK8S_STAGING_NAMESPACE}"
kubectl get pods -n "$CODEXK8S_STAGING_NAMESPACE" || true

log "Bootstrap finished. Recommended checks:"
log "  kubectl get pods -n ${CODEXK8S_STAGING_NAMESPACE}"
log "  kubectl get pods -n ${CODEXK8S_RUNNER_NAMESPACE}"
log "  helm list -n ${CODEXK8S_RUNNER_NAMESPACE} | grep ${CODEXK8S_RUNNER_SCALE_SET_NAME}"
log "  git push origin main  # should trigger staging deploy workflow"
