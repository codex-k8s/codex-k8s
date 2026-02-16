-- name: runtimedeploytask__append_log :exec
UPDATE runtime_deploy_tasks
SET
    logs_json = (
        WITH current AS (
            SELECT COALESCE(logs_json, '[]'::jsonb) || jsonb_build_array(jsonb_build_object(
                'stage', $2::text,
                'level', $3::text,
                'message', $4::text,
                'created_at', NOW()
            )) AS value
        )
        SELECT CASE
            WHEN $5::int > 0 AND jsonb_array_length(value) > $5::int THEN (
                SELECT COALESCE(jsonb_agg(elem ORDER BY ord), '[]'::jsonb)
                FROM (
                    SELECT elem, ord
                    FROM jsonb_array_elements(value) WITH ORDINALITY AS v(elem, ord)
                    WHERE ord > jsonb_array_length(value) - $5::int
                ) trimmed
            )
            ELSE value
        END
        FROM current
    ),
    updated_at = NOW()
WHERE run_id = $1::uuid;
