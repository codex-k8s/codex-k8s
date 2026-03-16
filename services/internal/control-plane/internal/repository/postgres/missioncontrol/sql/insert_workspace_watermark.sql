-- name: missioncontrol__insert_workspace_watermark :one
INSERT INTO mission_control_workspace_watermarks (
    project_id,
    watermark_kind,
    status,
    summary,
    window_started_at,
    window_ended_at,
    observed_at,
    payload
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5,
    $6,
    COALESCE($7, NOW()),
    $8
)
RETURNING
    id,
    project_id::text AS project_id,
    watermark_kind,
    status,
    summary,
    window_started_at,
    window_ended_at,
    observed_at,
    payload AS payload_json,
    created_at;
