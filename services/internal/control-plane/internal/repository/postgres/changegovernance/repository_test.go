package changegovernance

import (
	"strings"
	"testing"
)

func TestUpdatePackageSummaryQueryUsesProjectionVersionGuard(t *testing.T) {
	t.Parallel()

	required := []string{
		"active_projection_version = change_governance_packages.active_projection_version + 1",
		"AND active_projection_version = $13",
	}
	for _, item := range required {
		if !strings.Contains(queryUpdatePackageSummary, item) {
			t.Fatalf("update_package_summary query must contain %q", item)
		}
	}
}

func TestProjectionQueriesKeepCurrentSnapshotSemantics(t *testing.T) {
	t.Parallel()

	if !strings.Contains(queryListCurrentProjectionSnapshots, "WHERE package_id = $1::uuid") {
		t.Fatal("list_current_projection_snapshots query must filter by package_id")
	}
	if !strings.Contains(queryDeactivateCurrentProjections, "is_current = false") {
		t.Fatal("deactivate_current_projections query must mark previous rows not current")
	}
	if !strings.Contains(queryInsertProjectionSnapshot, "is_current") {
		t.Fatal("insert_projection_snapshot query must persist is_current marker")
	}
}
