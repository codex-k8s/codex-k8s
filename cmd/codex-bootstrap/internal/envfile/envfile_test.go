package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ParsesExportAndQuotes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.env")
	content := `
# comment
SIMPLE=value
export QUOTED="value with spaces"
SINGLE='quoted'
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	values, err := Load(path)
	if err != nil {
		t.Fatalf("load env file: %v", err)
	}

	if got, want := values["SIMPLE"], "value"; got != want {
		t.Fatalf("unexpected SIMPLE value: got %q want %q", got, want)
	}
	if got, want := values["QUOTED"], "value with spaces"; got != want {
		t.Fatalf("unexpected QUOTED value: got %q want %q", got, want)
	}
	if got, want := values["SINGLE"], "quoted"; got != want {
		t.Fatalf("unexpected SINGLE value: got %q want %q", got, want)
	}
}

func TestLoad_InvalidLine(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "broken.env")
	if err := os.WriteFile(path, []byte("BROKEN_LINE\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatalf("expected parse error")
	}
}
