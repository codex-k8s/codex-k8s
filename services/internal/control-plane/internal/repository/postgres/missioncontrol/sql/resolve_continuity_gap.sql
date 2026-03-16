-- name: missioncontrol__resolve_continuity_gap :exec
UPDATE mission_control_continuity_gaps
SET
    status = 'resolved',
    resolution_entity_id = NULL,
    resolved_at = COALESCE($3, NOW()),
    updated_at = COALESCE($3, NOW())
WHERE project_id = $1::uuid
  AND id = $2
  AND status = 'open';
