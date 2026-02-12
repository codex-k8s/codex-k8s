#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[agent-runner] %s\n' "$*" >&2
}

require_env() {
  local key="$1"
  if [ -z "${!key:-}" ]; then
    log "missing required env: ${key}"
    exit 2
  fi
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    log "missing command: $1"
    exit 2
  }
}

read_json_file_or_empty() {
  local path="$1"
  if [ -z "$path" ] || [ ! -f "$path" ]; then
    printf '%s' "null"
    return 0
  fi
  if ! jq -e . "$path" >/dev/null 2>&1; then
    printf '%s' "null"
    return 0
  fi
  cat "$path"
}

latest_session_file() {
  local sessions_dir="$1"
  if [ ! -d "$sessions_dir" ]; then
    return 0
  fi
  find "$sessions_dir" -type f -name '*.json' -printf '%T@ %p\n' 2>/dev/null \
    | sort -nr \
    | awk 'NR==1 { print $2 }'
}

bearer_auth_header() {
  printf 'Authorization: Bearer %s' "$CODEXK8S_MCP_BEARER_TOKEN"
}

api_post_json() {
  local endpoint="$1"
  local payload="$2"
  curl -fsS -X POST "${CONTROL_PLANE_HTTP_BASE}${endpoint}" \
    -H "$(bearer_auth_header)" \
    -H "Content-Type: application/json" \
    --data "$payload"
}

api_get_latest_session() {
  curl -fsS -G "${CONTROL_PLANE_HTTP_BASE}/internal/agent/session/latest" \
    -H "$(bearer_auth_header)" \
    --data-urlencode "repository_full_name=${CODEXK8S_REPOSITORY_FULL_NAME}" \
    --data-urlencode "branch_name=${TARGET_BRANCH}"
}

emit_event() {
  local event_type="$1"
  local payload="${2:-{}}"
  local event_body
  event_body="$(jq -cn --arg event_type "$event_type" --argjson payload "$payload" '{event_type:$event_type, payload:$payload}')"
  if ! api_post_json "/internal/agent/event" "$event_body" >/dev/null; then
    log "warning: failed to emit event ${event_type}"
  fi
}

persist_session_snapshot() {
  local status="$1"
  local finished_at="$2"
  local issue_number_json="null"
  local pr_number_json="null"
  local report_json="{}"
  local codex_session_json="null"

  if [ "${ISSUE_NUMBER}" -gt 0 ]; then
    issue_number_json="${ISSUE_NUMBER}"
  fi
  if [ "${PR_NUMBER}" -gt 0 ]; then
    pr_number_json="${PR_NUMBER}"
  fi
  if [ -n "${REPORT_JSON}" ] && jq -e . >/dev/null 2>&1 <<<"${REPORT_JSON}"; then
    report_json="${REPORT_JSON}"
  fi
  if [ -n "${SESSION_FILE_PATH}" ]; then
    codex_session_json="$(read_json_file_or_empty "${SESSION_FILE_PATH}")"
  fi

  local payload
  payload="$(
    jq -cn \
      --arg run_id "$CODEXK8S_RUN_ID" \
      --arg correlation_id "$CODEXK8S_CORRELATION_ID" \
      --arg project_id "${CODEXK8S_PROJECT_ID:-}" \
      --arg repository_full_name "$CODEXK8S_REPOSITORY_FULL_NAME" \
      --arg branch_name "$TARGET_BRANCH" \
      --arg pr_url "$PR_URL" \
      --arg trigger_kind "$TRIGGER_KIND" \
      --arg template_kind "$PROMPT_TEMPLATE_KIND" \
      --arg template_source "$PROMPT_TEMPLATE_SOURCE" \
      --arg template_locale "$PROMPT_TEMPLATE_LOCALE" \
      --arg model "$MODEL" \
      --arg reasoning_effort "$REASONING_EFFORT" \
      --arg status "$status" \
      --arg session_id "$SESSION_ID" \
      --arg codex_cli_session_path "$SESSION_FILE_PATH" \
      --arg finished_at "$finished_at" \
      --argjson issue_number "$issue_number_json" \
      --argjson pr_number "$pr_number_json" \
      --argjson session_json "$report_json" \
      --argjson codex_cli_session_json "$codex_session_json" \
      '{
        run_id: $run_id,
        correlation_id: $correlation_id,
        project_id: $project_id,
        repository_full_name: $repository_full_name,
        issue_number: $issue_number,
        branch_name: $branch_name,
        pr_number: $pr_number,
        pr_url: $pr_url,
        trigger_kind: $trigger_kind,
        template_kind: $template_kind,
        template_source: $template_source,
        template_locale: $template_locale,
        model: $model,
        reasoning_effort: $reasoning_effort,
        status: $status,
        session_id: $session_id,
        session_json: $session_json,
        codex_cli_session_path: $codex_cli_session_path,
        codex_cli_session_json: $codex_cli_session_json
      } + (if ($finished_at|length) > 0 then {finished_at:$finished_at} else {} end)'
  )"
  api_post_json "/internal/agent/session" "$payload" >/dev/null
}

