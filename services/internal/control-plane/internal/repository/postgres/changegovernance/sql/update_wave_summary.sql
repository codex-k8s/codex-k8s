-- name: changegovernance__update_wave_summary :exec
UPDATE change_governance_waves
SET
    evidence_completeness_state = $2,
    verification_minimum_state = $3,
    updated_at = NOW()
WHERE id = $1::uuid;
