-- name: agentsession__cleanup_session_payloads_finished_before :exec
UPDATE agent_sessions
SET session_json = '{}'::jsonb,
    codex_cli_session_json = NULL,
    updated_at = NOW()
WHERE finished_at IS NOT NULL
  AND finished_at < $1::timestamptz
  AND (
      COALESCE(session_json, '{}'::jsonb) <> '{}'::jsonb
      OR codex_cli_session_json IS NOT NULL
  );
