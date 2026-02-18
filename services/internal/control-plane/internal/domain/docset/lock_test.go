package docset

import (
	"testing"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestLock_MarshalParseRoundtrip(t *testing.T) {
	in := NewLock(
		"docset-1",
		"main",
		"en",
		[]string{"base", "templates"},
		[]valuetypes.DocsetLockFile{{
			Path:       "docs/a.md",
			SHA256:     "sha_a",
			SourcePath: "src/a_en.md",
		}},
	)

	blob, err := MarshalLock(in)
	if err != nil {
		t.Fatalf("MarshalLock: %v", err)
	}

	out, err := ParseLock(blob)
	if err != nil {
		t.Fatalf("ParseLock: %v", err)
	}

	if out.LockVersion != 1 {
		t.Fatalf("LockVersion=%d, want 1", out.LockVersion)
	}
	if out.Docset.ID != in.Docset.ID || out.Docset.Ref != in.Docset.Ref || out.Docset.Locale != in.Docset.Locale {
		t.Fatalf("Docset mismatch: got=%+v want=%+v", out.Docset, in.Docset)
	}
	if len(out.Docset.SelectedGroups) != len(in.Docset.SelectedGroups) {
		t.Fatalf("SelectedGroups len=%d, want %d", len(out.Docset.SelectedGroups), len(in.Docset.SelectedGroups))
	}
	for i := range in.Docset.SelectedGroups {
		if out.Docset.SelectedGroups[i] != in.Docset.SelectedGroups[i] {
			t.Fatalf("SelectedGroups[%d]=%q, want %q", i, out.Docset.SelectedGroups[i], in.Docset.SelectedGroups[i])
		}
	}
	if len(out.Files) != 1 {
		t.Fatalf("Files len=%d, want 1", len(out.Files))
	}
	if out.Files[0] != in.Files[0] {
		t.Fatalf("File mismatch: got=%+v want=%+v", out.Files[0], in.Files[0])
	}
}

func TestParseLock_UnsupportedVersion(t *testing.T) {
	_, err := ParseLock([]byte(`{"lock_version":2,"docset":{"id":"x","ref":"y","locale":"en","selected_groups":[]},"files":[]}`))
	if err == nil {
		t.Fatalf("expected error")
	}
}
