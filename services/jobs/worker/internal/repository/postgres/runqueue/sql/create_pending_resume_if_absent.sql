-- name: runqueue__create_pending_resume_if_absent :one
WITH source_run AS (
    SELECT project_id, agent_id, run_payload, learning_mode
    FROM agent_runs
    WHERE id = $2
)
INSERT INTO agent_runs (
    id,
    correlation_id,
    project_id,
    agent_id,
    status,
    run_payload,
    learning_mode
)
SELECT
    $1,
    $3,
    source_run.project_id,
    source_run.agent_id,
    'pending',
    source_run.run_payload,
    source_run.learning_mode
FROM source_run
ON CONFLICT (correlation_id) DO NOTHING
RETURNING id;
