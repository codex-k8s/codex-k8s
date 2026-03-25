-- name: changegovernance__list_drafts :many
SELECT
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
    created_at
FROM change_governance_internal_drafts
WHERE package_id = $1::uuid
ORDER BY occurred_at DESC, id DESC;
