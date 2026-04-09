package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("KODEX_DB_HOST", "postgres")
	t.Setenv("KODEX_DB_NAME", "kodex")
	t.Setenv("KODEX_DB_USER", "kodex")
	t.Setenv("KODEX_DB_PASSWORD", "secret")
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "kodex-control-plane:9090")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.PollInterval != "5s" {
		t.Fatalf("expected default poll interval 5s, got %s", cfg.PollInterval)
	}
	if cfg.Mode != "service" {
		t.Fatalf("expected default worker mode service, got %s", cfg.Mode)
	}
	if cfg.K8sNamespace != "kodex-prod" {
		t.Fatalf("expected default namespace kodex-prod, got %s", cfg.K8sNamespace)
	}
	if cfg.JobImage != "busybox:1.36" {
		t.Fatalf("expected default job image busybox:1.36, got %s", cfg.JobImage)
	}
	if cfg.RunNamespacePrefix != "codex-issue" {
		t.Fatalf("expected default run namespace prefix codex-issue, got %s", cfg.RunNamespacePrefix)
	}
	if !cfg.RunNamespaceCleanup {
		t.Fatal("expected namespace cleanup to be enabled by default")
	}
	if cfg.ServicesConfigPath != "services.yaml" {
		t.Fatalf("expected default services config path services.yaml, got %s", cfg.ServicesConfigPath)
	}
	if cfg.ServicesConfigEnv != "production" {
		t.Fatalf("expected default services config env production, got %s", cfg.ServicesConfigEnv)
	}
	if cfg.NamespaceLeaseSweepLimit != 200 {
		t.Fatalf("expected default namespace lease sweep limit 200, got %d", cfg.NamespaceLeaseSweepLimit)
	}
	if cfg.ControlPlaneMCPBaseURL != "" {
		t.Fatalf("expected empty control-plane mcp url before runtime fallback, got %s", cfg.ControlPlaneMCPBaseURL)
	}
}

func TestLoadConfigMissingDB(t *testing.T) {
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected error for missing required db environment values")
	}
}
