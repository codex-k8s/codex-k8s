-- name: missioncontrol__list_timeline_entries :many
SELECT
    id,
    project_id::text AS project_id,
    entity_id,
    source_kind,
    entry_external_key,
    command_id::text AS command_id,
    summary,
    body_markdown,
    payload AS payload_json,
    occurred_at,
    provider_url,
    is_read_only,
    created_at
FROM mission_control_timeline_entries
WHERE project_id = $1
  AND entity_id = $2
ORDER BY occurred_at DESC, id DESC
LIMIT $3;
