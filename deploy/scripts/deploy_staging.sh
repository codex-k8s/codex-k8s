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

escape_sed_replacement() {
  printf '%s' "$1" | sed -e 's/[&|]/\\&/g'
}

require_cmd kubectl

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_PORT="${CODEXK8S_INTERNAL_REGISTRY_PORT:-5000}"
CODEXK8S_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/codex-k8s}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}}"
CODEXK8S_IMAGE="${CODEXK8S_IMAGE:-${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_INTERNAL_IMAGE_REPOSITORY}:latest}"
CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY="${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY:-codex-k8s/web-console}"
CODEXK8S_WEB_CONSOLE_IMAGE="${CODEXK8S_WEB_CONSOLE_IMAGE:-${CODEXK8S_INTERNAL_REGISTRY_HOST}/${CODEXK8S_WEB_CONSOLE_INTERNAL_IMAGE_REPOSITORY}:latest}"
CODEXK8S_STAGING_DOMAIN="${CODEXK8S_STAGING_DOMAIN:-}"
CODEXK8S_WAIT_ROLLOUT="${CODEXK8S_WAIT_ROLLOUT:-true}"
CODEXK8S_ROLLOUT_TIMEOUT="${CODEXK8S_ROLLOUT_TIMEOUT:-1800s}"
CODEXK8S_APPLY_NAMESPACE="${CODEXK8S_APPLY_NAMESPACE:-false}"
CODEXK8S_WORKER_REPLICAS="${CODEXK8S_WORKER_REPLICAS:-1}"
CODEXK8S_WORKER_POLL_INTERVAL="${CODEXK8S_WORKER_POLL_INTERVAL:-5s}"
CODEXK8S_WORKER_CLAIM_LIMIT="${CODEXK8S_WORKER_CLAIM_LIMIT:-2}"
CODEXK8S_WORKER_RUNNING_CHECK_LIMIT="${CODEXK8S_WORKER_RUNNING_CHECK_LIMIT:-200}"
CODEXK8S_WORKER_SLOTS_PER_PROJECT="${CODEXK8S_WORKER_SLOTS_PER_PROJECT:-2}"
CODEXK8S_WORKER_SLOT_LEASE_TTL="${CODEXK8S_WORKER_SLOT_LEASE_TTL:-10m}"
CODEXK8S_WORKER_K8S_NAMESPACE="${CODEXK8S_WORKER_K8S_NAMESPACE:-$CODEXK8S_STAGING_NAMESPACE}"
CODEXK8S_WORKER_JOB_IMAGE="${CODEXK8S_WORKER_JOB_IMAGE:-$CODEXK8S_IMAGE}"
CODEXK8S_WORKER_JOB_COMMAND="${CODEXK8S_WORKER_JOB_COMMAND:-echo codex-k8s run; sleep 2}"
CODEXK8S_WORKER_JOB_TTL_SECONDS="${CODEXK8S_WORKER_JOB_TTL_SECONDS:-600}"
CODEXK8S_WORKER_JOB_BACKOFF_LIMIT="${CODEXK8S_WORKER_JOB_BACKOFF_LIMIT:-0}"
CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS="${CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS:-900}"

CODEXK8S_POSTGRES_DB="${CODEXK8S_POSTGRES_DB:-codex_k8s}"
CODEXK8S_POSTGRES_USER="${CODEXK8S_POSTGRES_USER:-codex_k8s}"
CODEXK8S_POSTGRES_PASSWORD="${CODEXK8S_POSTGRES_PASSWORD:-}"
CODEXK8S_APP_SECRET_KEY="${CODEXK8S_APP_SECRET_KEY:-}"
CODEXK8S_TOKEN_ENCRYPTION_KEY="${CODEXK8S_TOKEN_ENCRYPTION_KEY:-}"
CODEXK8S_GITHUB_WEBHOOK_SECRET="${CODEXK8S_GITHUB_WEBHOOK_SECRET:-}"
CODEXK8S_GITHUB_WEBHOOK_URL="${CODEXK8S_GITHUB_WEBHOOK_URL:-}"
CODEXK8S_GITHUB_WEBHOOK_EVENTS="${CODEXK8S_GITHUB_WEBHOOK_EVENTS:-}"
CODEXK8S_PUBLIC_BASE_URL="${CODEXK8S_PUBLIC_BASE_URL:-}"
CODEXK8S_BOOTSTRAP_OWNER_EMAIL="${CODEXK8S_BOOTSTRAP_OWNER_EMAIL:-}"
CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS="${CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS:-}"
CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS="${CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS:-}"
CODEXK8S_LEARNING_MODE_DEFAULT="${CODEXK8S_LEARNING_MODE_DEFAULT:-}"
CODEXK8S_GITHUB_OAUTH_CLIENT_ID="${CODEXK8S_GITHUB_OAUTH_CLIENT_ID:-}"
CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET="${CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET:-}"
CODEXK8S_JWT_SIGNING_KEY="${CODEXK8S_JWT_SIGNING_KEY:-}"
CODEXK8S_JWT_TTL="${CODEXK8S_JWT_TTL:-15m}"
CODEXK8S_OPENAI_API_KEY="${CODEXK8S_OPENAI_API_KEY:-}"
CODEXK8S_CONTEXT7_API_KEY="${CODEXK8S_CONTEXT7_API_KEY:-}"
CODEXK8S_VITE_DEV_UPSTREAM="${CODEXK8S_VITE_DEV_UPSTREAM:-http://codex-k8s-web-console:5173}"

