package runner

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadProjectDocsForPrompt_FiltersByRoleAndKeepsOrder(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: kodex.works/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
    ai:
      namespaceTemplate: "{{ .Project }}-dev-1"
  projectDocs:
    - repository: policy-docs
      path: docs/common.md
    - path: docs/dev.md
      roles: [dev]
    - path: docs/ops.md
      roles: [sre]
`
	if err := os.WriteFile(filepath.Join(repoDir, "services.yaml"), []byte(servicesYAML), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}

	docs, total, trimmed := loadProjectDocsForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("trimmed=%v, want false", trimmed)
	}
	if total != 2 {
		t.Fatalf("total=%d, want 2", total)
	}
	if len(docs) != 2 {
		t.Fatalf("len(docs)=%d, want 2", len(docs))
	}
	if docs[0].Path != "docs/common.md" || docs[1].Path != "docs/dev.md" {
		t.Fatalf("unexpected docs list: %+v", docs)
	}
	if docs[0].Repository != "policy-docs" {
		t.Fatalf("docs[0].repository=%q, want policy-docs", docs[0].Repository)
	}
}

func TestLoadProjectDocsForPrompt_DedupWithRepositoryPriority(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: kodex.works/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
    ai:
      namespaceTemplate: "{{ .Project }}-dev-1"
  projectDocs:
    - repository: service-a
      path: docs/architecture.md
      description: service copy
    - repository: policy-docs
      path: docs/architecture.md
      description: policy copy
`
	if err := os.WriteFile(filepath.Join(repoDir, "services.yaml"), []byte(servicesYAML), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}

	docs, total, trimmed := loadProjectDocsForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("trimmed=%v, want false", trimmed)
	}
	if total != 1 || len(docs) != 1 {
		t.Fatalf("expected one deduped doc, total=%d len=%d", total, len(docs))
	}
	if docs[0].Repository != "policy-docs" {
		t.Fatalf("repository=%q, want policy-docs", docs[0].Repository)
	}
	if docs[0].Description != "policy copy" {
		t.Fatalf("description=%q, want policy copy", docs[0].Description)
	}
}

