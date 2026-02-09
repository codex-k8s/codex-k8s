#!/usr/bin/env bash
set -euo pipefail

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing command: $1" >&2
    exit 1
  }
}

require_cmd kubectl
require_cmd curl

CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_STAGING_DOMAIN="${CODEXK8S_STAGING_DOMAIN:-}"
CODEXK8S_SMOKE_PORTFWD_PORT="${CODEXK8S_SMOKE_PORTFWD_PORT:-18080}"

pf_pid=""
cleanup() {
  if [ -n "${pf_pid}" ] && kill -0 "${pf_pid}" >/dev/null 2>&1; then
    kill "${pf_pid}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

echo "[smoke] namespace=${CODEXK8S_STAGING_NAMESPACE}"

echo "[smoke] rollout status"
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status deploy/codex-k8s --timeout=600s
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status deploy/codex-k8s-worker --timeout=600s
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status deploy/oauth2-proxy --timeout=600s
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" rollout status deploy/codex-k8s-web-console --timeout=600s

echo "[smoke] pods"
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get pods -o wide

echo "[smoke] verify last migrate job completed"
migrate_job="$(
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get jobs -o name 2>/dev/null \
    | sed 's|^job/||' \
    | grep -E '^codex-k8s-migrate-' \
    | tail -n 1 || true
)"
if [ -n "${migrate_job}" ]; then
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=condition=complete "job/${migrate_job}" --timeout=600s
else
  echo "[smoke] WARN: no migrate job found (expected a codex-k8s-migrate-<suffix> job)" >&2
fi

echo "[smoke] port-forward svc/codex-k8s and check /healthz /readyz /metrics"
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" port-forward svc/codex-k8s "${CODEXK8S_SMOKE_PORTFWD_PORT}:80" >/tmp/codexk8s-portfwd.log 2>&1 &
pf_pid="$!"
sleep 2

curl -fsS "http://127.0.0.1:${CODEXK8S_SMOKE_PORTFWD_PORT}/healthz" >/dev/null
curl -fsS "http://127.0.0.1:${CODEXK8S_SMOKE_PORTFWD_PORT}/readyz" >/dev/null
curl -fsS "http://127.0.0.1:${CODEXK8S_SMOKE_PORTFWD_PORT}/metrics" >/dev/null

echo "[smoke] ingress TLS secret exists"
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get secret codex-k8s-staging-tls >/dev/null

if [ -n "${CODEXK8S_STAGING_DOMAIN}" ]; then
  echo "[smoke] webhook endpoint is reachable without OAuth (expected HTTP 401 for invalid signature)"
  code="$(
    curl -sk -o /dev/null -w '%{http_code}' \
      -X POST "https://${CODEXK8S_STAGING_DOMAIN}/api/v1/webhooks/github" \
      -H 'X-GitHub-Event: ping' \
      -H 'X-GitHub-Delivery: smoke-delivery' \
      -H 'X-Hub-Signature-256: sha256=deadbeef' \
      -d '{}' || true
  )"
  echo "[smoke] webhook HTTP status=${code}"
fi

echo "[smoke] OK"

