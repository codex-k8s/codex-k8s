#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing command: $1" >&2
    exit 1
  }
}

render_registry_template() {
  local tpl="$1"
  sed \
    -e "s|\${CODEXK8S_STAGING_NAMESPACE}|${CODEXK8S_STAGING_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_INTERNAL_REGISTRY_SERVICE}|${CODEXK8S_INTERNAL_REGISTRY_SERVICE}|g" \
    -e "s|\${CODEXK8S_INTERNAL_REGISTRY_PORT}|${CODEXK8S_INTERNAL_REGISTRY_PORT}|g" \
    -e "s|\${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE}|${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE}|g" \
    "$tpl"
}

render_kaniko_job_template() {
  local tpl="$1"
  sed \
    -e "s|\${CODEXK8S_STAGING_NAMESPACE}|${CODEXK8S_STAGING_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_INTERNAL_REGISTRY_HOST}|${CODEXK8S_INTERNAL_REGISTRY_HOST}|g" \
    -e "s|\${CODEXK8S_GITHUB_REPO}|${CODEXK8S_GITHUB_REPO}|g" \
    -e "s|\${CODEXK8S_BUILD_REF}|${CODEXK8S_BUILD_REF}|g" \
    -e "s|\${CODEXK8S_KANIKO_JOB_NAME}|${CODEXK8S_KANIKO_JOB_NAME}|g" \
    -e "s|\${CODEXK8S_KANIKO_COMPONENT}|${CODEXK8S_KANIKO_COMPONENT}|g" \
    -e "s|\${CODEXK8S_KANIKO_CONTEXT}|${CODEXK8S_KANIKO_CONTEXT}|g" \
    -e "s|\${CODEXK8S_KANIKO_DOCKERFILE}|${CODEXK8S_KANIKO_DOCKERFILE}|g" \
    -e "s|\${CODEXK8S_KANIKO_DESTINATION_LATEST}|${CODEXK8S_KANIKO_DESTINATION_LATEST}|g" \
    -e "s|\${CODEXK8S_KANIKO_DESTINATION_SHA}|${CODEXK8S_KANIKO_DESTINATION_SHA}|g" \
    "$tpl"
}

normalize_sha_tag() {
  local ref="$1"
  if [[ "$ref" =~ ^[0-9a-fA-F]{12,40}$ ]]; then
    printf '%s' "${ref:0:12}" | tr '[:upper:]' '[:lower:]'
    return 0
  fi
  printf '%s' "$ref" | sha256sum | awk '{print $1}' | cut -c1-12
}

build_with_kaniko() {
  local job_name="$1"
  local component="$2"
  local context="$3"
  local dockerfile="$4"
  local image_latest="$5"
  local image_sha="$6"

  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete job "${job_name}" --ignore-not-found=true >/dev/null 2>&1 || true
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=delete "job/${job_name}" --timeout=120s >/dev/null 2>&1 || true

  CODEXK8S_KANIKO_JOB_NAME="${job_name}"
  CODEXK8S_KANIKO_COMPONENT="${component}"
  CODEXK8S_KANIKO_CONTEXT="${context}"
  CODEXK8S_KANIKO_DOCKERFILE="${dockerfile}"
  CODEXK8S_KANIKO_DESTINATION_LATEST="${image_latest}"
  CODEXK8S_KANIKO_DESTINATION_SHA="${image_sha}"
  export CODEXK8S_KANIKO_JOB_NAME CODEXK8S_KANIKO_COMPONENT CODEXK8S_KANIKO_CONTEXT CODEXK8S_KANIKO_DOCKERFILE CODEXK8S_KANIKO_DESTINATION_LATEST CODEXK8S_KANIKO_DESTINATION_SHA

  render_kaniko_job_template "${ROOT_DIR}/deploy/base/kaniko/kaniko-build-job.yaml.tpl" | kubectl apply -f -

  if ! kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=condition=complete "job/${job_name}" --timeout="${CODEXK8S_KANIKO_TIMEOUT}"; then
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get pods -l "job-name=${job_name}" -o wide || true
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" logs "job/${job_name}" --all-containers=true --tail=200 || true
    exit 1
  fi
}

