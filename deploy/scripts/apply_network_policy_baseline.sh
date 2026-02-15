#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing command: $1" >&2
    exit 1
  }
}

render_platform_policy() {
  local tpl="$1"
  sed \
    -e "s|\${CODEXK8S_STAGING_NAMESPACE}|${CODEXK8S_STAGING_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_K8S_API_CIDR}|${CODEXK8S_K8S_API_CIDR}|g" \
    -e "s|\${CODEXK8S_K8S_API_PORT}|${CODEXK8S_K8S_API_PORT}|g" \
    "$tpl"
}

render_project_agent_policy() {
  local tpl="$1"
  sed \
    -e "s|\${CODEXK8S_TARGET_NAMESPACE}|${CODEXK8S_TARGET_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_PLATFORM_MCP_PORT}|${CODEXK8S_PLATFORM_MCP_PORT}|g" \
    "$tpl"
}

label_namespace_if_exists() {
  local ns="$1"
  local key="$2"
  local value="$3"
  if kubectl get namespace "$ns" >/dev/null 2>&1; then
    kubectl label namespace "$ns" "${key}=${value}" --overwrite >/dev/null
  fi
}

require_cmd kubectl
require_cmd sed

CODEXK8S_NETWORK_POLICY_BASELINE="${CODEXK8S_NETWORK_POLICY_BASELINE:-true}"
CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_APPLY_PROJECT_AGENT_POLICY="${CODEXK8S_APPLY_PROJECT_AGENT_POLICY:-false}"
CODEXK8S_TARGET_NAMESPACE="${CODEXK8S_TARGET_NAMESPACE:-}"
CODEXK8S_PLATFORM_MCP_PORT="${CODEXK8S_PLATFORM_MCP_PORT:-8081}"
CODEXK8S_K8S_API_CIDR="${CODEXK8S_K8S_API_CIDR:-0.0.0.0/0}"
CODEXK8S_K8S_API_PORT="${CODEXK8S_K8S_API_PORT:-6443}"

if [ "$CODEXK8S_NETWORK_POLICY_BASELINE" != "true" ]; then
  echo "Network policy baseline disabled by CODEXK8S_NETWORK_POLICY_BASELINE=${CODEXK8S_NETWORK_POLICY_BASELINE}"
  exit 0
fi

echo "Label namespaces for network zoning"
label_namespace_if_exists "$CODEXK8S_STAGING_NAMESPACE" "codexk8s.io/network-zone" "platform"
label_namespace_if_exists "ingress-nginx" "codexk8s.io/network-zone" "system"
label_namespace_if_exists "cert-manager" "codexk8s.io/network-zone" "system"
label_namespace_if_exists "kube-system" "codexk8s.io/network-zone" "system"

echo "Apply platform network policy baseline in namespace ${CODEXK8S_STAGING_NAMESPACE}"
render_platform_policy "${ROOT_DIR}/deploy/base/network-policies/platform-baseline.yaml.tpl" | kubectl apply -f -

if [ "$CODEXK8S_APPLY_PROJECT_AGENT_POLICY" = "true" ]; then
  [ -n "$CODEXK8S_TARGET_NAMESPACE" ] || {
    echo "Missing CODEXK8S_TARGET_NAMESPACE for project/agent policy apply" >&2
    exit 1
  }
  kubectl get namespace "$CODEXK8S_TARGET_NAMESPACE" >/dev/null 2>&1 || {
    echo "Namespace not found: ${CODEXK8S_TARGET_NAMESPACE}" >&2
    exit 1
  }
  label_namespace_if_exists "$CODEXK8S_TARGET_NAMESPACE" "codexk8s.io/network-zone" "project"
  echo "Apply project/agent network policy baseline in namespace ${CODEXK8S_TARGET_NAMESPACE}"
  render_project_agent_policy "${ROOT_DIR}/deploy/base/network-policies/project-agent-baseline.yaml.tpl" | kubectl apply -f -
fi

echo "Network policy baseline apply completed"
