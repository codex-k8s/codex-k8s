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
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:5000}"

log "Staging namespace: ${CODEXK8S_STAGING_NAMESPACE}"
kubectl get pods -n "$CODEXK8S_STAGING_NAMESPACE" || true
log "Internal registry deployment (no auth, node loopback ${CODEXK8S_INTERNAL_REGISTRY_HOST}):"
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get deploy "${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" || true

log "Bootstrap finished. Recommended checks:"
log "  kubectl get pods -n ${CODEXK8S_STAGING_NAMESPACE}"
log "  kubectl get deploy -n ${CODEXK8S_STAGING_NAMESPACE} ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}"
log "  sudo cat /etc/rancher/k3s/registries.yaml"
log "  kubectl get pods -n ${CODEXK8S_RUNNER_NAMESPACE}"
log "  helm list -n ${CODEXK8S_RUNNER_NAMESPACE} | grep ${CODEXK8S_RUNNER_SCALE_SET_NAME}"
log "  git push origin main  # should trigger staging deploy workflow"
