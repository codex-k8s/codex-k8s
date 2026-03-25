-- name: changegovernance__deactivate_current_projections :exec
UPDATE change_governance_projection_snapshots
SET is_current = false
WHERE package_id = $1::uuid
  AND is_current = true;
