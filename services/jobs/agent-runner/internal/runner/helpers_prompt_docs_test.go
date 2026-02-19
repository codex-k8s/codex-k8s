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
    - path: docs/common.md
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
