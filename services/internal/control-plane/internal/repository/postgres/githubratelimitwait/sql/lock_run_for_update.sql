-- name: githubratelimitwait__lock_run_for_update :one
SELECT id
FROM agent_runs
WHERE id = $1::uuid
FOR UPDATE;
