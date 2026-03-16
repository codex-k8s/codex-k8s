-- name: missioncontrol__select_open_continuity_gaps :many
SELECT
    id,
    project_id::text AS project_id,
    subject_entity_id,
    gap_kind,
    severity,
    status,
    expected_entity_kind,
    expected_stage_label,
    resolution_entity_id,
    resolution_hint,
    payload AS payload_json,
    detected_at,
    resolved_at,
    updated_at
FROM mission_control_continuity_gaps
WHERE project_id = $1::uuid
  AND status = 'open'
ORDER BY subject_entity_id ASC, gap_kind ASC, detected_at DESC, id DESC;
