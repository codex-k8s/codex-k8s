package docset

import "testing"

func TestBuildImportPlan_RejectsTraversal(t *testing.T) {
	m := Manifest{
		ManifestVersion: 1,
		Groups: []ManifestGroup{{
			ID:              "core",
			DefaultSelected: true,
			Items: []ManifestGroupItem{{
				ImportPath:  "../escape.md",
				SourcePaths: LocalizedText{EN: "docs/a_en.md"},
				SHA256:      LocalizedText{EN: "x"},
			}},
		}},
	}
	if _, _, err := BuildImportPlan(m, "en", nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildImportPlan_DefaultSelected(t *testing.T) {
	m := Manifest{
		ManifestVersion: 1,
		Groups: []ManifestGroup{
			{
				ID:              "core",
				DefaultSelected: true,
				Items: []ManifestGroupItem{{
					ImportPath:  "docs/a.md",
					SourcePaths: LocalizedText{RU: "docs/a_ru.md", EN: "docs/a_en.md"},
					SHA256:      LocalizedText{RU: "sha_ru", EN: "sha_en"},
				}},
			},
			{
				ID:              "examples",
				DefaultSelected: false,
				Items: []ManifestGroupItem{{
					ImportPath:  "docs/b.md",
					SourcePaths: LocalizedText{EN: "docs/b_en.md"},
					SHA256:      LocalizedText{EN: "sha_b"},
				}},
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
	m := Manifest{
		ManifestVersion: 1,
		Groups: []ManifestGroup{
			{
				ID:              "examples",
				DefaultSelected: true,
				Items: []ManifestGroupItem{{
					ImportPath:  "docs/examples.md",
					SourcePaths: LocalizedText{EN: "docs/examples_en.md"},
					SHA256:      LocalizedText{EN: "sha_examples"},
				}},
			},
			{
				ID:              "core",
				DefaultSelected: true,
				Items: []ManifestGroupItem{{
					ImportPath:  "docs/a.md",
					SourcePaths: LocalizedText{EN: "docs/a_en.md"},
					SHA256:      LocalizedText{EN: "sha_a"},
				}},
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
