-- name: missioncontrol__get_warmup_summary :one
SELECT
    $1::text AS project_id,
    COALESCE((SELECT COUNT(*) FROM mission_control_entities WHERE project_id = $1), 0) AS entity_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_relations WHERE project_id = $1), 0) AS relation_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_timeline_entries WHERE project_id = $1), 0) AS timeline_entry_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_commands WHERE project_id = $1), 0) AS command_count,
    COALESCE((SELECT MAX(projection_version) FROM mission_control_entities WHERE project_id = $1), 0) AS max_projection_version;
