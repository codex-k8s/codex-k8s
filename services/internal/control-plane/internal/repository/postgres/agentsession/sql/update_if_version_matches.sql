-- name: agentsession__update_if_version_matches :one
UPDATE agent_sessions
SET correlation_id = $2,
    project_id = $3::uuid,
    repository_full_name = $4,
    agent_key = $5,
    issue_number = $6,
    branch_name = $7,
    pr_number = $8,
    pr_url = $9,
    trigger_kind = $10,
    template_kind = $11,
    template_source = $12,
    template_locale = $13,
    model = $14,
    reasoning_effort = $15,
    status = $16,
    session_id = $17,
    session_json = COALESCE($18::jsonb, '{}'::jsonb),
    codex_cli_session_path = $19,
    codex_cli_session_json = $20::jsonb,
    started_at = COALESCE($21, started_at),
    finished_at = COALESCE($22, finished_at),
    snapshot_version = CASE
        WHEN COALESCE(snapshot_checksum, '') = COALESCE(NULLIF($24, ''), '') THEN snapshot_version
        ELSE snapshot_version + 1
    END,
    snapshot_checksum = NULLIF($24, ''),
    snapshot_updated_at = CASE
        WHEN COALESCE(snapshot_checksum, '') = COALESCE(NULLIF($24, ''), '') THEN snapshot_updated_at
        ELSE COALESCE($25, NOW())
    END,
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND snapshot_version = $23
RETURNING snapshot_version, snapshot_checksum, snapshot_updated_at;
