#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

require_root

CODEXK8S_FIREWALL_ENABLED="${CODEXK8S_FIREWALL_ENABLED:-true}"
CODEXK8S_SSH_PORT="${CODEXK8S_SSH_PORT:-22}"

if [ "$CODEXK8S_FIREWALL_ENABLED" != "true" ]; then
  log "Firewall hardening disabled by CODEXK8S_FIREWALL_ENABLED=${CODEXK8S_FIREWALL_ENABLED}"
  exit 0
fi

case "$CODEXK8S_SSH_PORT" in
  ''|*[!0-9]*) die "CODEXK8S_SSH_PORT must be an integer";;
esac

if [ "$CODEXK8S_SSH_PORT" -lt 1 ] || [ "$CODEXK8S_SSH_PORT" -gt 65535 ]; then
  die "CODEXK8S_SSH_PORT out of range: ${CODEXK8S_SSH_PORT}"
fi

if ! command -v nft >/dev/null 2>&1; then
  log "Install nftables"
  apt-get update -y
  apt-get install -y nftables
fi

log "Apply host firewall policy (allow tcp:${CODEXK8S_SSH_PORT},80,443 only)"
cat > /etc/nftables.conf <<EOF
#!/usr/sbin/nft -f

table inet codexk8s_fw {
  chain input {
    type filter hook input priority -5; policy drop;
    ct state { established, related } accept
    iifname "lo" accept

    # Keep in-cluster and pod-to-host traffic functional.
    iifname "cni0" accept
    iifname "flannel.1" accept

    # Keep DHCP renew functional on VPS where DHCP is used.
    udp sport 67 udp dport 68 accept

    # Basic diagnostics + IPv6 neighbor discovery/path MTU.
    ip protocol icmp accept
    ip6 nexthdr ipv6-icmp accept

    # Public ingress surface.
    tcp dport { ${CODEXK8S_SSH_PORT}, 80, 443 } accept
  }

  chain forward {
    type filter hook forward priority -5; policy drop;
    ct state { established, related } accept

    # Allow pod egress and pod-internal forwarding.
    iifname "cni0" accept
    iifname "flannel.1" accept
  }

  chain output {
    type filter hook output priority -5; policy accept;
  }
}
EOF

nft delete table inet codexk8s_fw >/dev/null 2>&1 || true
systemctl enable nftables >/dev/null 2>&1 || true
systemctl restart nftables
nft list table inet codexk8s_fw >/dev/null

log "Firewall policy applied"
