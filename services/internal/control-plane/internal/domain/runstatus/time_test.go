package runstatus

import (
	"testing"
	"time"
)

func TestFormatRecentStatusTimeLabel_SameDay(t *testing.T) {
	t.Parallel()

	got := formatRecentStatusTimeLabel("2026-03-12T10:06:00Z", localeRU, time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC))
	if got != "10:06" {
		t.Fatalf("expected same-day label %q, got %q", "10:06", got)
	}
}

func TestFormatRecentStatusTimeLabel_OlderDay(t *testing.T) {
	t.Parallel()

	got := formatRecentStatusTimeLabel("2026-03-11T16:00:00Z", localeRU, time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC))
	if got != "11 мар 16:00" {
		t.Fatalf("expected older label %q, got %q", "11 мар 16:00", got)
	}
}
