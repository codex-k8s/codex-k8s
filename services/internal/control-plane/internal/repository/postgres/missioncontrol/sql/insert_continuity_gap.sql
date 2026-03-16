-- name: missioncontrol__insert_continuity_gap :exec
INSERT INTO mission_control_continuity_gaps (
    project_id,
    subject_entity_id,
    gap_kind,
    severity,
    status,
    expected_entity_kind,
    expected_stage_label,
    resolution_hint,
    payload,
    detected_at,
    updated_at
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    'open',
    $5,
    $6,
    $7,
    $8,
    COALESCE($9, NOW()),
    COALESCE($10, NOW())
);
