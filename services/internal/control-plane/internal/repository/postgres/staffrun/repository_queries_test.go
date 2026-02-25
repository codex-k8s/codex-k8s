package staffrun

import (
	"strings"
	"testing"
)

func TestRunListQueriesIncludeReviewerTriggerLabel(t *testing.T) {
	t.Parallel()

	queries := map[string]string{
		"list_all":      queryListAll,
		"list_for_user": queryListForUser,
	}

	for name, query := range queries {
		if !strings.Contains(query, "ILIKE 'run:%'") {
			t.Fatalf("%s query must keep run trigger filter", name)
		}
		if !strings.Contains(query, "ILIKE 'need:reviewer'") {
			t.Fatalf("%s query must include reviewer trigger filter", name)
		}
	}
}
