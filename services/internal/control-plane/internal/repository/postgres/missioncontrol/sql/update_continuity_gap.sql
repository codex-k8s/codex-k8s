-- name: missioncontrol__update_continuity_gap :exec
UPDATE mission_control_continuity_gaps
SET
    severity = $3,
    status = 'open',
    expected_entity_kind = $4,
    expected_stage_label = $5,
    resolution_entity_id = NULL,
    resolution_hint = $6,
    payload = $7,
    detected_at = COALESCE($8, detected_at),
    resolved_at = NULL,
    updated_at = COALESCE($9, NOW())
WHERE project_id = $1::uuid
  AND id = $2;
