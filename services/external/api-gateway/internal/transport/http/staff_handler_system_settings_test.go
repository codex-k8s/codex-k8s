package http

import (
	"testing"

	sharedsystemsettings "github.com/codex-k8s/codex-k8s/libs/go/systemsettings"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func TestFilterRealtimeSystemSettings_ExcludesHiddenVisibilities(t *testing.T) {
	t.Parallel()

	items := []models.SystemSetting{
		{Key: "visible", Visibility: sharedsystemsettings.VisibilityStaffVisible},
		{Key: "internal", Visibility: sharedsystemsettings.VisibilityInternalOnly},
		{Key: "no-ws", Visibility: sharedsystemsettings.VisibilitySecretForbiddenWS},
	}

	got := filterRealtimeSystemSettings(items)
	if len(got) != 1 {
		t.Fatalf("filterRealtimeSystemSettings returned %d items, want 1", len(got))
	}
	if got[0].Key != "visible" {
		t.Fatalf("filterRealtimeSystemSettings kept key %q, want %q", got[0].Key, "visible")
	}
}
