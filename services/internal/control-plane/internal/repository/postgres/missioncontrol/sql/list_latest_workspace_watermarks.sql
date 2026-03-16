-- name: missioncontrol__list_latest_workspace_watermarks :many
SELECT
    id,
    project_id,
    watermark_kind,
    status,
    summary,
    window_started_at,
    window_ended_at,
    observed_at,
    payload_json,
    created_at
FROM (
    SELECT
        id,
        project_id::text AS project_id,
        watermark_kind,
        status,
        summary,
        window_started_at,
        window_ended_at,
        observed_at,
        payload AS payload_json,
        created_at,
        ROW_NUMBER() OVER (
            PARTITION BY watermark_kind
            ORDER BY observed_at DESC, id DESC
        ) AS row_no
    FROM mission_control_workspace_watermarks
    WHERE project_id = $1::uuid
) AS latest_watermarks
WHERE row_no = 1
ORDER BY observed_at DESC, id DESC;
