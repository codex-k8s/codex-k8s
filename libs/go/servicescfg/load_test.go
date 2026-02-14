package servicescfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "services.yaml")
	content := `
project: codex-k8s
templates:
  partials:
    - deploy/_tpl/*.gohtml
deploy:
  environments:
    ai-staging:
      namespace_env_var: CODEXK8S_STAGING_NAMESPACE
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, root, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Project != "codex-k8s" {
		t.Fatalf("unexpected project = %q", cfg.Project)
	}
	if root != tmpDir {
		t.Fatalf("unexpected root = %q", root)
	}
	if got := cfg.Deploy.Environments["ai-staging"].NamespaceEnvVar; got != "CODEXK8S_STAGING_NAMESPACE" {
		t.Fatalf("unexpected namespace env var = %q", got)
	}
}