[ -n "$CODEXK8S_STAGING_DOMAIN" ] || {
  echo "Missing required CODEXK8S_STAGING_DOMAIN" >&2
  exit 1
}
[ -n "$CODEXK8S_GITHUB_WEBHOOK_SECRET" ] || {
  echo "Missing required CODEXK8S_GITHUB_WEBHOOK_SECRET" >&2
  exit 1
}
[ -n "$CODEXK8S_PUBLIC_BASE_URL" ] || {
  echo "Missing required CODEXK8S_PUBLIC_BASE_URL" >&2
  exit 1
}
[ -n "$CODEXK8S_BOOTSTRAP_OWNER_EMAIL" ] || {
  echo "Missing required CODEXK8S_BOOTSTRAP_OWNER_EMAIL" >&2
  exit 1
}
[ -n "$CODEXK8S_GITHUB_OAUTH_CLIENT_ID" ] || {
  echo "Missing required CODEXK8S_GITHUB_OAUTH_CLIENT_ID" >&2
  exit 1
}
[ -n "$CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET" ] || {
  echo "Missing required CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET" >&2
  exit 1
}
[ -n "$CODEXK8S_JWT_SIGNING_KEY" ] || {
  echo "Missing required CODEXK8S_JWT_SIGNING_KEY" >&2
  exit 1
}

echo "Apply network policy baseline (if enabled)"
bash "${ROOT_DIR}/deploy/scripts/apply_network_policy_baseline.sh"

if [ -z "$CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS" ] && kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  # Preserve existing value on deploy to avoid wiping allowlist when GitHub vars are not synced yet.
  CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS="$(
    kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime \
      -o jsonpath='{.data.CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS}' 2>/dev/null | base64 -d || true
  )"
fi
if [ -z "$CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS" ] && kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS="$(
    kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime \
      -o jsonpath='{.data.CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS}' 2>/dev/null | base64 -d || true
  )"
fi

