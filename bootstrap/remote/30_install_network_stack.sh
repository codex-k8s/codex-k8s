#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_RUNNER_NAMESPACE="${CODEXK8S_RUNNER_NAMESPACE:-actions-runner-staging}"

log "Create base namespaces"
kubectl get ns actions-runner-system >/dev/null 2>&1 || kubectl create ns actions-runner-system
kubectl get ns "$CODEXK8S_RUNNER_NAMESPACE" >/dev/null 2>&1 || kubectl create ns "$CODEXK8S_RUNNER_NAMESPACE"
kubectl get ns "$CODEXK8S_STAGING_NAMESPACE" >/dev/null 2>&1 || kubectl create ns "$CODEXK8S_STAGING_NAMESPACE"
kubectl get ns cert-manager >/dev/null 2>&1 || kubectl create ns cert-manager

log "Install ingress-nginx controller (idempotent apply)"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

log "Install cert-manager (idempotent apply)"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml
