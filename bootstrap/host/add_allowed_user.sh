#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_FILE="${ROOT_DIR}/host/config.env"

die() { echo "ERROR: $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

require_cmd ssh

[ -f "$CONFIG_FILE" ] || die "Missing config: ${CONFIG_FILE}"
# shellcheck disable=SC1090
source "$CONFIG_FILE"

IS_ADMIN="false"
if [ "${1:-}" = "--admin" ]; then
  IS_ADMIN="true"
  shift
fi

EMAIL="${1:-}"
[ -n "$EMAIL" ] || die "Usage: $0 [--admin] <email>"

TARGET_PORT="${TARGET_PORT:-22}"
TARGET_ROOT_USER="${TARGET_ROOT_USER:-root}"
[ -n "${TARGET_HOST:-}" ] || die "TARGET_HOST is required in ${CONFIG_FILE}"
[ -n "${TARGET_ROOT_SSH_KEY:-}" ] || die "TARGET_ROOT_SSH_KEY is required in ${CONFIG_FILE}"
[ -n "${CODEXK8S_STAGING_NAMESPACE:-}" ] || die "CODEXK8S_STAGING_NAMESPACE is required in ${CONFIG_FILE}"

ns="${CODEXK8S_STAGING_NAMESPACE}"

ssh -i "$TARGET_ROOT_SSH_KEY" -p "$TARGET_PORT" "${TARGET_ROOT_USER}@${TARGET_HOST}" "set -euo pipefail
if command -v kubectl >/dev/null 2>&1; then K='kubectl'; else K='k3s kubectl'; fi
ns='${ns}'
email='${EMAIL}'

db=\$(sudo \$K -n \"\$ns\" get secret codex-k8s-postgres -o jsonpath='{.data.CODEXK8S_POSTGRES_DB}' | base64 -d)
user=\$(sudo \$K -n \"\$ns\" get secret codex-k8s-postgres -o jsonpath='{.data.CODEXK8S_POSTGRES_USER}' | base64 -d)
pass=\$(sudo \$K -n \"\$ns\" get secret codex-k8s-postgres -o jsonpath='{.data.CODEXK8S_POSTGRES_PASSWORD}' | base64 -d)

sql=\"INSERT INTO users (email, is_platform_admin) VALUES (LOWER('\$email'), FALSE)
ON CONFLICT (email) DO UPDATE SET updated_at = NOW()
RETURNING id, email, is_platform_admin;\"

if [ '${IS_ADMIN}' = 'true' ]; then
  sql=\"INSERT INTO users (email, is_platform_admin) VALUES (LOWER('\$email'), TRUE)
ON CONFLICT (email) DO UPDATE SET is_platform_admin = TRUE, updated_at = NOW()
RETURNING id, email, is_platform_admin;\"
fi

sudo \$K -n \"\$ns\" exec postgres-0 -- sh -ec \"PGPASSWORD=\\\"\$pass\\\" psql -U \\\"\$user\\\" -d \\\"\$db\\\" -v ON_ERROR_STOP=1 -c \\\"\$sql\\\"\""
