#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

if [ "${CODEXK8S_ENABLE_GITHUB_RUNNER:-false}" != "true" ]; then
  log "CODEXK8S_ENABLE_GITHUB_RUNNER=false, skip ARC/runner setup"
  exit 0
fi

: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"
: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"

require_cmd kubectl
require_cmd helm
require_cmd envsubst

CODEXK8S_RUNNER_MIN="${CODEXK8S_RUNNER_MIN:-1}"
CODEXK8S_RUNNER_MAX="${CODEXK8S_RUNNER_MAX:-2}"
CODEXK8S_RUNNER_NAMESPACE="${CODEXK8S_RUNNER_NAMESPACE:-actions-runner-staging}"
CODEXK8S_RUNNER_SCALE_SET_NAME="${CODEXK8S_RUNNER_SCALE_SET_NAME:-${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}}"
CODEXK8S_RUNNER_IMAGE="${CODEXK8S_RUNNER_IMAGE:-ghcr.io/actions/actions-runner:latest}"
CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_WORKER_RUN_ROLE_NAME="${CODEXK8S_WORKER_RUN_ROLE_NAME:-}"
CODEXK8S_HELM_TIMEOUT="${CODEXK8S_HELM_TIMEOUT:-20m}"

REPO_DIR="$(repo_dir)"

apply_actions_runner_system_ns() {
  local file="${REPO_DIR}/deploy/runner/namespace.yaml"
  [ -f "${file}" ] || die "ARC namespace manifest not found: ${file}"
  kubectl apply -f "${file}"
}

apply_runner_namespace() {
  local tpl="${REPO_DIR}/deploy/runner/runner-namespace.yaml.tpl"
  [ -f "${tpl}" ] || die "ARC runner namespace template not found: ${tpl}"
  export CODEXK8S_RUNNER_NAMESPACE
  envsubst < "${tpl}" | kubectl apply -f -
}

apply_staging_deployer_rbac() {
  local tpl="${REPO_DIR}/deploy/runner/staging-deployer-rbac.yaml.tpl"
  [ -f "${tpl}" ] || die "ARC staging deployer RBAC template not found: ${tpl}"
  export CODEXK8S_STAGING_NAMESPACE CODEXK8S_RUNNER_NAMESPACE CODEXK8S_RUNNER_SCALE_SET_NAME CODEXK8S_WORKER_RUN_ROLE_NAME
  envsubst < "${tpl}" | kubectl apply -f -
}

apply_actions_runner_system_ns
apply_runner_namespace

kubectl -n "${CODEXK8S_RUNNER_NAMESPACE}" create secret generic gha-runner-scale-set-secret \
  --from-literal=github_token="${CODEXK8S_GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "Install ARC runner scale-set controller via Helm"
helm upgrade --install gha-rs-controller oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller \
  --namespace actions-runner-system \
  --create-namespace \
  --wait \
  --timeout "${CODEXK8S_HELM_TIMEOUT}"

log "Install ARC runner scale set via Helm"
VALUES_FILE="$(mktemp)"
values_tpl="${REPO_DIR}/deploy/runner/values-ai-staging.yaml.tpl"
[ -f "${values_tpl}" ] || die "ARC values template not found: ${values_tpl}"
export CODEXK8S_GITHUB_REPO CODEXK8S_RUNNER_MIN CODEXK8S_RUNNER_MAX CODEXK8S_RUNNER_IMAGE CODEXK8S_RUNNER_SCALE_SET_NAME
envsubst < "${values_tpl}" > "${VALUES_FILE}"
helm upgrade --install "${CODEXK8S_RUNNER_SCALE_SET_NAME}" oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  --namespace "${CODEXK8S_RUNNER_NAMESPACE}" \
  --create-namespace \
  -f "${VALUES_FILE}" \
  --wait \
  --timeout "${CODEXK8S_HELM_TIMEOUT}"
rm -f "${VALUES_FILE}"

log "Grant staging deploy RBAC to runner service account"
apply_staging_deployer_rbac
