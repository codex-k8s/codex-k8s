-- name: missioncontrol__list_continuity_gaps :many
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
  AND (
      COALESCE(array_length($2::bigint[], 1), 0) = 0
      OR subject_entity_id = ANY($2::bigint[])
  )
  AND (
      COALESCE(array_length($3::text[], 1), 0) = 0
      OR status = ANY($3::text[])
  )
ORDER BY detected_at DESC, id DESC;
