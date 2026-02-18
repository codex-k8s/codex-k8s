-- name: runtimeerror__list_all :many
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
WHERE (
    $2::text = 'all'
    OR ($2::text = 'active' AND re.viewed_at IS NULL)
    OR ($2::text = 'viewed' AND re.viewed_at IS NOT NULL)
)
  AND ($3::text IS NULL OR re.level = $3::text)
  AND ($4::text IS NULL OR re.source = $4::text)
  AND ($5::text IS NULL OR re.run_id::text = $5::text)
  AND ($6::text IS NULL OR re.correlation_id = $6::text)
ORDER BY re.created_at DESC
LIMIT $1;
