-- name: missioncontrol__get_warmup_summary :one
SELECT
    $1::text AS project_id,
    COALESCE((SELECT COUNT(*) FROM mission_control_entities WHERE project_id = $1::uuid), 0) AS entity_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_relations WHERE project_id = $1::uuid), 0) AS relation_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_timeline_entries WHERE project_id = $1::uuid), 0) AS timeline_entry_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_commands WHERE project_id = $1::uuid), 0) AS command_count,
    COALESCE((SELECT MAX(projection_version) FROM mission_control_entities WHERE project_id = $1::uuid), 0) AS max_projection_version,
    COALESCE((SELECT COUNT(*) FROM mission_control_entities WHERE project_id = $1::uuid AND entity_kind = 'run'), 0) AS run_entity_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_entities WHERE project_id = $1::uuid AND entity_kind = 'agent'), 0) AS legacy_agent_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_continuity_gaps WHERE project_id = $1::uuid), 0) AS continuity_gap_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_continuity_gaps WHERE project_id = $1::uuid AND status = 'open'), 0) AS open_continuity_gap_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_continuity_gaps WHERE project_id = $1::uuid AND status = 'open' AND severity = 'blocking'), 0) AS blocking_gap_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_continuity_gaps WHERE project_id = $1::uuid AND status = 'open' AND gap_kind = 'missing_pull_request'), 0) AS missing_pull_request_gap_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_continuity_gaps WHERE project_id = $1::uuid AND status = 'open' AND gap_kind = 'missing_follow_up_issue'), 0) AS missing_follow_up_issue_gap_count,
    COALESCE((SELECT COUNT(*) FROM mission_control_workspace_watermarks WHERE project_id = $1::uuid), 0) AS watermark_count;
