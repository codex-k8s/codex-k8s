-- name: missioncontrol__delete_relations_for_source :exec
DELETE FROM mission_control_relations
WHERE project_id = $1
  AND source_entity_id = $2;
