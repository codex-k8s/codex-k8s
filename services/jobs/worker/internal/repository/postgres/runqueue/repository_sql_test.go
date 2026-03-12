package runqueue

import (
	"strings"
	"testing"
)

func TestClaimRunningSQL_ReturnsLeaseColumns(t *testing.T) {
	returningStart := strings.Index(queryClaimRunning, "RETURNING")
	if returningStart == -1 {
		t.Fatal("expected RETURNING clause in claim_running.sql")
	}

	returningEnd := strings.Index(queryClaimRunning[returningStart:], "\n)\nSELECT")
	if returningEnd == -1 {
		t.Fatal("expected SELECT clause after RETURNING in claim_running.sql")
	}

	returningClause := queryClaimRunning[returningStart : returningStart+returningEnd]
	for _, column := range []string{"r.lease_owner", "r.lease_until"} {
		if !strings.Contains(returningClause, column) {
			t.Fatalf("expected RETURNING clause to include %s", column)
		}
	}

	for _, selectedColumn := range []string{"c.lease_owner", "c.lease_until"} {
		if !strings.Contains(queryClaimRunning, selectedColumn) {
			t.Fatalf("expected outer SELECT to include %s", selectedColumn)
		}
	}
}
