#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

require_root

log "Disable swap and apply sysctl for k8s"
swapoff -a || true
sed -ri 's/^([^#].*\sswap\s.*)$/#\1/g' /etc/fstab || true
modprobe br_netfilter || true
cat >/etc/sysctl.d/99-k8s.conf <<'SYSCTL'
net.ipv4.ip_forward=1
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
SYSCTL
sysctl --system >/dev/null

log "Install base packages"
apt-get update -y
apt-get install -y curl ca-certificates jq git tar unzip gettext-base open-iscsi nfs-common golang-go
systemctl enable --now iscsid || true
