#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_RUNNER_NAMESPACE="${CODEXK8S_RUNNER_NAMESPACE:-actions-runner-staging}"
CODEXK8S_INGRESS_READY_TIMEOUT="${CODEXK8S_INGRESS_READY_TIMEOUT:-1200s}"
CODEXK8S_CERT_MANAGER_READY_TIMEOUT="${CODEXK8S_CERT_MANAGER_READY_TIMEOUT:-1200s}"
CODEXK8S_INGRESS_HOST_NETWORK="${CODEXK8S_INGRESS_HOST_NETWORK:-true}"

log "Create base namespaces"
kubectl get ns actions-runner-system >/dev/null 2>&1 || kubectl create ns actions-runner-system
kubectl get ns "$CODEXK8S_RUNNER_NAMESPACE" >/dev/null 2>&1 || kubectl create ns "$CODEXK8S_RUNNER_NAMESPACE"
kubectl get ns "$CODEXK8S_STAGING_NAMESPACE" >/dev/null 2>&1 || kubectl create ns "$CODEXK8S_STAGING_NAMESPACE"
kubectl get ns cert-manager >/dev/null 2>&1 || kubectl create ns cert-manager

log "Install ingress-nginx controller (idempotent apply)"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
if [ "$CODEXK8S_INGRESS_HOST_NETWORK" = "true" ]; then
  kubectl -n ingress-nginx patch deployment ingress-nginx-controller --type=merge \
    -p '{"spec":{"template":{"spec":{"hostNetwork":true,"dnsPolicy":"ClusterFirstWithHostNet"}}}}'
  # In hostNetwork mode ingress is reachable via host :80/:443 directly, NodePorts must stay closed.
  kubectl -n ingress-nginx patch service ingress-nginx-controller --type=merge \
    -p '{"spec":{"type":"ClusterIP","externalTrafficPolicy":null,"healthCheckNodePort":null}}'
fi
kubectl -n ingress-nginx rollout status deployment/ingress-nginx-controller --timeout="${CODEXK8S_INGRESS_READY_TIMEOUT}"

log "Install cert-manager (idempotent apply)"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml
kubectl -n cert-manager rollout status deployment/cert-manager --timeout="${CODEXK8S_CERT_MANAGER_READY_TIMEOUT}"
kubectl -n cert-manager rollout status deployment/cert-manager-webhook --timeout="${CODEXK8S_CERT_MANAGER_READY_TIMEOUT}"
kubectl -n cert-manager rollout status deployment/cert-manager-cainjector --timeout="${CODEXK8S_CERT_MANAGER_READY_TIMEOUT}"
