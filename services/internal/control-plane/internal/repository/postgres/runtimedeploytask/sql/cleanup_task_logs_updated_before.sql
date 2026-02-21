-- name: runtimedeploytask__cleanup_task_logs_updated_before :exec
UPDATE runtime_deploy_tasks
SET logs_json = '[]'::jsonb,
    updated_at = NOW()
WHERE updated_at < $1::timestamptz
  AND COALESCE(logs_json, '[]'::jsonb) <> '[]'::jsonb;
