-- name: missioncontrol__update_entity_projection :one
UPDATE mission_control_entities
SET
    provider_url = $4,
    title = $5,
    active_state = $6,
    sync_status = $7,
    continuity_status = $8,
    coverage_class = $9,
    card_payload = $10,
    detail_payload = $11,
    last_timeline_at = $12,
    provider_updated_at = $13,
    projected_at = COALESCE($14, NOW()),
    stale_after = $15,
    projection_version = mission_control_entities.projection_version + 1,
    updated_at = NOW()
WHERE project_id = $1
  AND entity_kind = $2
  AND entity_external_key = $3
  AND projection_version = $16
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
    continuity_status,
    coverage_class,
    projection_version,
    card_payload AS card_payload_json,
    detail_payload AS detail_payload_json,
    last_timeline_at,
    provider_updated_at,
    projected_at,
    stale_after,
    created_at,
    updated_at;
