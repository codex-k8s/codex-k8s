-- name: runtimeerror__insert :one
INSERT INTO runtime_errors (
    source,
    level,
    message,
    details_json,
    stack_trace,
    correlation_id,
    run_id,
    project_id,
    namespace,
    job_name
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
    id,
    source,
    level,
    message,
    details_json,
    stack_trace,
    correlation_id,
    run_id::text AS run_id,
    project_id::text AS project_id,
    namespace,
    job_name,
    viewed_at,
    viewed_by::text AS viewed_by,
    created_at;
