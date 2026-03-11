#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

require_root
require_cmd cmp
require_cmd install
require_cmd systemctl

CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT="${CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT:-35}"
CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT="${CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT:-30}"
CODEXK8S_K3S_IMAGE_PRUNE_TIMER_ENABLED="${CODEXK8S_K3S_IMAGE_PRUNE_TIMER_ENABLED:-true}"
CODEXK8S_K3S_IMAGE_PRUNE_ONCALENDAR="${CODEXK8S_K3S_IMAGE_PRUNE_ONCALENDAR:-*-*-* 04:17:00 UTC}"
CODEXK8S_NODE_DISCOVERY_TIMEOUT="${CODEXK8S_NODE_DISCOVERY_TIMEOUT:-300}"

validate_percent() {
  local name="$1"
  local value="$2"
  case "$value" in
    ''|*[!0-9]*) die "${name} must be an integer percent";;
  esac
  if [ "$value" -lt 1 ] || [ "$value" -gt 100 ]; then
    die "${name} must be between 1 and 100"
  fi
}

validate_percent "CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT" "$CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT"
validate_percent "CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT" "$CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT"
if [ "$CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT" -ge "$CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT" ]; then
  die "CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT must be lower than CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT"
fi

kubelet_conf_dir="/var/lib/rancher/k3s/agent/etc/kubelet.conf.d"
kubelet_conf_file="${kubelet_conf_dir}/10-codex-k8s-image-gc.conf"
tmp_kubelet_conf="$(mktemp)"
cat > "${tmp_kubelet_conf}" <<EOF
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
imageGCHighThresholdPercent: ${CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT}
imageGCLowThresholdPercent: ${CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT}
EOF

restart_k3s="false"
mkdir -p "${kubelet_conf_dir}"
if [ ! -f "${kubelet_conf_file}" ] || ! cmp -s "${tmp_kubelet_conf}" "${kubelet_conf_file}"; then
  log "Configure kubelet image GC thresholds (${CODEXK8S_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT}/${CODEXK8S_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT})"
  install -m 600 "${tmp_kubelet_conf}" "${kubelet_conf_file}"
  restart_k3s="true"
fi
rm -f "${tmp_kubelet_conf}"

prune_script="/usr/local/sbin/codex-k8s-image-prune.sh"
prune_service="/etc/systemd/system/codex-k8s-image-prune.service"
prune_timer="/etc/systemd/system/codex-k8s-image-prune.timer"

tmp_prune_script="$(mktemp)"
cat > "${tmp_prune_script}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
before="$(df -h / | awk 'NR==2 {print $3 "/" $2 " used (" $5 ")"}')"
echo "[codex-k8s-image-prune] before: ${before}"
/usr/local/bin/k3s crictl --timeout 120s rmi --prune
after="$(df -h / | awk 'NR==2 {print $3 "/" $2 " used (" $5 ")"}')"
echo "[codex-k8s-image-prune] after: ${after}"
EOF
install -m 755 "${tmp_prune_script}" "${prune_script}"
rm -f "${tmp_prune_script}"

tmp_prune_service="$(mktemp)"
cat > "${tmp_prune_service}" <<EOF
[Unit]
Description=codex-k8s host containerd image prune
After=k3s.service
Requires=k3s.service

[Service]
Type=oneshot
ExecStart=${prune_script}
EOF
install -m 644 "${tmp_prune_service}" "${prune_service}"
rm -f "${tmp_prune_service}"

tmp_prune_timer="$(mktemp)"
cat > "${tmp_prune_timer}" <<EOF
[Unit]
Description=Run codex-k8s host containerd image prune on schedule

[Timer]
OnCalendar=${CODEXK8S_K3S_IMAGE_PRUNE_ONCALENDAR}
Persistent=true

[Install]
WantedBy=timers.target
EOF
install -m 644 "${tmp_prune_timer}" "${prune_timer}"
rm -f "${tmp_prune_timer}"

systemctl daemon-reload

if [ "${CODEXK8S_K3S_IMAGE_PRUNE_TIMER_ENABLED}" = "true" ]; then
  log "Enable codex-k8s host image prune timer (${CODEXK8S_K3S_IMAGE_PRUNE_ONCALENDAR})"
  systemctl enable --now codex-k8s-image-prune.timer
else
  log "Disable codex-k8s host image prune timer"
  systemctl disable --now codex-k8s-image-prune.timer >/dev/null 2>&1 || true
fi

if [ "${restart_k3s}" = "true" ]; then
  log "Restart k3s to apply kubelet image GC thresholds"
  systemctl restart k3s
  deadline=$((SECONDS + CODEXK8S_NODE_DISCOVERY_TIMEOUT))
  while [ "$SECONDS" -lt "$deadline" ]; do
    if systemctl is-active --quiet k3s && [ -S /run/k3s/containerd/containerd.sock ] && [ -s /etc/rancher/k3s/k3s.yaml ]; then
      break
    fi
    sleep 5
  done
  if ! systemctl is-active --quiet k3s; then
    die "k3s service is not active after ${CODEXK8S_NODE_DISCOVERY_TIMEOUT}s"
  fi
fi
