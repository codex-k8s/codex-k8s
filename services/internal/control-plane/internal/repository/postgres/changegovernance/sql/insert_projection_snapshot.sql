-- name: changegovernance__insert_projection_snapshot :exec
INSERT INTO change_governance_projection_snapshots (
    package_id,
    projection_kind,
    projection_version,
    is_current,
    payload_json,
    refreshed_at
)
VALUES (
    $1::uuid,
    $2,
    $3,
    true,
    $4::jsonb,
    COALESCE($5::timestamptz, NOW())
);