restore_session_from_latest() {
  local latest_json
  if ! latest_json="$(api_get_latest_session)"; then
    return 3
  fi
  if [ -z "$latest_json" ] || ! jq -e . >/dev/null 2>&1 <<<"$latest_json"; then
    return 3
  fi
  if [ "$(jq -r '.found // false' <<<"$latest_json")" != "true" ]; then
    return 1
  fi

  EXISTING_PR_NUMBER="$(jq -r '.session.pr_number // 0' <<<"$latest_json")"
  if ! [[ "${EXISTING_PR_NUMBER}" =~ ^[0-9]+$ ]] || [ "${EXISTING_PR_NUMBER}" -le 0 ]; then
    return 2
  fi

  local codex_json
  codex_json="$(jq -c '.session.codex_cli_session_json // null' <<<"$latest_json")"
  if [ "${codex_json}" = "null" ]; then
    return 0
  fi

  mkdir -p "$SESSIONS_DIR"
  RESTORED_SESSION_PATH="${SESSIONS_DIR}/restored-${CODEXK8S_RUN_ID}.json"
  printf '%s\n' "$codex_json" > "${RESTORED_SESSION_PATH}"
  emit_event "${EVENT_RUN_AGENT_SESSION_RESTORED}" "$(jq -cn --arg path "$RESTORED_SESSION_PATH" '{restored_session_path:$path}')"
  return 0
}

fail_revise_pr_not_found() {
  emit_event "${EVENT_RUN_REVISE_PR_NOT_FOUND}" "$(jq -cn --arg branch "$TARGET_BRANCH" '{branch:$branch, reason:"pr_not_found"}')"
  persist_session_snapshot "failed_precondition" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  FINALIZED=1
  exit 42
}

EVENT_RUN_AGENT_STARTED="run.agent.started"
EVENT_RUN_AGENT_SESSION_RESTORED="run.agent.session.restored"
EVENT_RUN_AGENT_SESSION_SAVED="run.agent.session.saved"
EVENT_RUN_AGENT_RESUME_USED="run.agent.resume.used"
EVENT_RUN_PR_CREATED="run.pr.created"
EVENT_RUN_PR_UPDATED="run.pr.updated"
EVENT_RUN_REVISE_PR_NOT_FOUND="run.revise.pr_not_found"

require_cmd jq
require_cmd curl
require_cmd git
require_cmd codex

require_env CODEXK8S_RUN_ID
require_env CODEXK8S_CORRELATION_ID
require_env CODEXK8S_REPOSITORY_FULL_NAME
require_env CODEXK8S_MCP_BASE_URL
require_env CODEXK8S_MCP_BEARER_TOKEN
require_env CODEXK8S_GIT_BOT_TOKEN
require_env CODEXK8S_OPENAI_API_KEY

HOME_DIR="${HOME:-/root}"
CODEX_DIR="${HOME_DIR}/.codex"
SESSIONS_DIR="${CODEX_DIR}/sessions"
WORKSPACE_ROOT="/workspace"
REPO_DIR="${WORKSPACE_ROOT}/repo"
mkdir -p "${CODEX_DIR}" "${SESSIONS_DIR}" "${WORKSPACE_ROOT}"

