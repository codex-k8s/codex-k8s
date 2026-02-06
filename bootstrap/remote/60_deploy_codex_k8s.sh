#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

REPO_DIR="$(repo_dir)"
STAGING_NAMESPACE="${STAGING_NAMESPACE:-codex-k8s-ai-staging}"

kubectl -n "$STAGING_NAMESPACE" create secret generic codex-k8s-postgres \
  --from-literal=POSTGRES_DB="${POSTGRES_DB}" \
  --from-literal=POSTGRES_USER="${POSTGRES_USER}" \
  --from-literal=POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "$STAGING_NAMESPACE" create secret generic codex-k8s-runtime \
  --from-literal=OPENAI_API_KEY="${OPENAI_API_KEY}" \
  --from-literal=CONTEXT7_API_KEY="${CONTEXT7_API_KEY:-}" \
  --from-literal=APP_SECRET_KEY="${APP_SECRET_KEY}" \
  --from-literal=TOKEN_ENCRYPTION_KEY="${TOKEN_ENCRYPTION_KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -

export STAGING_NAMESPACE CODEXK8S_IMAGE
bash "${REPO_DIR}/deploy/scripts/deploy_staging.sh"
