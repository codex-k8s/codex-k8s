package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProjectDocsForPrompt_FiltersByRoleAndKeepsOrder(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: codex-k8s.dev/v1alpha1
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

	docs, total, trimmed := loadProjectDocsForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv)
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
apiVersion: codex-k8s.dev/v1alpha1
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

	docs, total, trimmed := loadProjectDocsForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv)
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

	if got, want := resolvePromptDocsEnv("dev", runtimeModeFullEnv), "ai"; got != want {
		t.Fatalf("resolvePromptDocsEnv(dev, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("plan", runtimeModeFullEnv), "production"; got != want {
		t.Fatalf("resolvePromptDocsEnv(plan, full-env)=%q, want %q", got, want)
	}
	if got, want := resolvePromptDocsEnv("dev", runtimeModeCodeOnly), "production"; got != want {
		t.Fatalf("resolvePromptDocsEnv(dev, code-only)=%q, want %q", got, want)
	}
}

func TestLoadRoleDocTemplatesForPrompt_FiltersByRoleAndDedupsByPath(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	servicesYAML := `
apiVersion: codex-k8s.dev/v1alpha1
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

	templates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv)
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
apiVersion: codex-k8s.dev/v1alpha1
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

	templates, total, trimmed := loadRoleDocTemplatesForPrompt(repoDir, "dev", "dev", runtimeModeFullEnv)
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
