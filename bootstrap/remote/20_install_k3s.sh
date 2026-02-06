#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

require_root

if systemctl list-unit-files | grep -q '^k3s.service'; then
  log "k3s already installed; skip"
else
  log "Install k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
    --write-kubeconfig-mode 600 \
    --disable traefik \
    --disable servicelb" sh -
fi

kube_env
kubectl wait --for=condition=Ready node --all --timeout=300s
