-- name: changegovernance__insert_evidence_block :exec
INSERT INTO change_governance_evidence_blocks (
    package_id,
    wave_id,
    block_kind,
    state,
    verification_state,
    required_by_tier,
    source_kind,
    artifact_links_json,
    latest_signal_id,
    observed_at
)
VALUES (
    $1::uuid,
    $2::uuid,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8::jsonb,
    $9,
    COALESCE($10::timestamptz, NOW())
);