CONTROL_PLANE_HTTP_BASE="${CODEXK8S_MCP_BASE_URL}"
CONTROL_PLANE_HTTP_BASE="${CONTROL_PLANE_HTTP_BASE%/}"
if [[ "${CONTROL_PLANE_HTTP_BASE}" == */mcp ]]; then
  CONTROL_PLANE_HTTP_BASE="${CONTROL_PLANE_HTTP_BASE%/mcp}"
fi

ISSUE_NUMBER=0
if [[ "${CODEXK8S_ISSUE_NUMBER:-0}" =~ ^[0-9]+$ ]]; then
  ISSUE_NUMBER="${CODEXK8S_ISSUE_NUMBER}"
fi

TRIGGER_KIND="$(printf '%s' "${CODEXK8S_RUN_TRIGGER_KIND:-dev}" | tr '[:upper:]' '[:lower:]')"
if [ "${TRIGGER_KIND}" != "dev_revise" ]; then
  TRIGGER_KIND="dev"
fi

PROMPT_TEMPLATE_KIND="${CODEXK8S_PROMPT_TEMPLATE_KIND:-work}"
if [ "${TRIGGER_KIND}" = "dev_revise" ]; then
  PROMPT_TEMPLATE_KIND="review"
fi
PROMPT_TEMPLATE_SOURCE="${CODEXK8S_PROMPT_TEMPLATE_SOURCE:-repo_seed}"
PROMPT_TEMPLATE_LOCALE="${CODEXK8S_PROMPT_TEMPLATE_LOCALE:-ru}"
MODEL="${CODEXK8S_AGENT_MODEL:-gpt-5.3-codex}"
REASONING_EFFORT="${CODEXK8S_AGENT_REASONING_EFFORT:-medium}"
BASE_BRANCH="${CODEXK8S_AGENT_BASE_BRANCH:-main}"

if [ "${ISSUE_NUMBER}" -gt 0 ]; then
  TARGET_BRANCH="codex/issue-${ISSUE_NUMBER}"
else
  TARGET_BRANCH="codex/run-${CODEXK8S_RUN_ID:0:12}"
fi

SESSION_FILE_PATH=""
SESSION_ID=""
RESTORED_SESSION_PATH=""
EXISTING_PR_NUMBER=0
RESTORE_STATUS=0
PR_NUMBER=0
PR_URL=""
REPORT_JSON=""
FINALIZED=0

on_exit() {
  local exit_code=$?
  if [ "${FINALIZED}" -eq 1 ]; then
    return "${exit_code}"
  fi
  if [ "${exit_code}" -ne 0 ]; then
    local finished_at
    finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    persist_session_snapshot "failed" "${finished_at}" || true
    emit_event "${EVENT_RUN_AGENT_SESSION_SAVED}" "$(jq -cn --arg status "failed" '{status:$status}')" || true
  fi
  return "${exit_code}"
}
trap on_exit EXIT

emit_event "${EVENT_RUN_AGENT_STARTED}" "$(jq -cn \
  --arg branch "$TARGET_BRANCH" \
  --arg trigger_kind "$TRIGGER_KIND" \
  --arg model "$MODEL" \
  --arg reasoning_effort "$REASONING_EFFORT" \
  '{branch:$branch, trigger_kind:$trigger_kind, model:$model, reasoning_effort:$reasoning_effort}')"

if [ "${TRIGGER_KIND}" = "dev_revise" ]; then
  if restore_session_from_latest; then
    RESTORE_STATUS=0
  else
    RESTORE_STATUS=$?
    case "${RESTORE_STATUS}" in
      1)
        ;;
      2)
        fail_revise_pr_not_found
        ;;
      *)
        log "failed to restore agent session snapshot (status=${RESTORE_STATUS})"
        exit 5
        ;;
    esac
  fi
fi

