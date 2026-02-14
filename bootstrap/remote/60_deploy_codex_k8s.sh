#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/lib.sh"
load_env_file "${BOOTSTRAP_ENV_FILE:?}"

kube_env

REPO_DIR="$(repo_dir)"
CODEXK8S_STAGING_NAMESPACE="${CODEXK8S_STAGING_NAMESPACE:-codex-k8s-ai-staging}"
CODEXK8S_DEPLOY_ENVIRONMENT="${CODEXK8S_DEPLOY_ENVIRONMENT:-ai-staging}"
CODEXK8S_LETSENCRYPT_SERVER="${CODEXK8S_LETSENCRYPT_SERVER:-https://acme-v02.api.letsencrypt.org/directory}"

: "${CODEXK8S_STAGING_DOMAIN:?CODEXK8S_STAGING_DOMAIN is required}"
: "${CODEXK8S_LETSENCRYPT_EMAIL:?CODEXK8S_LETSENCRYPT_EMAIL is required}"

ensure_domain_resolves "$CODEXK8S_STAGING_DOMAIN"

log "Apply cert-manager ClusterIssuer for staging domain"
export CODEXK8S_LETSENCRYPT_EMAIL CODEXK8S_LETSENCRYPT_SERVER
envsubst < "${REPO_DIR}/deploy/base/cert-manager/clusterissuer.yaml.tpl" | kubectl apply -f -
kubectl wait --for=condition=Ready clusterissuer/codex-k8s-letsencrypt --timeout=600s

log "Run declarative reconcile via codex-bootstrap"
export CODEXK8S_STAGING_NAMESPACE CODEXK8S_DEPLOY_ENVIRONMENT
export CODEXK8S_SERVICES_CONFIG_PATH="${REPO_DIR}/services.yaml"
bash "${REPO_DIR}/deploy/scripts/deploy_staging.sh"
