-- name: configentry__exists :one
SELECT EXISTS (
    SELECT 1
    FROM config_entries
    WHERE scope = $1::text
      AND project_id IS NOT DISTINCT FROM NULLIF($2::text, '')::uuid
      AND repository_id IS NOT DISTINCT FROM NULLIF($3::text, '')::uuid
      AND key = $4::text
) AS exists;

