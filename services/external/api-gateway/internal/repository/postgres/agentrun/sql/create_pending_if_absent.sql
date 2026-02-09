-- name: agentrun__create_pending_if_absent :one
INSERT INTO agent_runs (
    id,
    correlation_id,
    status,
    run_payload,
    learning_mode
)
VALUES (
    $1,
    $2,
    'pending',
    $3::jsonb,
    false
)
ON CONFLICT (correlation_id) DO NOTHING
RETURNING id;