if [ -z "$CODEXK8S_POSTGRES_PASSWORD" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-postgres >/dev/null 2>&1; then
  CODEXK8S_POSTGRES_PASSWORD="$(rand_hex 24)"
fi

if [ -z "$CODEXK8S_APP_SECRET_KEY" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  CODEXK8S_APP_SECRET_KEY="$(rand_hex 32)"
fi

if [ -z "$CODEXK8S_TOKEN_ENCRYPTION_KEY" ] && ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-runtime >/dev/null 2>&1; then
  CODEXK8S_TOKEN_ENCRYPTION_KEY="$(rand_hex 32)"
fi

export CODEXK8S_STAGING_NAMESPACE CODEXK8S_IMAGE CODEXK8S_WEB_CONSOLE_IMAGE CODEXK8S_STAGING_DOMAIN

render_template() {
  local tpl="$1"
  local image_escaped
  local domain_escaped
  local worker_job_image_escaped
  local worker_job_command_escaped
  local web_console_image_escaped
  local vite_dev_upstream_escaped
  image_escaped="$(escape_sed_replacement "$CODEXK8S_IMAGE")"
  domain_escaped="$(escape_sed_replacement "$CODEXK8S_STAGING_DOMAIN")"
  worker_job_image_escaped="$(escape_sed_replacement "$CODEXK8S_WORKER_JOB_IMAGE")"
  worker_job_command_escaped="$(escape_sed_replacement "$CODEXK8S_WORKER_JOB_COMMAND")"
  web_console_image_escaped="$(escape_sed_replacement "$CODEXK8S_WEB_CONSOLE_IMAGE")"
  vite_dev_upstream_escaped="$(escape_sed_replacement "$CODEXK8S_VITE_DEV_UPSTREAM")"
  sed \
    -e "s|\${CODEXK8S_STAGING_NAMESPACE}|${CODEXK8S_STAGING_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_IMAGE}|${image_escaped}|g" \
    -e "s|\${CODEXK8S_WEB_CONSOLE_IMAGE}|${web_console_image_escaped}|g" \
    -e "s|\${CODEXK8S_STAGING_DOMAIN}|${domain_escaped}|g" \
    -e "s|\${CODEXK8S_WORKER_REPLICAS}|${CODEXK8S_WORKER_REPLICAS}|g" \
    -e "s|\${CODEXK8S_WORKER_POLL_INTERVAL}|${CODEXK8S_WORKER_POLL_INTERVAL}|g" \
    -e "s|\${CODEXK8S_WORKER_CLAIM_LIMIT}|${CODEXK8S_WORKER_CLAIM_LIMIT}|g" \
    -e "s|\${CODEXK8S_WORKER_RUNNING_CHECK_LIMIT}|${CODEXK8S_WORKER_RUNNING_CHECK_LIMIT}|g" \
    -e "s|\${CODEXK8S_WORKER_SLOTS_PER_PROJECT}|${CODEXK8S_WORKER_SLOTS_PER_PROJECT}|g" \
    -e "s|\${CODEXK8S_WORKER_SLOT_LEASE_TTL}|${CODEXK8S_WORKER_SLOT_LEASE_TTL}|g" \
    -e "s|\${CODEXK8S_WORKER_K8S_NAMESPACE}|${CODEXK8S_WORKER_K8S_NAMESPACE}|g" \
    -e "s|\${CODEXK8S_WORKER_JOB_IMAGE}|${worker_job_image_escaped}|g" \
    -e "s|\${CODEXK8S_WORKER_JOB_COMMAND}|${worker_job_command_escaped}|g" \
    -e "s|\${CODEXK8S_WORKER_JOB_TTL_SECONDS}|${CODEXK8S_WORKER_JOB_TTL_SECONDS}|g" \
    -e "s|\${CODEXK8S_WORKER_JOB_BACKOFF_LIMIT}|${CODEXK8S_WORKER_JOB_BACKOFF_LIMIT}|g" \
    -e "s|\${CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS}|${CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS}|g" \
    -e "s|\${CODEXK8S_VITE_DEV_UPSTREAM}|${vite_dev_upstream_escaped}|g" \
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

kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-runtime \
  --from-literal=CODEXK8S_OPENAI_API_KEY="$CODEXK8S_OPENAI_API_KEY" \
  --from-literal=CODEXK8S_CONTEXT7_API_KEY="$CODEXK8S_CONTEXT7_API_KEY" \
  --from-literal=CODEXK8S_APP_SECRET_KEY="$CODEXK8S_APP_SECRET_KEY" \
  --from-literal=CODEXK8S_TOKEN_ENCRYPTION_KEY="$CODEXK8S_TOKEN_ENCRYPTION_KEY" \
  --from-literal=CODEXK8S_LEARNING_MODE_DEFAULT="$CODEXK8S_LEARNING_MODE_DEFAULT" \
  --from-literal=CODEXK8S_GITHUB_WEBHOOK_SECRET="$CODEXK8S_GITHUB_WEBHOOK_SECRET" \
  --from-literal=CODEXK8S_GITHUB_WEBHOOK_URL="$CODEXK8S_GITHUB_WEBHOOK_URL" \
  --from-literal=CODEXK8S_GITHUB_WEBHOOK_EVENTS="$CODEXK8S_GITHUB_WEBHOOK_EVENTS" \
  --from-literal=CODEXK8S_PUBLIC_BASE_URL="$CODEXK8S_PUBLIC_BASE_URL" \
  --from-literal=CODEXK8S_BOOTSTRAP_OWNER_EMAIL="$CODEXK8S_BOOTSTRAP_OWNER_EMAIL" \
  --from-literal=CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS="$CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS" \
  --from-literal=CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS="$CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS" \
  --from-literal=CODEXK8S_GITHUB_OAUTH_CLIENT_ID="$CODEXK8S_GITHUB_OAUTH_CLIENT_ID" \
  --from-literal=CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET="$CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET" \
  --from-literal=CODEXK8S_JWT_SIGNING_KEY="$CODEXK8S_JWT_SIGNING_KEY" \
  --from-literal=CODEXK8S_JWT_TTL="$CODEXK8S_JWT_TTL" \
  --from-literal=CODEXK8S_VITE_DEV_UPSTREAM="$CODEXK8S_VITE_DEV_UPSTREAM" \
  --dry-run=client -o yaml | kubectl apply -f -

if ! kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-oauth2-proxy >/dev/null 2>&1; then
  cookie_secret="$(openssl rand -hex 16)"
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-oauth2-proxy \
    --from-literal=OAUTH2_PROXY_CLIENT_ID="$CODEXK8S_GITHUB_OAUTH_CLIENT_ID" \
    --from-literal=OAUTH2_PROXY_CLIENT_SECRET="$CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET" \
    --from-literal=OAUTH2_PROXY_COOKIE_SECRET="$cookie_secret"
else
  # oauth2-proxy expects the decoded cookie secret to be 16/24/32 bytes.
  # Early revisions generated a longer string; self-heal on staging.
  existing_cookie_secret="$(
    kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get secret codex-k8s-oauth2-proxy \
      -o jsonpath='{.data.OAUTH2_PROXY_COOKIE_SECRET}' 2>/dev/null | base64 -d || true
  )"
  case "${#existing_cookie_secret}" in
    16|24|32) ;;
    *)
      echo "Fix oauth2-proxy cookie secret length (${#existing_cookie_secret}); rotating secret"
      cookie_secret="$(openssl rand -hex 16)"
      kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create secret generic codex-k8s-oauth2-proxy \
        --from-literal=OAUTH2_PROXY_CLIENT_ID="$CODEXK8S_GITHUB_OAUTH_CLIENT_ID" \
        --from-literal=OAUTH2_PROXY_CLIENT_SECRET="$CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET" \
        --from-literal=OAUTH2_PROXY_COOKIE_SECRET="$cookie_secret" \
        --dry-run=client -o yaml | kubectl apply -f -
      ;;
  esac
