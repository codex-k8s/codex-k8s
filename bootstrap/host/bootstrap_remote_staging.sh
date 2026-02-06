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
prompt_if_empty CODEXK8S_GITHUB_REPO "GitHub repo (owner/name)"
prompt_if_empty CODEXK8S_GITHUB_USERNAME "GitHub username (for GHCR pull secret)"
prompt_if_empty CODEXK8S_GITHUB_PAT "GitHub PAT" true
prompt_if_empty CODEXK8S_OPENAI_API_KEY "CODEXK8S_OPENAI_API_KEY" true
prompt_if_empty CODEXK8S_STAGING_NAMESPACE "Staging namespace"
prompt_if_empty CODEXK8S_IMAGE "codex-k8s image"

CODEXK8S_ENABLE_GITHUB_RUNNER="${CODEXK8S_ENABLE_GITHUB_RUNNER:-true}"
CODEXK8S_RUNNER_MIN="${CODEXK8S_RUNNER_MIN:-0}"
CODEXK8S_RUNNER_MAX="${CODEXK8S_RUNNER_MAX:-2}"
CODEXK8S_RUNNER_NAMESPACE="${CODEXK8S_RUNNER_NAMESPACE:-actions-runner-staging}"
CODEXK8S_RUNNER_SCALE_SET_NAME="${CODEXK8S_RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"
CODEXK8S_RUNNER_IMAGE="${CODEXK8S_RUNNER_IMAGE:-ghcr.io/actions/actions-runner:latest}"
CODEXK8S_INSTALL_LONGHORN="${CODEXK8S_INSTALL_LONGHORN:-false}"
TARGET_PORT="${TARGET_PORT:-22}"
TARGET_ROOT_USER="${TARGET_ROOT_USER:-root}"
CODEXK8S_LEARNING_MODE_DEFAULT="${CODEXK8S_LEARNING_MODE_DEFAULT-true}"

[ -f "$TARGET_ROOT_SSH_KEY" ] || die "SSH private key not found: $TARGET_ROOT_SSH_KEY"
[ -f "$OPERATOR_SSH_PUBKEY_PATH" ] || die "Operator public key not found: $OPERATOR_SSH_PUBKEY_PATH"
OPERATOR_SSH_PUBKEY="$(cat "$OPERATOR_SSH_PUBKEY_PATH")"

CODEXK8S_POSTGRES_DB="codex_k8s"
CODEXK8S_POSTGRES_USER="codex_k8s"
CODEXK8S_POSTGRES_PASSWORD="$(rand_hex 24)"
CODEXK8S_APP_SECRET_KEY="$(rand_hex 32)"
CODEXK8S_TOKEN_ENCRYPTION_KEY="$(rand_hex 32)"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

REMOTE_DIR="/root/codex-k8s-bootstrap"
REMOTE_ENV="${REMOTE_DIR}/bootstrap.env"

cat > "${TMP_DIR}/bootstrap.env" <<ENV
OPERATOR_USER='$(escape_squote "$OPERATOR_USER")'
OPERATOR_SSH_PUBKEY='$(escape_squote "$OPERATOR_SSH_PUBKEY")'
CODEXK8S_GITHUB_REPO='$(escape_squote "$CODEXK8S_GITHUB_REPO")'
CODEXK8S_GITHUB_USERNAME='$(escape_squote "$CODEXK8S_GITHUB_USERNAME")'
CODEXK8S_GITHUB_PAT='$(escape_squote "$CODEXK8S_GITHUB_PAT")'
CODEXK8S_OPENAI_API_KEY='$(escape_squote "$CODEXK8S_OPENAI_API_KEY")'
CODEXK8S_CONTEXT7_API_KEY='$(escape_squote "${CODEXK8S_CONTEXT7_API_KEY:-}")'
CODEXK8S_STAGING_NAMESPACE='$(escape_squote "$CODEXK8S_STAGING_NAMESPACE")'
CODEXK8S_STAGING_DOMAIN='$(escape_squote "${CODEXK8S_STAGING_DOMAIN:-}")'
CODEXK8S_LETSENCRYPT_EMAIL='$(escape_squote "${CODEXK8S_LETSENCRYPT_EMAIL:-}")'
CODEXK8S_IMAGE='$(escape_squote "$CODEXK8S_IMAGE")'
CODEXK8S_ENABLE_GITHUB_RUNNER='$(escape_squote "$CODEXK8S_ENABLE_GITHUB_RUNNER")'
CODEXK8S_RUNNER_MIN='$(escape_squote "$CODEXK8S_RUNNER_MIN")'
CODEXK8S_RUNNER_MAX='$(escape_squote "$CODEXK8S_RUNNER_MAX")'
CODEXK8S_RUNNER_NAMESPACE='$(escape_squote "$CODEXK8S_RUNNER_NAMESPACE")'
CODEXK8S_RUNNER_SCALE_SET_NAME='$(escape_squote "$CODEXK8S_RUNNER_SCALE_SET_NAME")'
CODEXK8S_RUNNER_IMAGE='$(escape_squote "$CODEXK8S_RUNNER_IMAGE")'
CODEXK8S_INSTALL_LONGHORN='$(escape_squote "$CODEXK8S_INSTALL_LONGHORN")'
CODEXK8S_LEARNING_MODE_DEFAULT='$(escape_squote "$CODEXK8S_LEARNING_MODE_DEFAULT")'
CODEXK8S_POSTGRES_DB='$(escape_squote "$CODEXK8S_POSTGRES_DB")'
CODEXK8S_POSTGRES_USER='$(escape_squote "$CODEXK8S_POSTGRES_USER")'
CODEXK8S_POSTGRES_PASSWORD='$(escape_squote "$CODEXK8S_POSTGRES_PASSWORD")'
CODEXK8S_APP_SECRET_KEY='$(escape_squote "$CODEXK8S_APP_SECRET_KEY")'
CODEXK8S_TOKEN_ENCRYPTION_KEY='$(escape_squote "$CODEXK8S_TOKEN_ENCRYPTION_KEY")'
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
