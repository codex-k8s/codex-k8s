#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

KODEX_PRODUCTION_NAMESPACE="${KODEX_PRODUCTION_NAMESPACE:-kodex-prod}"
KODEX_INTERNAL_REGISTRY_SERVICE="${KODEX_INTERNAL_REGISTRY_SERVICE:-kodex-registry}"
KODEX_INTERNAL_REGISTRY_HOST="${KODEX_INTERNAL_REGISTRY_HOST:-127.0.0.1:5000}"
KODEX_FIREWALL_ENABLED="${KODEX_FIREWALL_ENABLED:-true}"
KODEX_SSH_PORT="${KODEX_SSH_PORT:-22}"

log "Production namespace: ${KODEX_PRODUCTION_NAMESPACE}"
log "Internal registry endpoint (no auth, node loopback ${KODEX_INTERNAL_REGISTRY_HOST})"
if [ "$KODEX_FIREWALL_ENABLED" = "true" ] && command -v nft >/dev/null 2>&1; then
  log "Firewall policy active (public tcp ports: ${KODEX_SSH_PORT},80,443):"
  nft list table inet kodex_fw >/dev/null 2>&1 && echo "  nft table inet kodex_fw: present" || echo "  nft table inet kodex_fw: missing"
fi

log "Bootstrap finished. Recommended checks:"
log "  sudo cat /etc/rancher/k3s/registries.yaml"
log "  sudo cat /var/lib/rancher/k3s/agent/etc/kubelet.conf.d/10-kodex-image-gc.conf"
log "  sudo systemctl status kodex-image-prune.timer --no-pager"
log "  sudo nft list table inet kodex_fw"
log "  go run ./services/internal/control-plane/cmd/runtime-deploy --prerequisites-only --env-file /root/kodex-bootstrap/bootstrap.env --kubeconfig /etc/rancher/k3s/k3s.yaml"
