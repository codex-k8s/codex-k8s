-- name: missioncontrol__get_entity_by_id :one
SELECT
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
    updated_at
FROM mission_control_entities
WHERE project_id = $1
  AND id = $2;
