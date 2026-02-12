-- name: agentrun__create_pending_if_absent :one
INSERT INTO agent_runs (
    id,
    correlation_id,
    project_id,
    agent_id,
    status,
    run_payload,
    learning_mode
)
VALUES (
    $1,
    $2,
    NULLIF($3, '')::uuid,
    NULLIF($4, '')::uuid,
    'pending',
    $5::jsonb,
    $6
)
ON CONFLICT (correlation_id) DO NOTHING
RETURNING id;
