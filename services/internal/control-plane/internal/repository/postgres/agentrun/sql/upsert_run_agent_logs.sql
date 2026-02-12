-- name: agentrun__upsert_run_agent_logs :exec
UPDATE agent_runs
SET agent_logs_json = COALESCE($2::jsonb, '{}'::jsonb),
    updated_at = NOW()
WHERE id = $1::uuid;
