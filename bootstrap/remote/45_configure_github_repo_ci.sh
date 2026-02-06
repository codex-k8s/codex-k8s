#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"
: "${CODEXK8S_OPENAI_API_KEY:?CODEXK8S_OPENAI_API_KEY is required}"
: "${CODEXK8S_POSTGRES_PASSWORD:?CODEXK8S_POSTGRES_PASSWORD is required}"
: "${CODEXK8S_APP_SECRET_KEY:?CODEXK8S_APP_SECRET_KEY is required}"
: "${CODEXK8S_TOKEN_ENCRYPTION_KEY:?CODEXK8S_TOKEN_ENCRYPTION_KEY is required}"

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_IMAGE="${CODEXK8S_IMAGE:-ghcr.io/codex-k8s/codex-k8s:latest}"
CODEXK8S_WAIT_ROLLOUT="${CODEXK8S_WAIT_ROLLOUT:-false}"
CODEXK8S_RUNNER_SCALE_SET_NAME="${CODEXK8S_RUNNER_SCALE_SET_NAME:-codex-k8s-ai-staging}"
CODEXK8S_LEARNING_MODE_DEFAULT="${CODEXK8S_LEARNING_MODE_DEFAULT-true}"

log "Configure GitHub repository variables/secrets for ${CODEXK8S_GITHUB_REPO}"
printf %s "${CODEXK8S_GITHUB_PAT}" | gh auth login --with-token

# Variables used by staging deploy workflow.
gh variable set CODEXK8S_STAGING_NAMESPACE -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_STAGING_NAMESPACE}"
gh variable set CODEXK8S_IMAGE -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_IMAGE}"
gh variable set CODEXK8S_WAIT_ROLLOUT -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_WAIT_ROLLOUT}"
gh variable set CODEXK8S_STAGING_RUNNER -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_RUNNER_SCALE_SET_NAME}"
if [ -n "${CODEXK8S_GITHUB_USERNAME:-}" ]; then
  # GitHub Actions reserves the GITHUB_* prefix, use CODEXK8S_* instead.
  gh variable set CODEXK8S_GITHUB_USERNAME -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_GITHUB_USERNAME}"
fi
if [ -n "${CODEXK8S_LEARNING_MODE_DEFAULT}" ]; then
  gh variable set CODEXK8S_LEARNING_MODE_DEFAULT -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_LEARNING_MODE_DEFAULT}"
else
  gh variable delete CODEXK8S_LEARNING_MODE_DEFAULT -R "${CODEXK8S_GITHUB_REPO}" >/dev/null 2>&1 || true
fi
if [ -n "${CODEXK8S_STAGING_DOMAIN:-}" ]; then
  gh variable set CODEXK8S_STAGING_DOMAIN -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_STAGING_DOMAIN}"
fi
if [ -n "${CODEXK8S_POSTGRES_DB:-}" ]; then
  gh variable set CODEXK8S_POSTGRES_DB -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_POSTGRES_DB}"
fi
if [ -n "${CODEXK8S_POSTGRES_USER:-}" ]; then
  gh variable set CODEXK8S_POSTGRES_USER -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_POSTGRES_USER}"
fi

# Secrets used by staging deploy workflow.
gh secret set CODEXK8S_OPENAI_API_KEY -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_OPENAI_API_KEY}"
gh secret set CODEXK8S_POSTGRES_PASSWORD -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_POSTGRES_PASSWORD}"
gh secret set CODEXK8S_APP_SECRET_KEY -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_APP_SECRET_KEY}"
gh secret set CODEXK8S_TOKEN_ENCRYPTION_KEY -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_TOKEN_ENCRYPTION_KEY}"
if [ -n "${CODEXK8S_GITHUB_USERNAME:-}" ]; then
  gh secret set CODEXK8S_GITHUB_USERNAME -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_GITHUB_USERNAME}"
fi
if [ -n "${CODEXK8S_CONTEXT7_API_KEY:-}" ]; then
  gh secret set CODEXK8S_CONTEXT7_API_KEY -R "${CODEXK8S_GITHUB_REPO}" --body "${CODEXK8S_CONTEXT7_API_KEY}"
fi

# Cleanup legacy non-prefixed names to keep a single naming convention.
for legacy_secret in OPENAI_API_KEY CONTEXT7_API_KEY; do
  gh secret delete "${legacy_secret}" -R "${CODEXK8S_GITHUB_REPO}" >/dev/null 2>&1 || true
done
for legacy_var in STAGING_NAMESPACE STAGING_DOMAIN POSTGRES_DB POSTGRES_USER LEARNING_MODE_DEFAULT RUNNER_NAMESPACE RUNNER_SCALE_SET_NAME RUNNER_MIN RUNNER_MAX RUNNER_IMAGE; do
  gh variable delete "${legacy_var}" -R "${CODEXK8S_GITHUB_REPO}" >/dev/null 2>&1 || true
done

log "GitHub repository bootstrap CI settings configured"
