-- name: changegovernance__list_evidence_blocks :many
SELECT
    id::text AS id,
    package_id::text AS package_id,
    wave_id::text AS wave_id,
    block_kind,
    state,
    verification_state,
    required_by_tier,
    source_kind,
    artifact_links_json,
    latest_signal_id,
    observed_at,
    created_at,
    updated_at
FROM change_governance_evidence_blocks
WHERE package_id = $1::uuid
ORDER BY observed_at DESC, id DESC;
