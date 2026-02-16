#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

CODEXK8S_PRODUCTION_NAMESPACE="${CODEXK8S_PRODUCTION_NAMESPACE:-codex-k8s-prod}"
CODEXK8S_INTERNAL_REGISTRY_SERVICE="${CODEXK8S_INTERNAL_REGISTRY_SERVICE:-codex-k8s-registry}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:5000}"
CODEXK8S_FIREWALL_ENABLED="${CODEXK8S_FIREWALL_ENABLED:-true}"
CODEXK8S_SSH_PORT="${CODEXK8S_SSH_PORT:-22}"

log "Production namespace: ${CODEXK8S_PRODUCTION_NAMESPACE}"
log "Internal registry endpoint (no auth, node loopback ${CODEXK8S_INTERNAL_REGISTRY_HOST})"
if [ "$CODEXK8S_FIREWALL_ENABLED" = "true" ] && command -v nft >/dev/null 2>&1; then
  log "Firewall policy active (public tcp ports: ${CODEXK8S_SSH_PORT},80,443):"
  nft list table inet codexk8s_fw >/dev/null 2>&1 && echo "  nft table inet codexk8s_fw: present" || echo "  nft table inet codexk8s_fw: missing"
fi

log "Bootstrap finished. Recommended checks:"
log "  sudo cat /etc/rancher/k3s/registries.yaml"
log "  sudo nft list table inet codexk8s_fw"
log "  go run ./services/internal/control-plane/cmd/runtime-deploy --prerequisites-only --env-file /root/codex-k8s-bootstrap/bootstrap.env --kubeconfig /etc/rancher/k3s/k3s.yaml"
