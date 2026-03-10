-- name: runtimedeploytask__cancel_superseded_deploy_only :exec
UPDATE runtime_deploy_tasks
SET
    status = 'canceled',
    lease_owner = NULL,
    lease_until = NULL,
    cancel_requested_at = NOW(),
    cancel_requested_by = 'system',
    cancel_reason = $7,
    last_error = $7,
    terminal_status_source = 'system',
    terminal_event_seq = terminal_event_seq + 1,
    finished_at = NOW(),
    updated_at = NOW()
WHERE run_id <> $1::uuid
  AND deploy_only = TRUE
  AND status IN ('pending', 'running')
  AND repository_full_name = $2::text
  AND target_env = $3::text
  AND (
      (
          NULLIF($4::text, '') IS NOT NULL
          AND NULLIF(namespace, '') IS NOT NULL
          AND namespace = $4::text
      )
      OR (
          (
              NULLIF($4::text, '') IS NULL
              OR NULLIF(namespace, '') IS NULL
          )
          AND slot_no = $5::int
      )
  )
  AND COALESCE(NULLIF(build_ref, ''), '') <> $6::text;
