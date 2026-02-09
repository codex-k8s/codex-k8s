-- name: runqueue__claim_next_pending_for_update :one
SELECT id, correlation_id, project_id, learning_mode, run_payload
FROM agent_runs
WHERE status = 'pending'
ORDER BY created_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