REPO_URL="https://x-access-token:${CODEXK8S_GIT_BOT_TOKEN}@github.com/${CODEXK8S_REPOSITORY_FULL_NAME}.git"
rm -rf "${REPO_DIR}"
git clone "${REPO_URL}" "${REPO_DIR}" >/dev/null 2>&1
git -C "${REPO_DIR}" config user.name "${CODEXK8S_GIT_AUTHOR_NAME:-codex-bot}"
git -C "${REPO_DIR}" config user.email "${CODEXK8S_GIT_AUTHOR_EMAIL:-codex-bot@codex-k8s.local}"
git -C "${REPO_DIR}" fetch origin "${BASE_BRANCH}" --depth=50 >/dev/null 2>&1 || true

if git -C "${REPO_DIR}" ls-remote --exit-code --heads origin "${TARGET_BRANCH}" >/dev/null 2>&1; then
  git -C "${REPO_DIR}" checkout -B "${TARGET_BRANCH}" "origin/${TARGET_BRANCH}" >/dev/null 2>&1
  if [ "${TRIGGER_KIND}" = "dev_revise" ] && [ "${RESTORE_STATUS}" -eq 1 ]; then
    log "resume snapshot is missing; continuing revise with existing branch context"
  fi
else
  if [ "${TRIGGER_KIND}" = "dev_revise" ]; then
    fail_revise_pr_not_found
  fi
  git -C "${REPO_DIR}" checkout -B "${TARGET_BRANCH}" "origin/${BASE_BRANCH}" >/dev/null 2>&1
fi

printf '%s' "${CODEXK8S_OPENAI_API_KEY}" | codex login --with-api-key >/dev/null 2>&1 || true

cat > "${CODEX_DIR}/config.toml" <<EOF
model = "${MODEL}"
model_reasoning_effort = "${REASONING_EFFORT}"
approval_policy = "never"
sandbox_mode = "danger-full-access"
model_verbosity = "low"
web_search_request = true

[history]
persistence = "save-all"

[mcp_servers.codex_k8s]
url = "${CODEXK8S_MCP_BASE_URL}"
bearer_token_env_var = "CODEXK8S_MCP_BEARER_TOKEN"
tool_timeout_sec = 180
EOF

TEMPLATE_PATH="docs/product/prompt-seeds/dev-work.md"
if [ "${PROMPT_TEMPLATE_KIND}" = "review" ]; then
  TEMPLATE_PATH="docs/product/prompt-seeds/dev-review.md"
fi
if [ ! -f "${REPO_DIR}/${TEMPLATE_PATH}" ]; then
  log "template not found: ${TEMPLATE_PATH}"
  exit 3
fi
TASK_BODY="$(cat "${REPO_DIR}/${TEMPLATE_PATH}")"

OUTPUT_SCHEMA_FILE="/tmp/codex-output-schema.json"
cat > "${OUTPUT_SCHEMA_FILE}" <<'EOF'
{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "branch": { "type": "string" },
    "pr_number": { "type": "integer", "minimum": 1 },
    "pr_url": { "type": "string", "minLength": 1 },
    "session_id": { "type": "string" },
    "model": { "type": "string" },
    "reasoning_effort": { "type": "string" }
  },
  "required": ["summary", "branch", "pr_number", "pr_url"],
  "additionalProperties": true
}
EOF

PROMPT_FILE="/tmp/codex-prompt.md"
{
  printf '%s\n' "You are system agent dev for codex-k8s."
  printf '%s\n' "Repository: ${CODEXK8S_REPOSITORY_FULL_NAME}"
  printf '%s\n' "Run ID: ${CODEXK8S_RUN_ID}"
  printf '%s\n' "Issue number: ${ISSUE_NUMBER}"
  printf '%s\n' "Target branch: ${TARGET_BRANCH}"
  printf '%s\n' "Base branch: ${BASE_BRANCH}"
  printf '%s\n' "Trigger kind: ${TRIGGER_KIND}"
  if [ "${TRIGGER_KIND}" = "dev_revise" ] && [ "${EXISTING_PR_NUMBER}" -gt 0 ]; then
    printf '%s\n' "Existing PR number: ${EXISTING_PR_NUMBER}"
  fi
  printf '%s\n' ""
  printf '%s\n' "Mandatory rules:"
  printf '%s\n' "- Use MCP tools for issue/pr/comment/label operations."
  printf '%s\n' "- Git token may be used only for git fetch/commit/push transport."
  printf '%s\n' "- For dev run create or update PR to ${BASE_BRANCH}; for revise update existing PR only."
  printf '%s\n' "- Keep branch ${TARGET_BRANCH}."
  printf '%s\n' "- Return ONLY JSON object matching output schema."
  printf '%s\n' ""
  printf '%s\n' "Task body:"
  printf '%s\n' "${TASK_BODY}"
} > "${PROMPT_FILE}"

