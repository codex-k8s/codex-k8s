-- name: changegovernance__list_feedback_records :many
SELECT
    id::text AS id,
    package_id::text AS package_id,
    feedback_id,
    gap_kind,
    source_kind,
    severity,
    state,
    suggested_action,
    summary_markdown,
    related_artifact_ref,
    opened_at,
    closed_at,
    created_at,
    updated_at
FROM change_governance_feedback_records
WHERE package_id = $1::uuid
ORDER BY opened_at DESC, id DESC;