func TestResolvePromptDocsEnv(t *testing.T) {
	t.Parallel()

	if got, want := resolvePromptDocsEnv("dev", runtimeModeFullEnv, ""), "ai"; got != want {
		t.Fatalf("resolvePromptDocsEnv(dev, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("qa_revise", runtimeModeFullEnv, ""), "ai"; got != want {
		t.Fatalf("resolvePromptDocsEnv(qa_revise, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("release", runtimeModeFullEnv, ""), "ai"; got != want {
		t.Fatalf("resolvePromptDocsEnv(release, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("ops_revise", runtimeModeFullEnv, "production"), "production"; got != want {
		t.Fatalf("resolvePromptDocsEnv(ops_revise, full-env, production)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("plan", runtimeModeFullEnv, ""), "production"; got != want {
		t.Fatalf("resolvePromptDocsEnv(plan, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("dev", runtimeModeCodeOnly, ""), "production"; got != want {
		t.Fatalf("resolvePromptDocsEnv(dev, code-only)=%q, want %q", got, want)
	}
}

func TestLoadRoleDocTemplatesForPrompt_FiltersByRoleAndDedupsByPath(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: kodex.works/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
    ai:
      namespaceTemplate: "{{ .Project }}-dev-1"
  roleDocTemplates:
    DEV:
      - repository: service-api
        path: docs/templates/prd.md
        description: service copy
      - repository: policy-docs
        path: docs/templates/prd.md
        description: policy copy
      - path: docs/templates/data_model.md
    qa:
      - path: docs/templates/test_plan.md
`
	if err := os.WriteFile(filepath.Join(repoDir, "services.yaml"), []byte(servicesYAML), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}

	templates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("trimmed=%v, want false", trimmed)
	}
	if total != 2 {
		t.Fatalf("total=%d, want 2", total)
	}
	if len(templates) != 2 {
		t.Fatalf("len(templates)=%d, want 2", len(templates))
	}
	if templates[0].Repository != "policy-docs" {
		t.Fatalf("templates[0].repository=%q, want policy-docs", templates[0].Repository)
	}
	if templates[0].Path != "docs/templates/prd.md" {
		t.Fatalf("templates[0].path=%q, want docs/templates/prd.md", templates[0].Path)
	}
	if templates[0].TemplateName != "prd.md" {
		t.Fatalf("templates[0].templateName=%q, want prd.md", templates[0].TemplateName)
	}
	if templates[1].Path != "docs/templates/data_model.md" {
		t.Fatalf("templates[1].path=%q, want docs/templates/data_model.md", templates[1].Path)
	}
}

func TestLoadRoleDocTemplatesForPrompt_NoRoleMatch(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: kodex.works/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
    ai:
      namespaceTemplate: "{{ .Project }}-dev-1"
  roleDocTemplates:
    qa:
      - path: docs/templates/test_plan.md
`
	if err := os.WriteFile(filepath.Join(repoDir, "services.yaml"), []byte(servicesYAML), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}

	templates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("trimmed=%v, want false", trimmed)
	}
	if total != 0 {
		t.Fatalf("total=%d, want 0", total)
	}
	if len(templates) != 0 {
		t.Fatalf("len(templates)=%d, want 0", len(templates))
	}
}

func TestRepositoryServicesYAML_RoleDocTemplatesMatchGovernanceMatrix(t *testing.T) {
	t.Parallel()

	repoDir := findRepositoryRootForPromptTests(t)
	expected := map[string][]string{
		"pm":  {"problem.md", "scope_mvp.md", "constraints.md", "brief.md", "project_charter.md", "success_metrics.md", "prd.md", "nfr.md", "user_story.md"},
		"em":  {"delivery_plan.md", "epic.md", "definition_of_done.md", "release_plan.md", "release_notes.md", "rollback_plan.md"},
		"sa":  {"c4_context.md", "c4_container.md", "adr.md", "alternatives.md", "api_contract.md", "data_model.md", "design_doc.md", "migrations_policy.md"},
		"dev": {"user_story.md", "definition_of_done.md"},
		"qa":  {"test_strategy.md", "test_plan.md", "test_matrix.md", "regression_checklist.md", "postdeploy_review.md"},
		"sre": {"runbook.md", "monitoring.md", "alerts.md", "slo.md", "incident_playbook.md", "incident_postmortem.md"},
		"km":  {"issue_map.md", "delivery_plan.md", "roadmap.md", "docset_issue.md", "docset_pr.md"},
	}

	for role, want := range expected {
		templates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, role, role, runtimeModeFullEnv, "")
		if trimmed {
			t.Fatalf("role %q templates trimmed unexpectedly", role)
		}
		if total != len(want) {
			t.Fatalf("role %q total=%d, want %d", role, total, len(want))
		}

		got := make([]string, 0, len(templates))
		for _, template := range templates {
			got = append(got, template.TemplateName)
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("role %q templates=%v, want %v", role, got, want)
		}
	}

	reviewerTemplates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, "reviewer", "reviewer", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("reviewer templates trimmed unexpectedly")
	}
	if total != 0 || len(reviewerTemplates) != 0 {
		t.Fatalf("reviewer must not expose role templates, total=%d len=%d", total, len(reviewerTemplates))
	}
}

func TestRepositoryServicesYAML_ProjectDocsCoverDeliveryAndOpsRoles(t *testing.T) {
	t.Parallel()

	repoDir := findRepositoryRootForPromptTests(t)
	docs, total, trimmed := loadProjectDocsForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv, "")
	if trimmed {
		t.Fatalf("dev project docs trimmed unexpectedly")
	}
	if total == 0 || len(docs) == 0 {
		t.Fatalf("dev project docs must not be empty")
	}

	assertPromptDocVisible(t, repoDir, "dev", "dev", "docs/delivery")
	assertPromptDocVisible(t, repoDir, "qa", "qa", "docs/delivery")
	assertPromptDocVisible(t, repoDir, "qa", "qa", "docs/ops")
	assertPromptDocVisible(t, repoDir, "sre", "ops", "docs/delivery")
	assertPromptDocVisible(t, repoDir, "sre", "ops", "docs/ops")
}

func assertPromptDocVisible(t *testing.T, repoDir string, role string, triggerKind string, wantPath string) {
	t.Helper()

	docs, _, _ := loadProjectDocsForPrompt(repoDir, role, triggerKind, runtimeModeFullEnv, "")
	for _, doc := range docs {
		if doc.Path == wantPath {
			return
		}
	}
	t.Fatalf("role %q trigger %q must see project doc %q", role, triggerKind, wantPath)
}

func findRepositoryRootForPromptTests(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "services.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repository root with services.yaml not found from %q", dir)
		}
		dir = parent
	}
}
