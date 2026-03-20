-- name: changegovernance__list_artifact_links :many
SELECT
    id,
    package_id::text AS package_id,
    artifact_kind,
    artifact_ref,
    relation_kind,
    display_label,
    created_at
FROM change_governance_artifact_links
WHERE package_id = $1::uuid
ORDER BY created_at ASC, id ASC;
