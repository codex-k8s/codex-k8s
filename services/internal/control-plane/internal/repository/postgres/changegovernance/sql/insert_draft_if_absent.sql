-- name: changegovernance__insert_draft_if_absent :one
WITH existing_signal AS (
    SELECT 1
    FROM change_governance_internal_drafts
    WHERE signal_id = $3
),
cleared_latest AS (
    UPDATE change_governance_internal_drafts
    SET is_latest = false
    WHERE package_id = $1::uuid
      AND is_latest = true
      AND NOT EXISTS (SELECT 1 FROM existing_signal)
)
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
SELECT
    $1::uuid,
    $2::uuid,
    $3,
    $4,
    $5,
    $6,
    $7::jsonb,
    true,
    COALESCE($8::timestamptz, NOW())
WHERE NOT EXISTS (SELECT 1 FROM existing_signal)
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
