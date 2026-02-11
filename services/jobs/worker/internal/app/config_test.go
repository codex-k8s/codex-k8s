package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("CODEXK8S_DB_HOST", "postgres")
	t.Setenv("CODEXK8S_DB_NAME", "codex_k8s")
	t.Setenv("CODEXK8S_DB_USER", "codex_k8s")
	t.Setenv("CODEXK8S_DB_PASSWORD", "secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.PollInterval != "5s" {
		t.Fatalf("expected default poll interval 5s, got %s", cfg.PollInterval)
	}
	if cfg.K8sNamespace != "codex-k8s-ai-staging" {
		t.Fatalf("expected default namespace codex-k8s-ai-staging, got %s", cfg.K8sNamespace)
	}
	if cfg.JobImage != "busybox:1.36" {
		t.Fatalf("expected default job image busybox:1.36, got %s", cfg.JobImage)
	}
	if cfg.RunNamespacePrefix != "codex-issue" {
		t.Fatalf("expected default run namespace prefix codex-issue, got %s", cfg.RunNamespacePrefix)
	}
	if !cfg.RunNamespaceCleanup {
		t.Fatal("expected run namespace cleanup to be enabled by default")
	}
}

func TestLoadConfigMissingDB(t *testing.T) {
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected error for missing required db environment values")
	}
}
