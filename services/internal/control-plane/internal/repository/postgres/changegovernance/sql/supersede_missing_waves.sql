-- name: changegovernance__supersede_missing_waves :exec
UPDATE change_governance_waves
SET
    publication_state = 'superseded',
    updated_at = NOW()
WHERE package_id = $1::uuid
  AND NOT (wave_key = ANY($2::text[]));
