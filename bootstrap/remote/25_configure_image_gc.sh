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

KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT="${KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT:-35}"
KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT="${KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT:-30}"
KODEX_K3S_IMAGE_PRUNE_TIMER_ENABLED="${KODEX_K3S_IMAGE_PRUNE_TIMER_ENABLED:-true}"
KODEX_K3S_IMAGE_PRUNE_ONCALENDAR="${KODEX_K3S_IMAGE_PRUNE_ONCALENDAR:-*-*-* 04:17:00 UTC}"
KODEX_NODE_DISCOVERY_TIMEOUT="${KODEX_NODE_DISCOVERY_TIMEOUT:-300}"

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

validate_percent "KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT" "$KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT"
validate_percent "KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT" "$KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT"
if [ "$KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT" -ge "$KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT" ]; then
  die "KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT must be lower than KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT"
fi

kubelet_conf_dir="/var/lib/rancher/k3s/agent/etc/kubelet.conf.d"
kubelet_conf_file="${kubelet_conf_dir}/10-kodex-image-gc.conf"
tmp_kubelet_conf="$(mktemp)"
cat > "${tmp_kubelet_conf}" <<EOF
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
imageGCHighThresholdPercent: ${KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT}
imageGCLowThresholdPercent: ${KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT}
EOF

restart_k3s="false"
mkdir -p "${kubelet_conf_dir}"
if [ ! -f "${kubelet_conf_file}" ] || ! cmp -s "${tmp_kubelet_conf}" "${kubelet_conf_file}"; then
  log "Configure kubelet image GC thresholds (${KODEX_K3S_IMAGE_GC_HIGH_THRESHOLD_PERCENT}/${KODEX_K3S_IMAGE_GC_LOW_THRESHOLD_PERCENT})"
  install -m 600 "${tmp_kubelet_conf}" "${kubelet_conf_file}"
  restart_k3s="true"
fi
rm -f "${tmp_kubelet_conf}"

prune_script="/usr/local/sbin/kodex-image-prune.sh"
prune_service="/etc/systemd/system/kodex-image-prune.service"
prune_timer="/etc/systemd/system/kodex-image-prune.timer"

tmp_prune_script="$(mktemp)"
cat > "${tmp_prune_script}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
before="$(df -h / | awk 'NR==2 {print $3 "/" $2 " used (" $5 ")"}')"
echo "[kodex-image-prune] before: ${before}"
/usr/local/bin/k3s crictl --timeout 120s rmi --prune
after="$(df -h / | awk 'NR==2 {print $3 "/" $2 " used (" $5 ")"}')"
echo "[kodex-image-prune] after: ${after}"
EOF
install -m 755 "${tmp_prune_script}" "${prune_script}"
rm -f "${tmp_prune_script}"

tmp_prune_service="$(mktemp)"
cat > "${tmp_prune_service}" <<EOF
[Unit]
Description=kodex host containerd image prune
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
Description=Run kodex host containerd image prune on schedule

[Timer]
OnCalendar=${KODEX_K3S_IMAGE_PRUNE_ONCALENDAR}
Persistent=true

[Install]
WantedBy=timers.target
EOF
install -m 644 "${tmp_prune_timer}" "${prune_timer}"
rm -f "${tmp_prune_timer}"

systemctl daemon-reload

if [ "${KODEX_K3S_IMAGE_PRUNE_TIMER_ENABLED}" = "true" ]; then
  log "Enable kodex host image prune timer (${KODEX_K3S_IMAGE_PRUNE_ONCALENDAR})"
  systemctl enable --now kodex-image-prune.timer
else
  log "Disable kodex host image prune timer"
  systemctl disable --now kodex-image-prune.timer >/dev/null 2>&1 || true
fi

if [ "${restart_k3s}" = "true" ]; then
  log "Restart k3s to apply kubelet image GC thresholds"
  systemctl restart k3s
  deadline=$((SECONDS + KODEX_NODE_DISCOVERY_TIMEOUT))
  while [ "$SECONDS" -lt "$deadline" ]; do
    if systemctl is-active --quiet k3s && [ -S /run/k3s/containerd/containerd.sock ] && [ -s /etc/rancher/k3s/k3s.yaml ]; then
      break
    fi
    sleep 5
  done
  if ! systemctl is-active --quiet k3s; then
    die "k3s service is not active after ${KODEX_NODE_DISCOVERY_TIMEOUT}s"
  fi
fi
