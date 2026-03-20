-- name: changegovernance__get_evidence_block_by_scope :one
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
  AND (
      ($2::uuid IS NULL AND wave_id IS NULL)
      OR wave_id = $2::uuid
  )
  AND block_kind = $3
FOR UPDATE;
