#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

: "${GITHUB_REPO:?GITHUB_REPO is required}"
: "${GITHUB_PAT:?GITHUB_PAT is required}"

REPO_DIR="$(repo_dir)"

if [ ! -d "$REPO_DIR/.git" ]; then
  log "Clone repository $GITHUB_REPO"
  git clone "https://x-access-token:${GITHUB_PAT}@github.com/${GITHUB_REPO}.git" "$REPO_DIR"
else
  log "Update repository $GITHUB_REPO"
  git -C "$REPO_DIR" fetch --all --prune
  git -C "$REPO_DIR" checkout main
  git -C "$REPO_DIR" pull --ff-only
fi
