package mcp

import (
	"strings"
	"testing"
)

func TestNormalizeRunStatusReportText(t *testing.T) {
	t.Parallel()

	t.Run("trim and keep valid text", func(t *testing.T) {
		t.Parallel()

		got, err := normalizeRunStatusReportText("  анализирую код  ")
		if err != nil {
			t.Fatalf("normalizeRunStatusReportText returned error: %v", err)
		}
		if got != "анализирую код" {
			t.Fatalf("normalized status = %q, want %q", got, "анализирую код")
		}
	})

	t.Run("empty status rejected", func(t *testing.T) {
		t.Parallel()

		if _, err := normalizeRunStatusReportText("   "); err == nil {
			t.Fatal("expected error for empty status")
		}
	})

	t.Run("status longer than limit rejected", func(t *testing.T) {
		t.Parallel()

		tooLong := strings.Repeat("а", maxRunStatusReportChars+1)
		if _, err := normalizeRunStatusReportText(tooLong); err == nil {
			t.Fatal("expected error for status longer than 100 characters")
		}
	})
}
