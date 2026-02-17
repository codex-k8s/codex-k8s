-- name: repocfg__upsert_preflight_report :exec
UPDATE repositories
SET preflight_report_json = $2::jsonb,
    preflight_updated_at = NOW(),
    updated_at = NOW()
WHERE id = $1::uuid;

