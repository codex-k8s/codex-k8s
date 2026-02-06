#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

require_root
require_cmd cmp
require_cmd install

if systemctl list-unit-files | grep -q '^k3s.service'; then
  log "k3s already installed; skip"
else
  log "Install k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
    --write-kubeconfig-mode 600 \
    --disable traefik \
    --disable servicelb" sh -
fi

CODEXK8S_INTERNAL_REGISTRY_PORT="${CODEXK8S_INTERNAL_REGISTRY_PORT:-5000}"
CODEXK8S_INTERNAL_REGISTRY_HOST="${CODEXK8S_INTERNAL_REGISTRY_HOST:-127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}}"
K3S_REGISTRIES_FILE="/etc/rancher/k3s/registries.yaml"
tmp_registries="$(mktemp)"
cat > "${tmp_registries}" <<EOF
mirrors:
  "${CODEXK8S_INTERNAL_REGISTRY_HOST}":
    endpoint:
      - "http://${CODEXK8S_INTERNAL_REGISTRY_HOST}"
configs:
  "${CODEXK8S_INTERNAL_REGISTRY_HOST}":
    tls:
      insecure_skip_verify: true
EOF

if [ ! -f "${K3S_REGISTRIES_FILE}" ] || ! cmp -s "${tmp_registries}" "${K3S_REGISTRIES_FILE}"; then
  log "Configure k3s registry mirror for ${CODEXK8S_INTERNAL_REGISTRY_HOST}"
  mkdir -p "$(dirname "${K3S_REGISTRIES_FILE}")"
  install -m 600 "${tmp_registries}" "${K3S_REGISTRIES_FILE}"
  systemctl restart k3s
fi
rm -f "${tmp_registries}"

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