fi

kubectl -n "$CODEXK8S_STAGING_NAMESPACE" create configmap codex-k8s-migrations \
  --from-file="${ROOT_DIR}/services/internal/control-plane/cmd/cli/migrations" \
  --dry-run=client -o yaml | kubectl apply -f -

render_template "${ROOT_DIR}/deploy/base/postgres/postgres.yaml.tpl" | kubectl apply -f -
render_template "${ROOT_DIR}/deploy/base/web-console/web-console-dev.yaml.tpl" | kubectl apply -f -
render_template "${ROOT_DIR}/deploy/base/oauth2-proxy/oauth2-proxy.yaml.tpl" | kubectl apply -f -

# We run DB migrations via goose as a dedicated Job to avoid:
# - parsing SQL in shell;
# - running migrations concurrently in multiple initContainers;
# - worker starting before schema is ready.
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status statefulset/postgres --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" delete job/codex-k8s-migrate >/dev/null 2>&1 || true
render_template "${ROOT_DIR}/deploy/base/codex-k8s/migrate-job.yaml.tpl" | kubectl apply -f -
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" wait --for=condition=complete job/codex-k8s-migrate --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"

# Deployment selector is immutable; early staging revisions used an overly broad selector
# (`app.kubernetes.io/name=codex-k8s`) that overlapped with worker pods.
old_selector_component="$(
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get deployment codex-k8s \
    -o jsonpath='{.spec.selector.matchLabels.app\.kubernetes\.io/component}' 2>/dev/null || true
)"
if [ -n "$old_selector_component" ] && [ "$old_selector_component" != "api-gateway" ]; then
  echo "Refusing to apply: deployment/codex-k8s has unexpected selector component=${old_selector_component}" >&2
  exit 1
fi
if [ -z "$old_selector_component" ] && kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get deployment codex-k8s >/dev/null 2>&1; then
  echo "Recreating deployment/codex-k8s to update immutable selector"
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" delete deployment codex-k8s
fi

render_template "${ROOT_DIR}/deploy/base/codex-k8s/app.yaml.tpl" | kubectl apply -f -
render_template "${ROOT_DIR}/deploy/base/codex-k8s/ingress.yaml.tpl" | kubectl apply -f -

# When images are referenced via the `:latest` tag, `kubectl apply` won't trigger a rollout by itself.
# Force a restart so that staging always converges to the newest in-cluster registry images.
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout restart deployment/codex-k8s >/dev/null 2>&1 || true
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout restart deployment/codex-k8s-control-plane >/dev/null 2>&1 || true
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout restart deployment/codex-k8s-worker >/dev/null 2>&1 || true
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout restart deployment/codex-k8s-web-console >/dev/null 2>&1 || true
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout restart deployment/oauth2-proxy >/dev/null 2>&1 || true

if [ "$CODEXK8S_WAIT_ROLLOUT" = "true" ]; then
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status deployment/codex-k8s --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status deployment/codex-k8s-worker --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}"
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status deployment/codex-k8s-web-console --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}" || true
  kubectl -n "$CODEXK8S_STAGING_NAMESPACE" rollout status deployment/oauth2-proxy --timeout="${CODEXK8S_ROLLOUT_TIMEOUT}" || true
fi

echo "Staging apply completed for namespace ${CODEXK8S_STAGING_NAMESPACE}"
