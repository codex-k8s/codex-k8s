-- name: changegovernance__update_evidence_block :exec
UPDATE change_governance_evidence_blocks
SET
    state = $2,
    verification_state = $3,
    required_by_tier = $4,
    source_kind = 'agent_signal',
    artifact_links_json = $5::jsonb,
    latest_signal_id = $6,
    observed_at = COALESCE($7::timestamptz, NOW()),
    updated_at = NOW()
WHERE id = $1::uuid;
