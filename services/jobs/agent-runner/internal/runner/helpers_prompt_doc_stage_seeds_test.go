package runner

import (
	"strings"
	"testing"
)

func TestDocumentationStageSeeds_UseRoleAwareTemplateRefs(t *testing.T) {
	t.Parallel()

	documentationStageSeeds := []string{
		"intake-work.md",
		"intake-revise.md",
		"vision-work.md",
		"vision-revise.md",
		"prd-work.md",
		"prd-revise.md",
		"arch-work.md",
		"arch-revise.md",
		"design-work.md",
		"design-revise.md",
		"plan-work.md",
		"plan-revise.md",
		"doc-audit-work.md",
		"qa-work.md",
		"release-work.md",
		"postdeploy-work.md",
		"ops-work.md",
		"rethink-work.md",
	}

	forbiddenFragments := []string{
		"docs/templates/",
		"docs/product/",
		"docs/delivery/",
		"docs/architecture/",
		"`issue_map`",
		"`requirements_traceability`",
	}

	for _, seedFile := range documentationStageSeeds {
		seedFile := seedFile
		t.Run(seedFile, func(t *testing.T) {
			t.Parallel()

			body, err := promptSeedsFS.ReadFile(promptSeedsDirRelativePath + "/" + seedFile)
			if err != nil {
				t.Fatalf("read seed %s: %v", seedFile, err)
			}

			content := string(body)
			if !strings.Contains(content, "Шаблоны артефактов по роли") {
				t.Fatalf("seed %s must reference role-aware template block", seedFile)
			}

			for _, forbidden := range forbiddenFragments {
				if strings.Contains(content, forbidden) {
					t.Fatalf("seed %s must not contain hardcoded fragment %q", seedFile, forbidden)
				}
			}
		})
	}
}

func TestNonDocStageSeeds_DoNotRequireRoleTemplateBlock(t *testing.T) {
	t.Parallel()

	nonDocSeeds := []string{
		"dev-work.md",
		"dev-revise.md",
		"ai-repair-work.md",
		"self-improve-work.md",
	}

	for _, seedFile := range nonDocSeeds {
		seedFile := seedFile
		t.Run(seedFile, func(t *testing.T) {
			t.Parallel()

			body, err := promptSeedsFS.ReadFile(promptSeedsDirRelativePath + "/" + seedFile)
			if err != nil {
				t.Fatalf("read seed %s: %v", seedFile, err)
			}

			if strings.Contains(string(body), "Шаблоны артефактов по роли") {
				t.Fatalf("seed %s must not require doc template block", seedFile)
			}
		})
	}
}
