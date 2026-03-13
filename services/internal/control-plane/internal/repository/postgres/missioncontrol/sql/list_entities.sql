-- name: missioncontrol__list_entities :many
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
  AND (
      COALESCE(array_length($2::text[], 1), 0) = 0
      OR active_state = ANY($2::text[])
  )
  AND (
      COALESCE(array_length($3::text[], 1), 0) = 0
      OR sync_status = ANY($3::text[])
  )
ORDER BY last_timeline_at DESC NULLS LAST, projected_at DESC, id DESC
LIMIT $4;
