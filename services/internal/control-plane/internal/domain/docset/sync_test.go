package docset

import "testing"

func TestBuildSafeSyncPlan_UpdatesAndDrift(t *testing.T) {
	lock := NewLock(
		"docset-1",
		"oldref",
		"en",
		[]string{"base"},
		[]LockFile{
			{Path: "docs/a.md", SHA256: "sha_old_a", SourcePath: "src/a_en.md"},
			{Path: "docs/b.md", SHA256: "sha_old_b", SourcePath: "src/b_en.md"},
			{Path: "docs/c.md", SHA256: "sha_old_c", SourcePath: "src/c_en.md"},
			{Path: "docs/d.md", SHA256: "sha_old_d", SourcePath: "src/d_en.md"},
		},
	)

	manifest := Manifest{
		ManifestVersion: 1,
		ID:              "docset-1",
		Groups: []ManifestGroup{{
			ID: "base",
			Items: []ManifestGroupItem{
				{
					ImportPath:  "docs/a.md",
					SourcePaths: LocalizedText{EN: "src/a2_en.md"},
					SHA256:      LocalizedText{EN: "sha_new_a"},
				},
				{
					ImportPath:  "docs/b.md",
					SourcePaths: LocalizedText{EN: "src/b2_en.md"},
					SHA256:      LocalizedText{EN: "sha_old_b"},
				},
				{
					ImportPath:  "docs/d.md",
					SourcePaths: LocalizedText{EN: "src/d2_en.md"},
					SHA256:      LocalizedText{EN: ""},
				},
			},
		}},
	}

	current := map[string]string{
		// a has no local modifications, should update.
		"docs/a.md": "sha_old_a",
		// b has local modifications, should drift.
		"docs/b.md": "sha_local_b",
		// c is missing in current set, should drift.
		// d is ok locally but missing sha in manifest, should drift.
		"docs/d.md": "sha_old_d",
	}

	plan, err := BuildSafeSyncPlan(lock, manifest, "en", current)
	if err != nil {
		t.Fatalf("BuildSafeSyncPlan: %v", err)
	}

	if len(plan.Updates) != 1 {
		t.Fatalf("Updates len=%d, want 1", len(plan.Updates))
	}
	if got := plan.Updates[0]; got.DstPath != "docs/a.md" || got.SrcPath != "src/a2_en.md" || got.ExpectedSHA256 != "sha_new_a" {
		t.Fatalf("unexpected update: %+v", got)
	}

	if len(plan.Drift) != 3 {
		t.Fatalf("Drift len=%d, want 3", len(plan.Drift))
	}

	byPath := make(map[string]SyncDecision, len(plan.Drift))
	for _, d := range plan.Drift {
		byPath[d.Path] = d
	}

	if d, ok := byPath["docs/b.md"]; !ok || d.Reason != "local modifications detected" {
		t.Fatalf("drift docs/b.md missing or unexpected: %+v ok=%v", d, ok)
	}
	if d, ok := byPath["docs/c.md"]; !ok || d.Reason != "file missing" {
		t.Fatalf("drift docs/c.md missing or unexpected: %+v ok=%v", d, ok)
	}
	if d, ok := byPath["docs/d.md"]; !ok || d.Reason != "manifest missing sha256" {
		t.Fatalf("drift docs/d.md missing or unexpected: %+v ok=%v", d, ok)
	}
}

func TestUpdateLockForSync(t *testing.T) {
	lock := NewLock(
		"docset-1",
		"oldref",
		"en",
		[]string{"base"},
		[]LockFile{
			{Path: "docs/a.md", SHA256: "sha_old_a", SourcePath: "src/a_en.md"},
			{Path: "docs/b.md", SHA256: "sha_old_b", SourcePath: "src/b_en.md"},
		},
	)

	next, err := UpdateLockForSync(lock, "newref", []LockFile{
		{Path: "docs/a.md", SHA256: "sha_new_a", SourcePath: "src/a2_en.md"},
	})
	if err != nil {
		t.Fatalf("UpdateLockForSync: %v", err)
	}

	if next.Docset.Ref != "newref" {
		t.Fatalf("Ref=%q, want %q", next.Docset.Ref, "newref")
	}
	if next.Files[0].SHA256 != "sha_new_a" || next.Files[0].SourcePath != "src/a2_en.md" {
		t.Fatalf("file a not updated: %+v", next.Files[0])
	}
	if next.Files[1] != lock.Files[1] {
		t.Fatalf("file b changed: got=%+v want=%+v", next.Files[1], lock.Files[1])
	}

	lock.LockVersion = 2
	if _, err := UpdateLockForSync(lock, "x", nil); err == nil {
		t.Fatalf("expected error")
	}
}
