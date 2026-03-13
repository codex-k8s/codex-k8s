-- name: missioncontrol__list_relations_for_entity :many
SELECT
    id,
    project_id::text AS project_id,
    source_entity_id,
    relation_kind,
    target_entity_id,
    source_kind,
    created_at,
    updated_at
FROM mission_control_relations
WHERE project_id = $1
  AND (source_entity_id = $2 OR target_entity_id = $2)
ORDER BY updated_at DESC, id DESC;
