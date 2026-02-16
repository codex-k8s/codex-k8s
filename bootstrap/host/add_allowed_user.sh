#!/usr/bin/env bash
set -euo pipefail

cat >&2 <<'MSG'
add_allowed_user.sh is deprecated.
Use one of the following instead:
1) Staff web-console (Admin -> Users)
2) Staff API endpoint for user upsert
3) codex-bootstrap/user-management command once it is enabled in your environment
MSG
exit 1
