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
apt-get install -y curl ca-certificates jq git gh tar unzip gettext-base open-iscsi nfs-common
systemctl enable --now iscsid || true

GO_VERSION="${CODEXK8S_GO_VERSION:-1.25.6}"

if ! command -v go >/dev/null 2>&1 || ! go version | grep -q "go${GO_VERSION}"; then
  log "Install Go ${GO_VERSION}"
  arch="$(uname -m)"
  case "${arch}" in
    x86_64) go_arch="amd64" ;;
    aarch64) go_arch="arm64" ;;
    *)
      die "Unsupported CPU architecture for Go install: ${arch}"
      ;;
  esac
  tmp="$(mktemp -d)"
  tarball="go${GO_VERSION}.linux-${go_arch}.tar.gz"
  curl -fsSL -o "${tmp}/${tarball}" "https://go.dev/dl/${tarball}"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "${tmp}/${tarball}"
  ln -sf /usr/local/go/bin/go /usr/local/bin/go
  ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
  rm -rf "${tmp}"
fi

if ! command -v helm >/dev/null 2>&1; then
  log "Install helm"
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/get-helm-3" https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
  chmod 700 "${tmp}/get-helm-3"
  "${tmp}/get-helm-3"
  rm -rf "$tmp"
fi
