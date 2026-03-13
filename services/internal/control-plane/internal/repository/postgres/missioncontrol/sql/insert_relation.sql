-- name: missioncontrol__insert_relation :exec
INSERT INTO mission_control_relations (
    project_id,
    source_entity_id,
    relation_kind,
    target_entity_id,
    source_kind
)
VALUES ($1, $2, $3, $4, $5);
