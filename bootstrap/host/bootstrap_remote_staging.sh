#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_FILE="${ROOT_DIR}/host/config.env"

log() { echo "[$(date -Is)] $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

escape_squote() {
  printf "%s" "$1" | sed "s/'/'\\''/g"
}

rand_hex() {
  openssl rand -hex "$1" 2>/dev/null || head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

prompt_if_empty() {
  local var_name="$1"
  local prompt_text="$2"
  local secret="${3:-false}"
  if [ -z "${!var_name:-}" ]; then
    if [ "$secret" = "true" ]; then
      read -r -s -p "$prompt_text: " "$var_name"
      echo
    else
      read -r -p "$prompt_text: " "$var_name"
    fi
  fi
}

require_cmd ssh
require_cmd scp
require_cmd sed

if [ -f "$CONFIG_FILE" ]; then
  # shellcheck disable=SC1090
  source "$CONFIG_FILE"
fi

prompt_if_empty TARGET_HOST "Target host (IPv4/FQDN)"
prompt_if_empty TARGET_PORT "SSH port"
prompt_if_empty TARGET_ROOT_USER "Root SSH user"
prompt_if_empty TARGET_ROOT_SSH_KEY "Path to root SSH private key"
prompt_if_empty OPERATOR_USER "Operator username"
prompt_if_empty OPERATOR_SSH_PUBKEY_PATH "Path to operator public key"
prompt_if_empty GITHUB_REPO "GitHub repo (owner/name)"
prompt_if_empty GITHUB_USERNAME "GitHub username (for GHCR pull secret)"
prompt_if_empty GITHUB_PAT "GitHub PAT" true
prompt_if_empty OPENAI_API_KEY "OPENAI_API_KEY" true
prompt_if_empty STAGING_NAMESPACE "Staging namespace"
prompt_if_empty CODEXK8S_IMAGE "codex-k8s image"

ENABLE_GITHUB_RUNNER="${ENABLE_GITHUB_RUNNER:-true}"
RUNNER_MIN="${RUNNER_MIN:-0}"
RUNNER_MAX="${RUNNER_MAX:-2}"
RUNNER_NAMESPACE="${RUNNER_NAMESPACE:-actions-runner-staging}"
RUNNER_SCALE_SET_NAME="${RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"
RUNNER_IMAGE="${RUNNER_IMAGE:-ghcr.io/actions/actions-runner:latest}"
INSTALL_LONGHORN="${INSTALL_LONGHORN:-false}"
TARGET_PORT="${TARGET_PORT:-22}"
TARGET_ROOT_USER="${TARGET_ROOT_USER:-root}"
LEARNING_MODE_DEFAULT="${LEARNING_MODE_DEFAULT-true}"

[ -f "$TARGET_ROOT_SSH_KEY" ] || die "SSH private key not found: $TARGET_ROOT_SSH_KEY"
[ -f "$OPERATOR_SSH_PUBKEY_PATH" ] || die "Operator public key not found: $OPERATOR_SSH_PUBKEY_PATH"
OPERATOR_SSH_PUBKEY="$(cat "$OPERATOR_SSH_PUBKEY_PATH")"

POSTGRES_DB="codex_k8s"
POSTGRES_USER="codex_k8s"
POSTGRES_PASSWORD="$(rand_hex 24)"
APP_SECRET_KEY="$(rand_hex 32)"
TOKEN_ENCRYPTION_KEY="$(rand_hex 32)"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

REMOTE_DIR="/root/codex-k8s-bootstrap"
REMOTE_ENV="${REMOTE_DIR}/bootstrap.env"

cat > "${TMP_DIR}/bootstrap.env" <<ENV
OPERATOR_USER='$(escape_squote "$OPERATOR_USER")'
OPERATOR_SSH_PUBKEY='$(escape_squote "$OPERATOR_SSH_PUBKEY")'
GITHUB_REPO='$(escape_squote "$GITHUB_REPO")'
GITHUB_USERNAME='$(escape_squote "$GITHUB_USERNAME")'
GITHUB_PAT='$(escape_squote "$GITHUB_PAT")'
OPENAI_API_KEY='$(escape_squote "$OPENAI_API_KEY")'
CONTEXT7_API_KEY='$(escape_squote "${CONTEXT7_API_KEY:-}")'
STAGING_NAMESPACE='$(escape_squote "$STAGING_NAMESPACE")'
STAGING_DOMAIN='$(escape_squote "${STAGING_DOMAIN:-}")'
LETSENCRYPT_EMAIL='$(escape_squote "${LETSENCRYPT_EMAIL:-}")'
CODEXK8S_IMAGE='$(escape_squote "$CODEXK8S_IMAGE")'
ENABLE_GITHUB_RUNNER='$(escape_squote "$ENABLE_GITHUB_RUNNER")'
RUNNER_MIN='$(escape_squote "$RUNNER_MIN")'
RUNNER_MAX='$(escape_squote "$RUNNER_MAX")'
RUNNER_NAMESPACE='$(escape_squote "$RUNNER_NAMESPACE")'
RUNNER_SCALE_SET_NAME='$(escape_squote "$RUNNER_SCALE_SET_NAME")'
RUNNER_IMAGE='$(escape_squote "$RUNNER_IMAGE")'
INSTALL_LONGHORN='$(escape_squote "$INSTALL_LONGHORN")'
LEARNING_MODE_DEFAULT='$(escape_squote "$LEARNING_MODE_DEFAULT")'
POSTGRES_DB='$(escape_squote "$POSTGRES_DB")'
POSTGRES_USER='$(escape_squote "$POSTGRES_USER")'
POSTGRES_PASSWORD='$(escape_squote "$POSTGRES_PASSWORD")'
APP_SECRET_KEY='$(escape_squote "$APP_SECRET_KEY")'
TOKEN_ENCRYPTION_KEY='$(escape_squote "$TOKEN_ENCRYPTION_KEY")'
ENV

log "Copy remote bootstrap scripts to ${TARGET_ROOT_USER}@${TARGET_HOST}:${REMOTE_DIR}"
ssh -i "$TARGET_ROOT_SSH_KEY" -p "$TARGET_PORT" "${TARGET_ROOT_USER}@${TARGET_HOST}" "mkdir -p '${REMOTE_DIR}'"
scp -i "$TARGET_ROOT_SSH_KEY" -P "$TARGET_PORT" -r "${ROOT_DIR}/remote" "${TARGET_ROOT_USER}@${TARGET_HOST}:${REMOTE_DIR}/"
scp -i "$TARGET_ROOT_SSH_KEY" -P "$TARGET_PORT" "${TMP_DIR}/bootstrap.env" "${TARGET_ROOT_USER}@${TARGET_HOST}:${REMOTE_ENV}"

log "Run remote bootstrap"
ssh -i "$TARGET_ROOT_SSH_KEY" -p "$TARGET_PORT" "${TARGET_ROOT_USER}@${TARGET_HOST}" \
  "bash '${REMOTE_DIR}/remote/bootstrap_staging.sh' '${REMOTE_ENV}'"

log "Bootstrap completed"
log "Next: push to main should trigger staging deploy workflow once runner is online"
