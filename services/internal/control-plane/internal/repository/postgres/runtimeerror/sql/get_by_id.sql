-- name: runtimeerror__get_by_id :one
SELECT
    re.id,
    re.source,
    re.level,
    re.message,
    re.details_json,
    re.stack_trace,
    re.correlation_id,
    re.run_id::text AS run_id,
    re.project_id::text AS project_id,
    re.namespace,
    re.job_name,
    re.viewed_at,
    re.viewed_by::text AS viewed_by,
    re.created_at
FROM runtime_errors re
WHERE re.id = $1;
