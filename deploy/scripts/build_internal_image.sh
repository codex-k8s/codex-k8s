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
    -e "s|\${CODEXK8S_KANIKO_CACHE_ENABLED}|${CODEXK8S_KANIKO_CACHE_ENABLED}|g" \
    -e "s|\${CODEXK8S_KANIKO_CACHE_REPO}|${CODEXK8S_KANIKO_CACHE_REPO}|g" \
    -e "s|\${CODEXK8S_KANIKO_CACHE_TTL}|${CODEXK8S_KANIKO_CACHE_TTL}|g" \
    -e "s|\${CODEXK8S_KANIKO_CACHE_COMPRESSED}|${CODEXK8S_KANIKO_CACHE_COMPRESSED}|g" \
    -e "s|\${CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU}|${CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU}|g" \
    -e "s|\${CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY}|${CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY}|g" \
    -e "s|\${CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU}|${CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU}|g" \
    -e "s|\${CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY}|${CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY}|g" \
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

build_component() {
  local component="$1"
  case "$component" in
    api-gateway)
      build_with_kaniko \
        "codex-k8s-kaniko-api-gateway-${CODEXK8S_BUILD_SHA}" \
        "api-gateway" \
        "dir:///workspace" \
        "/workspace/services/external/api-gateway/Dockerfile" \
        "${CODEXK8S_API_GATEWAY_IMAGE_LATEST}" \
        "${CODEXK8S_API_GATEWAY_IMAGE_SHA}"
      ;;
    control-plane)
      build_with_kaniko \
        "codex-k8s-kaniko-control-plane-${CODEXK8S_BUILD_SHA}" \
        "control-plane" \
        "dir:///workspace" \
        "/workspace/services/internal/control-plane/Dockerfile" \
        "${CODEXK8S_CONTROL_PLANE_IMAGE_LATEST}" \
        "${CODEXK8S_CONTROL_PLANE_IMAGE_SHA}"
      ;;
    worker)
      build_with_kaniko \
        "codex-k8s-kaniko-worker-${CODEXK8S_BUILD_SHA}" \
        "worker" \
        "dir:///workspace" \
        "/workspace/services/jobs/worker/Dockerfile" \
        "${CODEXK8S_WORKER_IMAGE_LATEST}" \
        "${CODEXK8S_WORKER_IMAGE_SHA}"
      ;;
    agent-runner)
      build_with_kaniko \
        "codex-k8s-kaniko-agent-runner-${CODEXK8S_BUILD_SHA}" \
        "agent-runner" \
        "dir:///workspace" \
        "/workspace/services/jobs/agent-runner/Dockerfile" \
        "${CODEXK8S_AGENT_RUNNER_IMAGE_LATEST}" \
        "${CODEXK8S_AGENT_RUNNER_IMAGE_SHA}"
      ;;
    web-console)
      build_with_kaniko \
        "codex-k8s-kaniko-web-console-${CODEXK8S_BUILD_SHA}" \
        "web-console" \
        "dir:///workspace/services/staff/web-console" \
        "/workspace/services/staff/web-console/Dockerfile" \
        "${CODEXK8S_WEB_CONSOLE_IMAGE_LATEST}" \
        "${CODEXK8S_WEB_CONSOLE_IMAGE_SHA}"
      ;;
    *)
      echo "Unknown component in CODEXK8S_BUILD_COMPONENTS: ${component}" >&2
      return 1
      ;;
  esac
}

wait_for_kaniko_job() {
  local job_name="$1"
  local wait_status=0

  if timeout "${CODEXK8S_KANIKO_TIMEOUT}" bash -s -- "${CODEXK8S_STAGING_NAMESPACE}" "${job_name}" <<'EOF'
set -euo pipefail

namespace="$1"
job_name="$2"

while true; do
  complete_status="$(kubectl -n "${namespace}" get "job/${job_name}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null || true)"
  failed_status="$(kubectl -n "${namespace}" get "job/${job_name}" -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null || true)"

  if [ "${complete_status}" = "True" ]; then
    exit 0
  fi
  if [ "${failed_status}" = "True" ]; then
    exit 10
  fi

  sleep 5
done
EOF
  then
    return 0
  else
    wait_status=$?
  fi

  case "${wait_status}" in
    10)
      echo "Kaniko job ${job_name} failed before completion" >&2
      ;;
    124)
      echo "Timed out waiting for Kaniko job ${job_name} (timeout=${CODEXK8S_KANIKO_TIMEOUT})" >&2
      ;;
    *)
      echo "Kaniko job ${job_name} watcher exited with status ${wait_status}" >&2
      ;;
  esac
  return 1
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

  if ! wait_for_kaniko_job "${job_name}"; then
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get "job/${job_name}" -o wide || true
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get pods -l "job-name=${job_name}" -o wide || true
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" describe "job/${job_name}" || true
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" logs "job/${job_name}" --all-containers=true --tail=200 || true
    exit 1
  fi
}

