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
  if [ -f "${file}" ]; then
    kubectl apply -f "${file}"
    return 0
  fi
  kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Namespace
metadata:
  name: actions-runner-system
YAML
}

apply_runner_namespace() {
  local tpl="${REPO_DIR}/deploy/runner/runner-namespace.yaml.tpl"
  if [ -f "${tpl}" ]; then
    export CODEXK8S_RUNNER_NAMESPACE
    envsubst < "${tpl}" | kubectl apply -f -
    return 0
  fi
  kubectl apply -f - <<YAML
apiVersion: v1
kind: Namespace
metadata:
  name: ${CODEXK8S_RUNNER_NAMESPACE}
  labels:
    app.kubernetes.io/part-of: codex-k8s
    app.kubernetes.io/component: github-runner
YAML
}

apply_staging_deployer_rbac() {
  local tpl="${REPO_DIR}/deploy/runner/staging-deployer-rbac.yaml.tpl"
  if [ -f "${tpl}" ]; then
    export CODEXK8S_STAGING_NAMESPACE CODEXK8S_RUNNER_NAMESPACE CODEXK8S_RUNNER_SCALE_SET_NAME CODEXK8S_WORKER_RUN_ROLE_NAME
    envsubst < "${tpl}" | kubectl apply -f -
    return 0
  fi
  kubectl apply -f - <<YAML
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: codex-k8s-staging-deployer
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
rules:
  - apiGroups: [""]
    resources: ["secrets", "configmaps", "services", "pods", "pods/log", "persistentvolumeclaims", "serviceaccounts"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses", "networkpolicies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: codex-k8s-staging-deployer
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
subjects:
  - kind: ServiceAccount
    name: ${CODEXK8S_RUNNER_SCALE_SET_NAME}-gha-rs-no-permission
    namespace: ${CODEXK8S_RUNNER_NAMESPACE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: codex-k8s-staging-deployer
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: codex-k8s-staging-deployer-cluster
rules:
  - apiGroups: [""]
    resources: ["namespaces", "serviceaccounts", "resourcequotas", "limitranges", "pods", "pods/log", "events"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["configmaps", "endpoints", "services"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["daemonsets", "deployments", "replicasets", "statefulsets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["clusterroles", "clusterrolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles"]
    verbs: ["escalate", "bind"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: codex-k8s-staging-deployer-cluster
subjects:
  - kind: ServiceAccount
    name: ${CODEXK8S_RUNNER_SCALE_SET_NAME}-gha-rs-no-permission
    namespace: ${CODEXK8S_RUNNER_NAMESPACE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: codex-k8s-staging-deployer-cluster
YAML
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
if [ -f "${values_tpl}" ]; then
  export CODEXK8S_GITHUB_REPO CODEXK8S_RUNNER_MIN CODEXK8S_RUNNER_MAX CODEXK8S_RUNNER_IMAGE CODEXK8S_RUNNER_SCALE_SET_NAME
  envsubst < "${values_tpl}" > "${VALUES_FILE}"
else
  cat > "${VALUES_FILE}" <<YAML
githubConfigUrl: https://github.com/${CODEXK8S_GITHUB_REPO}
githubConfigSecret: gha-runner-scale-set-secret
runnerScaleSetName: ${CODEXK8S_RUNNER_SCALE_SET_NAME}
minRunners: ${CODEXK8S_RUNNER_MIN}
maxRunners: ${CODEXK8S_RUNNER_MAX}
template:
  spec:
    containers:
      - name: runner
        image: ${CODEXK8S_RUNNER_IMAGE}
        command: ["/home/runner/run.sh"]
YAML
fi
helm upgrade --install "${CODEXK8S_RUNNER_SCALE_SET_NAME}" oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  --namespace "${CODEXK8S_RUNNER_NAMESPACE}" \
  --create-namespace \
  -f "${VALUES_FILE}" \
  --wait \
  --timeout "${CODEXK8S_HELM_TIMEOUT}"
rm -f "${VALUES_FILE}"

log "Grant staging deploy RBAC to runner service account"
apply_staging_deployer_rbac
