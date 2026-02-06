#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

REPO_DIR="$(repo_dir)"
CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
: "${CODEXK8S_GITHUB_USERNAME:?CODEXK8S_GITHUB_USERNAME is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"

kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret docker-registry ghcr-pull-secret \
  --docker-server=ghcr.io \
  --docker-username="${CODEXK8S_GITHUB_USERNAME}" \
  --docker-password="${CODEXK8S_GITHUB_PAT}" \
  --docker-email="noreply@codex-k8s.dev" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-postgres \
  --from-literal=CODEXK8S_POSTGRES_DB="${CODEXK8S_POSTGRES_DB}" \
  --from-literal=CODEXK8S_POSTGRES_USER="${CODEXK8S_POSTGRES_USER}" \
  --from-literal=CODEXK8S_POSTGRES_PASSWORD="${CODEXK8S_POSTGRES_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-runtime \
  --from-literal=CODEXK8S_OPENAI_API_KEY="${CODEXK8S_OPENAI_API_KEY}" \
  --from-literal=CODEXK8S_CONTEXT7_API_KEY="${CODEXK8S_CONTEXT7_API_KEY:-}" \
  --from-literal=CODEXK8S_APP_SECRET_KEY="${CODEXK8S_APP_SECRET_KEY}" \
  --from-literal=CODEXK8S_TOKEN_ENCRYPTION_KEY="${CODEXK8S_TOKEN_ENCRYPTION_KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -

export CODEXK8S_STAGING_NAMESPACE CODEXK8S_IMAGE
bash "${REPO_DIR}/deploy/scripts/deploy_staging.sh"
