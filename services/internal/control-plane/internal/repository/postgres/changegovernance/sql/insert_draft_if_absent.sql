-- name: changegovernance__insert_draft_if_absent :one
INSERT INTO change_governance_internal_drafts (
    package_id,
    run_id,
    signal_id,
    draft_ref,
    draft_checksum,
    draft_kind,
    metadata_json,
    is_latest,
    occurred_at
)
VALUES (
    $1::uuid,
    $2::uuid,
    $3,
    $4,
    $5,
    $6,
    $7::jsonb,
    true,
    COALESCE($8::timestamptz, NOW())
)
ON CONFLICT (signal_id) DO NOTHING
RETURNING
    id::text AS id,
    package_id::text AS package_id,
    run_id::text AS run_id,
    signal_id,
    draft_ref,
    draft_checksum,
    draft_kind,
    metadata_json,
    is_latest,
    occurred_at,
    created_at;
