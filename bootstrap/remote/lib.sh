#!/usr/bin/env bash
set -euo pipefail

log() { echo "[$(date -Is)] $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

require_root() {
  [ "${EUID}" -eq 0 ] || die "Run as root"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

load_env_file() {
  local env_file="$1"
  [ -f "$env_file" ] || die "Env file not found: $env_file"
  set -a
  # shellcheck disable=SC1090
  source "$env_file"
  set +a
}

kube_env() {
  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
}

repo_dir() {
  echo "/opt/codex-k8s"
}
