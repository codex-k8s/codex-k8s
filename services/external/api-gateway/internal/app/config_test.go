package app

import "testing"

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("CODEXK8S_CONTROL_PLANE_GRPC_TARGET", "codex-k8s-control-plane:9090")
	t.Setenv("CODEXK8S_PUBLIC_BASE_URL", "https://platform.codex-k8s.dev")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("CODEXK8S_JWT_SIGNING_KEY", "jwt-key")
	t.Setenv("CODEXK8S_GITHUB_WEBHOOK_SECRET", "secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("expected default http addr :8080, got %q", cfg.HTTPAddr)
	}
	if cfg.WebhookMaxBodyBytes != 1048576 {
		t.Fatalf("expected default webhook body size 1048576, got %d", cfg.WebhookMaxBodyBytes)
	}
	if cfg.JWTTTL != "15m" {
		t.Fatalf("expected default jwt ttl 15m, got %q", cfg.JWTTTL)
	}
	if cfg.CookieSecure {
		t.Fatal("expected default cookie secure=false")
	}
	if !cfg.OpenAPIValidationEnabled {
		t.Fatal("expected default openapi validation enabled=true")
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	t.Setenv("CODEXK8S_CONTROL_PLANE_GRPC_TARGET", "codex-k8s-control-plane:9090")
	t.Setenv("CODEXK8S_PUBLIC_BASE_URL", "https://platform.codex-k8s.dev")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("CODEXK8S_JWT_SIGNING_KEY", "jwt-key")
	// CODEXK8S_GITHUB_WEBHOOK_SECRET intentionally unset

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing required webhook secret")
	}
}