require_cmd kubectl
require_cmd sed
require_cmd sha256sum
require_cmd awk
require_cmd tr
require_cmd cut

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_PORT="${CODEXK8S_INTERNAL_REGISTRY_PORT:-5000}"
CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE="${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE:-20Gi}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}}"
CODEXK8S_KANIKO_TIMEOUT="${CODEXK8S_KANIKO_TIMEOUT:-1800s}"
CODEXK8S_ENSURE_REGISTRY="${CODEXK8S_ENSURE_REGISTRY:-true}"
CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT:-600s}"
CODEXK8S_BUILD_REF="${CODEXK8S_BUILD_REF:-main}"
: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"

CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/api-gateway}"
CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/control-plane}"
CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/worker}"
CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/web-console}"

CODEXK8S_BUILD_SHA="$(normalize_sha_tag "$CODEXK8S_BUILD_REF")"

CODEXK8S_API_GATEWAY_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_API_GATEWAY_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_CONTROL_PLANE_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_CONTROL_PLANE_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_WORKER_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_WORKER_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_WEB_CONSOLE_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_WEB_CONSOLE_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

if [ "$CODEXK8S_ENSURE_REGISTRY" = "true" ]; then
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete statefulset "${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --ignore-not-found=true >/dev/null 2>&1 || true
  render_registry_template "${ROOT_DIR}/deploy/base/registry/registry.yaml.tpl" | kubectl apply -f -
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete service "${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --ignore-not-found=true >/dev/null 2>&1 || true
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status "deployment/${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --timeout="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT}"
fi

kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" create secret generic codex-k8s-git-token \
  --from-literal=token="${CODEXK8S_GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

build_with_kaniko \
  "codex-k8s-kaniko-api-gateway-${CODEXK8S_BUILD_SHA}" \
  "api-gateway" \
  "dir:///workspace" \
  "/workspace/services/external/api-gateway/Dockerfile" \
  "${CODEXK8S_API_GATEWAY_IMAGE_LATEST}" \
  "${CODEXK8S_API_GATEWAY_IMAGE_SHA}"

build_with_kaniko \
  "codex-k8s-kaniko-control-plane-${CODEXK8S_BUILD_SHA}" \
  "control-plane" \
  "dir:///workspace" \
  "/workspace/services/internal/control-plane/Dockerfile" \
  "${CODEXK8S_CONTROL_PLANE_IMAGE_LATEST}" \
  "${CODEXK8S_CONTROL_PLANE_IMAGE_SHA}"

build_with_kaniko \
  "codex-k8s-kaniko-worker-${CODEXK8S_BUILD_SHA}" \
  "worker" \
  "dir:///workspace" \
  "/workspace/services/jobs/worker/Dockerfile" \
  "${CODEXK8S_WORKER_IMAGE_LATEST}" \
  "${CODEXK8S_WORKER_IMAGE_SHA}"

build_with_kaniko \
  "codex-k8s-kaniko-web-console-${CODEXK8S_BUILD_SHA}" \
  "web-console" \
  "dir:///workspace/services/staff/web-console" \
  "/workspace/services/staff/web-console/Dockerfile" \
  "${CODEXK8S_WEB_CONSOLE_IMAGE_LATEST}" \
  "${CODEXK8S_WEB_CONSOLE_IMAGE_SHA}"

echo "Internal images build completed:"
echo "  api-gateway: ${CODEXK8S_API_GATEWAY_IMAGE_LATEST}"
echo "  api-gateway: ${CODEXK8S_API_GATEWAY_IMAGE_SHA}"
echo "  control-plane: ${CODEXK8S_CONTROL_PLANE_IMAGE_LATEST}"
echo "  control-plane: ${CODEXK8S_CONTROL_PLANE_IMAGE_SHA}"
echo "  worker: ${CODEXK8S_WORKER_IMAGE_LATEST}"
echo "  worker: ${CODEXK8S_WORKER_IMAGE_SHA}"
echo "  web-console(dev target): ${CODEXK8S_WEB_CONSOLE_IMAGE_LATEST}"
echo "  web-console(dev target): ${CODEXK8S_WEB_CONSOLE_IMAGE_SHA}"
