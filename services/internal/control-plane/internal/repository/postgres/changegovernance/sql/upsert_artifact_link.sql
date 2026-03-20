-- name: changegovernance__upsert_artifact_link :exec
INSERT INTO change_governance_artifact_links (
    package_id,
    artifact_kind,
    artifact_ref,
    relation_kind,
    display_label
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (package_id, artifact_kind, artifact_ref, relation_kind) DO NOTHING;
