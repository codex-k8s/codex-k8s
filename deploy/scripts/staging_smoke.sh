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
require_cmd openssl
require_cmd awk
require_cmd base64

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

echo "[smoke] verify migrate job completed"
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get job codex-k8s-migrate >/dev/null
kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" wait --for=condition=complete "job/codex-k8s-migrate" --timeout=600s

echo "[smoke] load postgres credentials from secret"
pg_db="$(
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get secret codex-k8s-postgres \
    -o jsonpath='{.data.CODEXK8S_POSTGRES_DB}' | base64 -d
)"
pg_user="$(
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get secret codex-k8s-postgres \
    -o jsonpath='{.data.CODEXK8S_POSTGRES_USER}' | base64 -d
)"
pg_pass="$(
  kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get secret codex-k8s-postgres \
    -o jsonpath='{.data.CODEXK8S_POSTGRES_PASSWORD}' | base64 -d
)"
[ -n "${pg_db}" ] || { echo "[smoke] FAIL: empty CODEXK8S_POSTGRES_DB" >&2; exit 1; }
[ -n "${pg_user}" ] || { echo "[smoke] FAIL: empty CODEXK8S_POSTGRES_USER" >&2; exit 1; }
[ -n "${pg_pass}" ] || { echo "[smoke] FAIL: empty CODEXK8S_POSTGRES_PASSWORD" >&2; exit 1; }

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
  if [ "${code}" != "401" ]; then
    echo "[smoke] FAIL: expected 401 for invalid signature, got ${code}" >&2
    exit 1
  fi

  echo "[smoke] webhook idempotency (202 accepted, then 200 duplicate)"
  webhook_secret="$(
    kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" get secret codex-k8s-runtime \
      -o jsonpath='{.data.CODEXK8S_GITHUB_WEBHOOK_SECRET}' | base64 -d
  )"
  payload='{"installation":{"id":1},"repository":{"id":1,"full_name":"smoke/test","name":"test","private":true},"sender":{"id":1,"login":"smoke"}}'
  sig_hex="$(printf '%s' "${payload}" | openssl dgst -sha256 -hmac "${webhook_secret}" | awk '{print $2}')"
  sig="sha256=${sig_hex}"
  delivery="smoke-$(date +%s)-$$"

  code1="$(
    curl -sk -o /dev/null -w '%{http_code}' \
      -X POST "https://${CODEXK8S_STAGING_DOMAIN}/api/v1/webhooks/github" \
      -H 'X-GitHub-Event: ping' \
      -H "X-GitHub-Delivery: ${delivery}" \
      -H "X-Hub-Signature-256: ${sig}" \
      -d "${payload}" || true
  )"
  echo "[smoke] webhook first HTTP status=${code1}"
  if [ "${code1}" != "202" ]; then
    echo "[smoke] FAIL: expected 202 for first webhook, got ${code1}" >&2
    exit 1
  fi

  code2="$(
    curl -sk -o /dev/null -w '%{http_code}' \
      -X POST "https://${CODEXK8S_STAGING_DOMAIN}/api/v1/webhooks/github" \
      -H 'X-GitHub-Event: ping' \
      -H "X-GitHub-Delivery: ${delivery}" \
      -H "X-Hub-Signature-256: ${sig}" \
      -d "${payload}" || true
  )"
  echo "[smoke] webhook replay HTTP status=${code2}"
  if [ "${code2}" != "200" ]; then
    echo "[smoke] FAIL: expected 200 for replay webhook, got ${code2}" >&2
    exit 1
  fi

  echo "[smoke] worker run loop (pending -> running -> succeeded/failed) for correlation_id=${delivery}"
  deadline=$((SECONDS + 180))
  run_status=""
  run_project_id=""
  while [ "$SECONDS" -lt "$deadline" ]; do
    row="$(
      kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" exec postgres-0 -- \
        env PGPASSWORD="${pg_pass}" psql -U "${pg_user}" -d "${pg_db}" -tA \
          -c "SELECT status, COALESCE(project_id::text, '') FROM agent_runs WHERE correlation_id='${delivery}' LIMIT 1;" 2>/dev/null || true
    )"
    run_status="$(printf '%s' "$row" | awk -F'|' '{print $1}')"
    run_project_id="$(printf '%s' "$row" | awk -F'|' '{print $2}')"

    if [ -z "${run_status}" ]; then
      sleep 2
      continue
    fi

    if [ "${run_status}" = "succeeded" ] || [ "${run_status}" = "failed" ] || [ "${run_status}" = "canceled" ]; then
      break
    fi

    sleep 2
  done

  if [ -z "${run_status}" ]; then
    echo "[smoke] FAIL: agent_runs row was not created for correlation_id=${delivery}" >&2
    exit 1
  fi
  if [ "${run_status}" != "succeeded" ] && [ "${run_status}" != "failed" ] && [ "${run_status}" != "canceled" ]; then
    echo "[smoke] FAIL: run did not reach final status in time (status=${run_status})" >&2
    exit 1
  fi

  # Slots are DB-backed: once a run is finished, it should not keep a leased slot.
  if [ -n "${run_project_id}" ]; then
    leased="$(
      kubectl -n "${CODEXK8S_STAGING_NAMESPACE}" exec postgres-0 -- \
        env PGPASSWORD="${pg_pass}" psql -U "${pg_user}" -d "${pg_db}" -tA \
          -c "SELECT COUNT(*) FROM slots WHERE project_id='${run_project_id}'::uuid AND state <> 'free';" 2>/dev/null || true
    )"
    leased="$(printf '%s' "${leased:-0}" | tr -d '[:space:]')"
    if [ -n "${leased}" ] && [ "${leased}" != "0" ]; then
      echo "[smoke] FAIL: slot leak detected for project_id=${run_project_id} leased=${leased}" >&2
      exit 1
    fi
  fi

  echo "[smoke] staff allowlist enforcement (expected 403 for unknown email via X-Auth-Request-Email)"
  code3="$(
    curl -sS -o /dev/null -w '%{http_code}' \
      "http://127.0.0.1:${CODEXK8S_SMOKE_PORTFWD_PORT}/api/v1/auth/me" \
      -H 'X-Auth-Request-Email: not-allowed-smoke@example.com' \
      -H 'X-Auth-Request-User: not-allowed-smoke' || true
  )"
  echo "[smoke] staff /api/v1/auth/me HTTP status=${code3}"
  if [ "${code3}" != "403" ]; then
    echo "[smoke] FAIL: expected 403 for not allowed email, got ${code3}" >&2
    exit 1
  fi
fi

echo "[smoke] OK"
