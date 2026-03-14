-- name: missioncontrol__upsert_entity :one
INSERT INTO mission_control_entities (
    project_id,
    entity_kind,
    entity_external_key,
    provider_kind,
    provider_url,
    title,
    active_state,
    projection_version,
    card_payload,
    detail_payload,
    last_timeline_at,
    provider_updated_at,
    projected_at,
    stale_after,
    sync_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, GREATEST($8::bigint, 1::bigint), $9, $10, $11, $12, COALESCE($13, NOW()), $14, $15)
ON CONFLICT (project_id, entity_kind, entity_external_key) DO UPDATE
SET
    provider_kind = EXCLUDED.provider_kind,
    provider_url = EXCLUDED.provider_url,
    title = EXCLUDED.title,
    active_state = EXCLUDED.active_state,
    sync_status = EXCLUDED.sync_status,
    projection_version = GREATEST(mission_control_entities.projection_version, EXCLUDED.projection_version),
    card_payload = EXCLUDED.card_payload,
    detail_payload = EXCLUDED.detail_payload,
    last_timeline_at = EXCLUDED.last_timeline_at,
    provider_updated_at = EXCLUDED.provider_updated_at,
    projected_at = EXCLUDED.projected_at,
    stale_after = EXCLUDED.stale_after,
    updated_at = NOW()
RETURNING
    id,
    project_id::text AS project_id,
    entity_kind,
    entity_external_key,
    provider_kind,
    provider_url,
    title,
    active_state,
    sync_status,
    projection_version,
    card_payload AS card_payload_json,
    detail_payload AS detail_payload_json,
    last_timeline_at,
    provider_updated_at,
    projected_at,
    stale_after,
    created_at,
    updated_at;
