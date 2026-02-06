#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

if [ "${ENABLE_GITHUB_RUNNER:-true}" != "true" ]; then
  log "ENABLE_GITHUB_RUNNER=false, skip ARC/runner setup"
  exit 0
fi

: "${GITHUB_PAT:?GITHUB_PAT is required}"
: "${GITHUB_REPO:?GITHUB_REPO is required}"
RUNNER_MIN="${RUNNER_MIN:-0}"
RUNNER_MAX="${RUNNER_MAX:-2}"
RUNNER_NAMESPACE="${RUNNER_NAMESPACE:-actions-runner-staging}"
RUNNER_SCALE_SET_NAME="${RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"
RUNNER_IMAGE="${RUNNER_IMAGE:-ghcr.io/actions/actions-runner:latest}"

REPO_DIR="$(repo_dir)"

kubectl apply -f "${REPO_DIR}/deploy/runner/namespace.yaml"
export RUNNER_NAMESPACE
envsubst < "${REPO_DIR}/deploy/runner/runner-namespace.yaml.tpl" | kubectl apply -f -

kubectl -n "${RUNNER_NAMESPACE}" create secret generic gha-runner-scale-set-secret \
  --from-literal=github_token="${GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "Install ARC runner scale-set controller via Helm"
helm upgrade --install gha-rs-controller oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller \
  --namespace actions-runner-system \
  --create-namespace

log "Install ARC runner scale set via Helm"
export GITHUB_REPO RUNNER_MIN RUNNER_MAX RUNNER_IMAGE RUNNER_SCALE_SET_NAME
VALUES_FILE="$(mktemp)"
envsubst < "${REPO_DIR}/deploy/runner/values-ai-staging.yaml.tpl" > "${VALUES_FILE}"
helm upgrade --install "${RUNNER_SCALE_SET_NAME}" oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  --namespace "${RUNNER_NAMESPACE}" \
  --create-namespace \
  -f "${VALUES_FILE}" \
  --wait
rm -f "${VALUES_FILE}"
