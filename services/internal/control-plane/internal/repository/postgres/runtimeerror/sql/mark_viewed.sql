-- name: runtimeerror__mark_viewed :one
UPDATE runtime_errors re
SET
    viewed_at = COALESCE(re.viewed_at, NOW()),
    viewed_by = COALESCE(re.viewed_by, $2::uuid)
WHERE re.id = $1
RETURNING
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
    re.created_at;
