#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

REPO_DIR="$(repo_dir)"
REPO_ARCHIVE="${CODEXK8S_REMOTE_REPO_ARCHIVE:-/root/codex-k8s-bootstrap/repo-src.tgz}"

[ -f "$REPO_ARCHIVE" ] || die "Repository archive not found: $REPO_ARCHIVE"

log "Extract repository archive ${REPO_ARCHIVE} -> ${REPO_DIR}"
rm -rf "$REPO_DIR"
mkdir -p "$REPO_DIR"
tar -xzf "$REPO_ARCHIVE" -C "$REPO_DIR"
