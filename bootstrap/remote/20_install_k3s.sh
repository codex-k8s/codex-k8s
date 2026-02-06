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

CODEXK8S_NODE_DISCOVERY_TIMEOUT="${CODEXK8S_NODE_DISCOVERY_TIMEOUT:-300}"
CODEXK8S_NODE_READY_TIMEOUT="${CODEXK8S_NODE_READY_TIMEOUT:-1200s}"

case "$CODEXK8S_NODE_DISCOVERY_TIMEOUT" in
  ''|*[!0-9]*) die "CODEXK8S_NODE_DISCOVERY_TIMEOUT must be integer seconds";;
esac

deadline=$((SECONDS + CODEXK8S_NODE_DISCOVERY_TIMEOUT))
while [ "$SECONDS" -lt "$deadline" ]; do
  if kubectl get nodes >/dev/null 2>&1; then
    break
  fi
  sleep 5
done

kubectl get nodes >/dev/null 2>&1 || die "k3s node discovery timed out after ${CODEXK8S_NODE_DISCOVERY_TIMEOUT}s"
kubectl wait --for=condition=Ready node --all --timeout="${CODEXK8S_NODE_READY_TIMEOUT}"
