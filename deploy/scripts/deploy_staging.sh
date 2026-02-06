#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing command: $1" >&2
    exit 1
  }
}

rand_hex() {
  openssl rand -hex "$1" 2>/dev/null || head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'
}

require_cmd kubectl

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_PORT="${CODEXK8S_INTERNAL_REGISTRY_PORT:-5000}"
CODEXK8S_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/codex-k8s}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}}"
CODEXK8S_IMAGE="${CODEXK8S_IMAGE:-${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_INTERNAL_IMAGE_REPOSITORY}:latest}"
CODEXK8S_STAGING_DOMAIN="${CODEXK8S_STAGING_DOMAIN:-}"
CODEXK8S_WAIT_ROLLOUT="${CODEXK8S_WAIT_ROLLOUT:-true}"
CODEXK8S_ROLLOUT_TIMEOUT="${CODEXK8S_ROLLOUT_TIMEOUT:-1800s}"
CODEXK8S_APPLY_NAMESPACE="${CODEXK8S_APPLY_NAMESPACE:-false}"

CODEXK8S_POSTGRES_DB="${CODEXK8S_POSTGRES_DB:-codex_k8s}"
CODEXK8S_POSTGRES_USER="${CODEXK8S_POSTGRES_USER:-codex_k8s}"
CODEXK8S_POSTGRES_PASSWORD="${CODEXK8S_POSTGRES_PASSWORD:-}"
CODEXK8S_APP_SECRET_KEY="${CODEXK8S_APP_SECRET_KEY:-}"
CODEXK8S_TOKEN_ENCRYPTION_KEY="${CODEXK8S_TOKEN_ENCRYPTION_KEY:-}"
CODEXK8S_OPENAI_API_KEY="${CODEXK8S_OPENAI_API_KEY:-}"
CODEXK8S_CONTEXT7_API_KEY="${CODEXK8S_CONTEXT7_API_KEY:-}"

[ -n "$CODEXK8S_STAGING_DOMAIN" ] || {
  echo "Missing required CODEXK8S_STAGING_DOMAIN" >&2
  exit 1
}

if [ -z "$CODEXK8S_POSTGRES_PASSWORD" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-postgres >/dev/null 2>&1; then
  CODEXK8S_POSTGRES_PASSWORD="$(rand_hex 24)"
fi

if [ -z "$CODEXK8S_APP_SECRET_KEY" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  CODEXK8S_APP_SECRET_KEY="$(rand_hex 32)"
fi

if [ -z "$CODEXK8S_TOKEN_ENCRYPTION_KEY" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  CODEXK8S_TOKEN_ENCRYPTION_KEY="$(rand_hex 32)"
fi

export CODEXK8S_STAGING_NAMESPACE CODEXK8S_IMAGE CODEXK8S_STAGING_DOMAIN

render_template() {
  local tpl="$1"
  sed \
    -e "s|\${CODEXK8S_STAGING_NAMESPACE}|${CODEXK8S_STAGING_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_IMAGE}|${CODEXK8S_IMAGE}|g" \
    -e "s|\${CODEXK8S_STAGING_DOMAIN}|${CODEXK8S_STAGING_DOMAIN}|g" \
    "$tpl"
}

if [ "$CODEXK8S_APPLY_NAMESPACE" = "true" ]; then
  render_template "${ROOT_DIR}/deploy/base/namespace/namespace.yaml.tpl" | kubectl apply -f -
fi

if ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-postgres >/dev/null 2>&1; then
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-postgres \
    --from-literal=CODEXK8S_POSTGRES_DB="$CODEXK8S_POSTGRES_DB" \
    --from-literal=CODEXK8S_POSTGRES_USER="$CODEXK8S_POSTGRES_USER" \
    --from-literal=CODEXK8S_POSTGRES_PASSWORD="$CODEXK8S_POSTGRES_PASSWORD"
fi

if ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-runtime \
    --from-literal=CODEXK8S_OPENAI_API_KEY="$CODEXK8S_OPENAI_API_KEY" \
    --from-literal=CODEXK8S_CONTEXT7_API_KEY="$CODEXK8S_CONTEXT7_API_KEY" \
    --from-literal=CODEXK8S_APP_SECRET_KEY="$CODEXK8S_APP_SECRET_KEY" \
    --from-literal=CODEXK8S_TOKEN_ENCRYPTION_KEY="$CODEXK8S_TOKEN_ENCRYPTION_KEY"
fi

render_template "${ROOT_DIR}/deploy/base/postgres/postgres.yaml.tpl" | kubectl apply -f -
render_template "${ROOT_DIR}/deploy/base/codex-k8s/app.yaml.tpl" | kubectl apply -f -
render_template "${ROOT_DIR}/deploy/base/codex-k8s/ingress.yaml.tpl" | kubectl apply -f -

if [ "$CODEXK8S_WAIT_ROLLOUT" = "true" ]; then
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status statefulset/postgres --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status deployment/codex-k8s --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"
fi

echo "Staging apply completed for namespace ${CODEXK8S_STAGING_NAMESPACE}"
