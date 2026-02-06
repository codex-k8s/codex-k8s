#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

: "${GITHUB_REPO:?GITHUB_REPO is required}"
: "${GITHUB_PAT:?GITHUB_PAT is required}"
: "${OPENAI_API_KEY:?OPENAI_API_KEY is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${APP_SECRET_KEY:?APP_SECRET_KEY is required}"
: "${TOKEN_ENCRYPTION_KEY:?TOKEN_ENCRYPTION_KEY is required}"

STAGING_NAMESPACE="${STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_IMAGE="${CODEXK8S_IMAGE:-ghcr.io/codex-k8s/codex-k8s:latest}"
CODEXK8S_WAIT_ROLLOUT="${CODEXK8S_WAIT_ROLLOUT:-false}"
RUNNER_SCALE_SET_NAME="${RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"
LEARNING_MODE_DEFAULT="${LEARNING_MODE_DEFAULT-true}"

log "Configure GitHub repository variables/secrets for ${GITHUB_REPO}"
printf %s "${GITHUB_PAT}" | gh auth login --with-token

# Variables used by staging deploy workflow.
gh variable set CODEXK8S_STAGING_NAMESPACE -R "${GITHUB_REPO}" --body "${STAGING_NAMESPACE}"
gh variable set CODEXK8S_IMAGE -R "${GITHUB_REPO}" --body "${CODEXK8S_IMAGE}"
gh variable set CODEXK8S_WAIT_ROLLOUT -R "${GITHUB_REPO}" --body "${CODEXK8S_WAIT_ROLLOUT}"
gh variable set CODEXK8S_STAGING_RUNNER -R "${GITHUB_REPO}" --body "${RUNNER_SCALE_SET_NAME}"
if [ -n "${LEARNING_MODE_DEFAULT}" ]; then
  gh variable set CODEXK8S_LEARNING_MODE_DEFAULT -R "${GITHUB_REPO}" --body "${LEARNING_MODE_DEFAULT}"
else
  gh variable delete CODEXK8S_LEARNING_MODE_DEFAULT -R "${GITHUB_REPO}" >/dev/null 2>&1 || true
fi
if [ -n "${STAGING_DOMAIN:-}" ]; then
  gh variable set CODEXK8S_STAGING_DOMAIN -R "${GITHUB_REPO}" --body "${STAGING_DOMAIN}"
fi
if [ -n "${POSTGRES_DB:-}" ]; then
  gh variable set CODEXK8S_POSTGRES_DB -R "${GITHUB_REPO}" --body "${POSTGRES_DB}"
fi
if [ -n "${POSTGRES_USER:-}" ]; then
  gh variable set CODEXK8S_POSTGRES_USER -R "${GITHUB_REPO}" --body "${POSTGRES_USER}"
fi

# Secrets used by staging deploy workflow.
gh secret set OPENAI_API_KEY -R "${GITHUB_REPO}" --body "${OPENAI_API_KEY}"
gh secret set CODEXK8S_POSTGRES_PASSWORD -R "${GITHUB_REPO}" --body "${POSTGRES_PASSWORD}"
gh secret set CODEXK8S_APP_SECRET_KEY -R "${GITHUB_REPO}" --body "${APP_SECRET_KEY}"
gh secret set CODEXK8S_TOKEN_ENCRYPTION_KEY -R "${GITHUB_REPO}" --body "${TOKEN_ENCRYPTION_KEY}"
if [ -n "${CONTEXT7_API_KEY:-}" ]; then
  gh secret set CONTEXT7_API_KEY -R "${GITHUB_REPO}" --body "${CONTEXT7_API_KEY}"
fi

log "GitHub repository bootstrap CI settings configured"
