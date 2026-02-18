package docset

import (
	"testing"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestBuildImportPlan_RejectsTraversal(t *testing.T) {
	m := entitytypes.DocsetManifest{
		ManifestVersion: 1,
		Docset:          entitytypes.DocsetManifestDocset{ID: "docset-1"},
		Groups: []entitytypes.DocsetManifestGroup{{
			ID:              "core",
			DefaultSelected: true,
			ItemIDs:         []string{"file:a"},
		}},
		Items: []entitytypes.DocsetManifestItem{{
			ID:          "file:a",
			ImportPath:  "../escape.md",
			SourcePaths: entitytypes.DocsetLocalizedText{EN: "docs/a_en.md"},
			SHA256:      entitytypes.DocsetLocalizedText{EN: "x"},
		}},
	}
	if _, _, err := BuildImportPlan(m, "en", nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildImportPlan_DefaultSelected(t *testing.T) {
	m := entitytypes.DocsetManifest{
		ManifestVersion: 1,
		Docset:          entitytypes.DocsetManifestDocset{ID: "docset-1"},
		Groups: []entitytypes.DocsetManifestGroup{
			{
				ID:              "core",
				DefaultSelected: true,
				ItemIDs:         []string{"file:a"},
			},
			{
				ID:              "examples",
				DefaultSelected: false,
				ItemIDs:         []string{"file:b"},
			},
		},
		Items: []entitytypes.DocsetManifestItem{
			{
				ID:          "file:a",
				ImportPath:  "docs/a.md",
				SourcePaths: entitytypes.DocsetLocalizedText{RU: "docs/a_ru.md", EN: "docs/a_en.md"},
				SHA256:      entitytypes.DocsetLocalizedText{RU: "sha_ru", EN: "sha_en"},
			},
			{
				ID:          "file:b",
				ImportPath:  "docs/b.md",
				SourcePaths: entitytypes.DocsetLocalizedText{EN: "docs/b_en.md"},
				SHA256:      entitytypes.DocsetLocalizedText{EN: "sha_b"},
			},
		},
	}

	plan, groups, err := BuildImportPlan(m, "ru", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 || groups[0] != "core" {
		t.Fatalf("unexpected groups: %#v", groups)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("unexpected files: %d", len(plan.Files))
	}
	if plan.Files[0].SrcPath != "docs/a_ru.md" || plan.Files[0].DstPath != "docs/a.md" {
		t.Fatalf("unexpected plan file: %#v", plan.Files[0])
	}
}

func TestBuildImportPlan_DefaultSelected_ExcludesExamples(t *testing.T) {
	m := entitytypes.DocsetManifest{
		ManifestVersion: 1,
		Docset:          entitytypes.DocsetManifestDocset{ID: "docset-1"},
		Groups: []entitytypes.DocsetManifestGroup{
			{
				ID:              "examples",
				DefaultSelected: true,
				ItemIDs:         []string{"file:examples"},
			},
			{
				ID:              "core",
				DefaultSelected: true,
				ItemIDs:         []string{"file:a"},
			},
		},
		Items: []entitytypes.DocsetManifestItem{
			{
				ID:          "file:examples",
				ImportPath:  "docs/examples.md",
				SourcePaths: entitytypes.DocsetLocalizedText{EN: "docs/examples_en.md"},
				SHA256:      entitytypes.DocsetLocalizedText{EN: "sha_examples"},
			},
			{
				ID:          "file:a",
				ImportPath:  "docs/a.md",
				SourcePaths: entitytypes.DocsetLocalizedText{EN: "docs/a_en.md"},
				SHA256:      entitytypes.DocsetLocalizedText{EN: "sha_a"},
			},
		},
	}

	plan, groups, err := BuildImportPlan(m, "en", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 || groups[0] != "core" {
		t.Fatalf("unexpected groups: %#v", groups)
	}
	if len(plan.Files) != 1 || plan.Files[0].DstPath != "docs/a.md" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}
