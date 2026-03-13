-- name: missioncontrol__upsert_timeline_entry :one
INSERT INTO mission_control_timeline_entries (
    project_id,
    entity_id,
    source_kind,
    entry_external_key,
    command_id,
    summary,
    body_markdown,
    payload,
    occurred_at,
    provider_url,
    is_read_only
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, NOW()), $10, $11)
ON CONFLICT (project_id, source_kind, entry_external_key) DO UPDATE
SET
    entity_id = EXCLUDED.entity_id,
    command_id = EXCLUDED.command_id,
    summary = EXCLUDED.summary,
    body_markdown = EXCLUDED.body_markdown,
    payload = EXCLUDED.payload,
    occurred_at = EXCLUDED.occurred_at,
    provider_url = EXCLUDED.provider_url,
    is_read_only = EXCLUDED.is_read_only
RETURNING
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
    created_at;
