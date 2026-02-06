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

normalize_sha_tag() {
  local ref="$1"
  if [[ "$ref" =~ ^[0-9a-fA-F]{12,40}$ ]]; then
    printf '%s' "${ref:0:12}" | tr '[:upper:]' '[:lower:]'
    return 0
  fi
  printf '%s' "$ref" | sha256sum | awk '{print $1}' | cut -c1-12
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
CODEXK8S_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/codex-k8s}"
CODEXK8S_BUILD_REF="${CODEXK8S_BUILD_REF:-main}"
CODEXK8S_KANIKO_TIMEOUT="${CODEXK8S_KANIKO_TIMEOUT:-1800s}"
CODEXK8S_ENSURE_REGISTRY="${CODEXK8S_ENSURE_REGISTRY:-true}"
CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT:-600s}"
: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"

CODEXK8S_BUILD_SHA="$(normalize_sha_tag "$CODEXK8S_BUILD_REF")"
CODEXK8S_IMAGE_LATEST="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_INTERNAL_IMAGE_REPOSITORY}:latest"
CODEXK8S_IMAGE_SHA="${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_INTERNAL_IMAGE_REPOSITORY}:sha-${CODEXK8S_BUILD_SHA}"
CODEXK8S_KANIKO_JOB_NAME="codex-k8s-kaniko-${CODEXK8S_BUILD_SHA}"

if [ "$CODEXK8S_ENSURE_REGISTRY" = "true" ]; then
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete statefulset "${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --ignore-not-found=true >/dev/null 2>&1 || true
  render_registry_template "${ROOT_DIR}/deploy/base/registry/registry.yaml.tpl" | kubectl apply -f -
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete service "${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --ignore-not-found=true >/dev/null 2>&1 || true
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status "deployment/${CODEXK8S_INTERNAL_REGISTRY_SERVICE}" --timeout="${CODEXK8S_REGISTRY_ROLLOUT_TIMEOUT}"
fi

kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" create secret generic codex-k8s-git-token \
  --from-literal=token="${CODEXK8S_GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" delete job "${CODEXK8S_KANIKO_JOB_NAME}" --ignore-not-found=true >/dev/null 2>&1 || true
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=delete "job/${CODEXK8S_KANIKO_JOB_NAME}" --timeout=120s >/dev/null 2>&1 || true

cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${CODEXK8S_KANIKO_JOB_NAME}
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-kaniko
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-kaniko
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Never
      volumes:
        - name: workspace
          emptyDir: {}
      initContainers:
        - name: clone
          image: alpine/git:2.47.2
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-git-token
                  key: token
            - name: CODEXK8S_GITHUB_REPO
              value: "${CODEXK8S_GITHUB_REPO}"
            - name: CODEXK8S_BUILD_REF
              value: "${CODEXK8S_BUILD_REF}"
          command:
            - sh
            - -ec
            - |
              git clone "https://x-access-token:\${GIT_TOKEN}@github.com/\${CODEXK8S_GITHUB_REPO}.git" /workspace
              cd /workspace
              git checkout --detach "\${CODEXK8S_BUILD_REF}"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
      containers:
        - name: kaniko
          image: gcr.io/kaniko-project/executor:v1.23.2-debug
          imagePullPolicy: IfNotPresent
          args:
            - --context=dir:///workspace
            - --dockerfile=/workspace/Dockerfile
            - --destination=${CODEXK8S_IMAGE_LATEST}
            - --destination=${CODEXK8S_IMAGE_SHA}
            - --insecure
            - --insecure-registry=${CODEXK8S_INTERNAL_REGISTRY_HOST}
            - --skip-tls-verify-registry=${CODEXK8S_INTERNAL_REGISTRY_HOST}
          volumeMounts:
            - name: workspace
              mountPath: /workspace
EOF

if ! kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=condition=complete "job/${CODEXK8S_KANIKO_JOB_NAME}" --timeout="${CODEXK8S_KANIKO_TIMEOUT}"; then
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get pods -l "job-name=${CODEXK8S_KANIKO_JOB_NAME}" -o wide || true
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" logs "job/${CODEXK8S_KANIKO_JOB_NAME}" --all-containers=true --tail=200 || true
  exit 1
fi

echo "Internal image build completed:"
echo "  ${CODEXK8S_IMAGE_LATEST}"
echo "  ${CODEXK8S_IMAGE_SHA}"