require_cmd kubectl
require_cmd sed
require_cmd sha256sum
require_cmd timeout
require_cmd awk
require_cmd tr
require_cmd cut

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_PORT="${CODEXK8S_INTERNAL_REGISTRY_PORT:-5000}"
CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE="${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE:-20Gi}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}}"
CODEXK8S_KANIKO_TIMEOUT="${CODEXK8S_KANIKO_TIMEOUT:-1800s}"
CODEXK8S_KANIKO_CACHE_ENABLED="${CODEXK8S_KANIKO_CACHE_ENABLED:-true}"
CODEXK8S_KANIKO_CACHE_REPO="${CODEXK8S_KANIKO_CACHE_REPO:-${CODEXK8S_INTERNAL_REGISTRY_HOST}/codex-k8s/kaniko-cache}"
CODEXK8S_KANIKO_CACHE_TTL="${CODEXK8S_KANIKO_CACHE_TTL:-168h}"
CODEXK8S_KANIKO_CACHE_COMPRESSED="${CODEXK8S_KANIKO_CACHE_COMPRESSED:-false}"
CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU="${CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU:-8}"
CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY="${CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY:-16Gi}"
CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU="${CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU:-16}"
CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY="${CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY:-32Gi}"
CODEXK8S_ENSURE_REGISTRY="${CODEXK8S_ENSURE_REGISTRY:-true}"
CODEXK8S_PREPARE_ONLY="${CODEXK8S_PREPARE_ONLY:-false}"
CODEXK8S_BUILD_COMPONENTS="${CODEXK8S_BUILD_COMPONENTS:-api-gateway,control-plane,worker,agent-runner,web-console}"
CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT:-600s}"
CODEXK8S_BUILD_REF="${CODEXK8S_BUILD_REF:-main}"
: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"

CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/api-gateway}"
CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/control-plane}"
CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/worker}"
CODEXK8S_AGENT_RUNNER_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_AGENT_RUNNER_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/agent-runner}"
CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/web-console}"

CODEXK8S_BUILD_SHA="$(normalize_sha_tag "$CODEXK8S_BUILD_REF")"

CODEXK8S_API_GATEWAY_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_API_GATEWAY_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_CONTROL_PLANE_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_CONTROL_PLANE_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_WORKER_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_WORKER_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_AGENT_RUNNER_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_AGENT_RUNNER_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_AGENT_RUNNER_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_AGENT_RUNNER_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

CODEXK8S_WEB_CONSOLE_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_WEB_CONSOLE_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"

if [ "$CODEXK8S_ENSURE_REGISTRY" = "true" ]; then
  render_registry_template "${ROOT_DIR}/deploy/base/registry/registry.yaml.tpl" | kubectl apply -f -
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status "deployment/${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --timeout="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT}"
fi

kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" create secret generic codex-k8s-git-token \
  --from-literal=token="${CODEXK8S_GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

if [ "${CODEXK8S_PREPARE_ONLY}" = "true" ]; then
  echo "Kaniko build prerequisites are prepared (registry + git token secret)."
  exit 0
fi

IFS=',' read -r -a build_components <<< "${CODEXK8S_BUILD_COMPONENTS}"
declare -a built_components

for raw_component in "${build_components[@]}"; do
  component="${raw_component//[[:space:]]/}"
  component="${component,,}"
  [ -n "$component" ] || continue
  build_component "$component"
  built_components+=("$component")
done

if [ "${#built_components[@]}" -eq 0 ]; then
  echo "CODEXK8S_BUILD_COMPONENTS does not include any known components." >&2
  exit 1
fi

echo "Internal images build completed:"
for component in "${built_components[@]}"; do
  case "$component" in
    api-gateway)
      echo "  api-gateway: ${CODEXK8S_API_GATEWAY_IMAGE_LATEST}"
      echo "  api-gateway: ${CODEXK8S_API_GATEWAY_IMAGE_SHA}"
      ;;
    control-plane)
      echo "  control-plane: ${CODEXK8S_CONTROL_PLANE_IMAGE_LATEST}"
      echo "  control-plane: ${CODEXK8S_CONTROL_PLANE_IMAGE_SHA}"
      ;;
    worker)
      echo "  worker: ${CODEXK8S_WORKER_IMAGE_LATEST}"
      echo "  worker: ${CODEXK8S_WORKER_IMAGE_SHA}"
      ;;
    agent-runner)
      echo "  agent-runner: ${CODEXK8S_AGENT_RUNNER_IMAGE_LATEST}"
      echo "  agent-runner: ${CODEXK8S_AGENT_RUNNER_IMAGE_SHA}"
      ;;
    web-console)
      echo "  web-console(dev target): ${CODEXK8S_WEB_CONSOLE_IMAGE_LATEST}"
      echo "  web-console(dev target): ${CODEXK8S_WEB_CONSOLE_IMAGE_SHA}"
      ;;
  esac
done