LAST_MESSAGE_FILE="/tmp/codex-last-message.json"
PROMPT_CONTENT="$(cat "${PROMPT_FILE}")"

if [ -n "${RESTORED_SESSION_PATH}" ]; then
  emit_event "${EVENT_RUN_AGENT_RESUME_USED}" "$(jq -cn --arg restored_session_path "${RESTORED_SESSION_PATH}" '{restored_session_path:$restored_session_path}')"
  codex exec resume --last --cd "${REPO_DIR}" --output-schema "${OUTPUT_SCHEMA_FILE}" --last-message-file "${LAST_MESSAGE_FILE}" "${PROMPT_CONTENT}"
else
  codex exec --cd "${REPO_DIR}" --output-schema "${OUTPUT_SCHEMA_FILE}" --last-message-file "${LAST_MESSAGE_FILE}" "${PROMPT_CONTENT}"
fi

REPORT_JSON="$(cat "${LAST_MESSAGE_FILE}")"
if ! jq -e . >/dev/null 2>&1 <<<"${REPORT_JSON}"; then
  log "invalid codex result json"
  exit 4
fi

PR_NUMBER="$(jq -r '.pr_number // 0' <<<"${REPORT_JSON}")"
if ! [[ "${PR_NUMBER}" =~ ^[0-9]+$ ]] || [ "${PR_NUMBER}" -le 0 ]; then
  log "invalid pr_number in codex result"
  exit 4
fi
PR_URL="$(jq -r '.pr_url // ""' <<<"${REPORT_JSON}")"
if [ -z "${PR_URL}" ]; then
  log "missing pr_url in codex result"
  exit 4
fi

REPORT_BRANCH="$(jq -r '.branch // ""' <<<"${REPORT_JSON}")"
if [ -n "${REPORT_BRANCH}" ] && [ "${REPORT_BRANCH}" != "${TARGET_BRANCH}" ]; then
  log "codex reported branch ${REPORT_BRANCH}, forcing ${TARGET_BRANCH}"
fi

SESSION_FILE_PATH="$(latest_session_file "${SESSIONS_DIR}")"
if [ -n "${SESSION_FILE_PATH}" ]; then
  SESSION_ID="$(jq -r '.session_id // .id // .conversation_id // .thread_id // empty' "${SESSION_FILE_PATH}" 2>/dev/null || true)"
fi
if [ -z "${SESSION_ID}" ]; then
  SESSION_ID="$(jq -r '.session_id // empty' <<<"${REPORT_JSON}")"
fi

git -C "${REPO_DIR}" push origin "${TARGET_BRANCH}" >/dev/null 2>&1 || true

persist_session_snapshot "succeeded" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
emit_event "${EVENT_RUN_AGENT_SESSION_SAVED}" "$(jq -cn --arg status "succeeded" '{status:$status}')"

if [ "${TRIGGER_KIND}" = "dev_revise" ]; then
  emit_event "${EVENT_RUN_PR_UPDATED}" "$(jq -cn --arg branch "${TARGET_BRANCH}" --arg pr_url "${PR_URL}" --argjson pr_number "${PR_NUMBER}" '{branch:$branch, pr_url:$pr_url, pr_number:$pr_number}')"
else
  emit_event "${EVENT_RUN_PR_CREATED}" "$(jq -cn --arg branch "${TARGET_BRANCH}" --arg pr_url "${PR_URL}" --argjson pr_number "${PR_NUMBER}" '{branch:$branch, pr_url:$pr_url, pr_number:$pr_number}')"
fi

FINALIZED=1
log "completed successfully: branch=${TARGET_BRANCH} pr=${PR_NUMBER}"
