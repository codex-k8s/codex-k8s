#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

: "${CODEXK8S_GITHUB_REPO:?CODEXK8S_GITHUB_REPO is required}"
: "${CODEXK8S_GITHUB_PAT:?CODEXK8S_GITHUB_PAT is required}"

REPO_DIR="$(repo_dir)"
GIT_REMOTE_URL="https://github.com/${CODEXK8S_GITHUB_REPO}.git"
GIT_AUTH_HEADER="AUTHORIZATION: basic $(printf 'x-access-token:%s' "${CODEXK8S_GITHUB_PAT}" | base64 | tr -d '\n')"

if [ ! -d "$REPO_DIR/.git" ]; then
  log "Clone repository $CODEXK8S_GITHUB_REPO"
  git -c "http.https://github.com/.extraheader=${GIT_AUTH_HEADER}" clone "${GIT_REMOTE_URL}" "$REPO_DIR"
else
  log "Update repository $CODEXK8S_GITHUB_REPO"
  git -C "$REPO_DIR" remote set-url origin "${GIT_REMOTE_URL}"
  git -C "$REPO_DIR" -c "http.https://github.com/.extraheader=${GIT_AUTH_HEADER}" fetch --all --prune
  # Bootstrap treats /opt/codex-k8s as an immutable deployment source.
  # Force-reset the working tree to avoid rerun failures caused by manual hotfixes or leftover files.
  git -C "$REPO_DIR" checkout -f main
  git -C "$REPO_DIR" reset --hard origin/main
  git -C "$REPO_DIR" clean -fd
fi
