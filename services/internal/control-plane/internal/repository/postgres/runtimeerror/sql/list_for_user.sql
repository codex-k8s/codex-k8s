-- name: runtimeerror__list_for_user :many
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
JOIN project_members pm
  ON pm.project_id = re.project_id
WHERE pm.user_id = $1
  AND (
      $3::text = 'all'
      OR ($3::text = 'active' AND re.viewed_at IS NULL)
      OR ($3::text = 'viewed' AND re.viewed_at IS NOT NULL)
  )
  AND ($4::text IS NULL OR re.level = $4::text)
  AND ($5::text IS NULL OR re.source = $5::text)
  AND ($6::text IS NULL OR re.run_id::text = $6::text)
  AND ($7::text IS NULL OR re.correlation_id = $7::text)
ORDER BY re.created_at DESC
LIMIT $2;
