package missioncontrol

import (
	"strings"
	"testing"
)

func TestUpdateEntityProjectionQueryGuardsProjectionVersion(t *testing.T) {
	t.Parallel()

	if !strings.Contains(queryUpdateEntityProjection, "projection_version = mission_control_entities.projection_version + 1") {
		t.Fatal("update_entity_projection query must increment projection_version")
	}
	if !strings.Contains(queryUpdateEntityProjection, "AND projection_version = $14") {
		t.Fatal("update_entity_projection query must enforce expected projection_version")
	}
}

func TestMissionControlReadQueriesStayProjectScoped(t *testing.T) {
	t.Parallel()

	if !strings.Contains(queryGetCommandByID, "WHERE project_id = $1") {
		t.Fatal("get_command_by_id query must scope lookups by project_id")
	}
	if !strings.Contains(queryListTimelineEntries, "WHERE project_id = $1") {
		t.Fatal("list_timeline_entries query must scope lookups by project_id")
	}
}

func TestUpdateCommandStatusQueryUsesPatchSemantics(t *testing.T) {
	t.Parallel()

	required := []string{
		"failure_reason = CASE WHEN $4::boolean THEN $5::text ELSE failure_reason END",
		"approval_request_id = CASE WHEN $6::boolean THEN $7::uuid ELSE approval_request_id END",
		"approval_state = CASE WHEN $8::boolean THEN $9::text ELSE approval_state END",
		"result_payload = CASE WHEN $14::boolean THEN $15::jsonb ELSE result_payload END",
		"provider_delivery_ids = CASE WHEN $16::boolean THEN $17::jsonb ELSE provider_delivery_ids END",
		"WHERE project_id = $1",
	}
	for _, item := range required {
		if !strings.Contains(queryUpdateCommandStatus, item) {
			t.Fatalf("update_command_status query must contain %q", item)
		}
	}
}

func TestWarmupSummaryQueryCountsAllFoundationTables(t *testing.T) {
	t.Parallel()

	required := []string{
		"mission_control_entities",
		"mission_control_relations",
		"mission_control_timeline_entries",
		"mission_control_commands",
	}
	for _, item := range required {
		if !strings.Contains(queryGetWarmupSummary, item) {
			t.Fatalf("warmup summary query must reference %s", item)
		}
	}
}
