-- name: changegovernance__list_current_projection_snapshots :many
SELECT
    id,
    package_id::text AS package_id,
    projection_kind,
    projection_version,
    is_current,
    payload_json,
    refreshed_at,
    created_at
FROM change_governance_projection_snapshots
WHERE package_id = $1::uuid
  AND is_current = true
ORDER BY projection_kind ASC, id DESC;
