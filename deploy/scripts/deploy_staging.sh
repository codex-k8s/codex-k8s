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
require_cmd envsubst

STAGING_NAMESPACE="${STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_IMAGE="${CODEXK8S_IMAGE:-ghcr.io/codex-k8s/codex-k8s:latest}"
STAGING_DOMAIN="${STAGING_DOMAIN:-}"
CODEXK8S_WAIT_ROLLOUT="${CODEXK8S_WAIT_ROLLOUT:-false}"

POSTGRES_DB="${POSTGRES_DB:-codex_k8s}"
POSTGRES_USER="${POSTGRES_USER:-codex_k8s}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"
APP_SECRET_KEY="${APP_SECRET_KEY:-}"
TOKEN_ENCRYPTION_KEY="${TOKEN_ENCRYPTION_KEY:-}"
OPENAI_API_KEY="${OPENAI_API_KEY:-}"
CONTEXT7_API_KEY="${CONTEXT7_API_KEY:-}"

if [ -z "$POSTGRES_PASSWORD" ] && ! kubectl -n "$STAGING_NAMESPACE" get secret codex-k8s-postgres >/dev/null 2>&1; then
  POSTGRES_PASSWORD="$(rand_hex 24)"
fi

if [ -z "$APP_SECRET_KEY" ] && ! kubectl -n "$STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  APP_SECRET_KEY="$(rand_hex 32)"
fi

if [ -z "$TOKEN_ENCRYPTION_KEY" ] && ! kubectl -n "$STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  TOKEN_ENCRYPTION_KEY="$(rand_hex 32)"
fi

export STAGING_NAMESPACE CODEXK8S_IMAGE STAGING_DOMAIN

envsubst < "${ROOT_DIR}/deploy/base/namespace/namespace.yaml.tpl" | kubectl apply -f -

if ! kubectl -n "$STAGING_NAMESPACE" get secret codex-k8s-postgres >/dev/null 2>&1; then
  kubectl -n "$STAGING_NAMESPACE" create secret generic codex-k8s-postgres \
    --from-literal=POSTGRES_DB="$POSTGRES_DB" \
    --from-literal=POSTGRES_USER="$POSTGRES_USER" \
    --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
fi

if ! kubectl -n "$STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  kubectl -n "$STAGING_NAMESPACE" create secret generic codex-k8s-runtime \
    --from-literal=OPENAI_API_KEY="$OPENAI_API_KEY" \
    --from-literal=CONTEXT7_API_KEY="$CONTEXT7_API_KEY" \
    --from-literal=APP_SECRET_KEY="$APP_SECRET_KEY" \
    --from-literal=TOKEN_ENCRYPTION_KEY="$TOKEN_ENCRYPTION_KEY"
fi

envsubst < "${ROOT_DIR}/deploy/base/postgres/postgres.yaml.tpl" | kubectl apply -f -
envsubst < "${ROOT_DIR}/deploy/base/codex-k8s/app.yaml.tpl" | kubectl apply -f -

if [ -n "$STAGING_DOMAIN" ]; then
  envsubst < "${ROOT_DIR}/deploy/base/codex-k8s/ingress.yaml.tpl" | kubectl apply -f -
fi

if [ "$CODEXK8S_WAIT_ROLLOUT" = "true" ]; then
  kubectl -n "$STAGING_NAMESPACE" rollout status statefulset/postgres --timeout=900s
  kubectl -n "$STAGING_NAMESPACE" rollout status deployment/codex-k8s --timeout=900s
fi

echo "Staging apply completed for namespace ${STAGING_NAMESPACE}"
