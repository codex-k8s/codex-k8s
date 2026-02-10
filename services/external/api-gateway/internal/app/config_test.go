package app

import "testing"

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("CODEXK8S_PUBLIC_BASE_URL", "https://staging.example.com")
	t.Setenv("CODEXK8S_BOOTSTRAP_OWNER_EMAIL", "owner@example.com")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("CODEXK8S_JWT_SIGNING_KEY", "jwt-key")
	t.Setenv("CODEXK8S_GITHUB_WEBHOOK_SECRET", "secret")
	t.Setenv("CODEXK8S_TOKEN_ENCRYPTION_KEY", "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	t.Setenv("CODEXK8S_DB_HOST", "postgres")
	t.Setenv("CODEXK8S_DB_NAME", "codex_k8s")
	t.Setenv("CODEXK8S_DB_USER", "codex")
	t.Setenv("CODEXK8S_DB_PASSWORD", "pass")

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
	if cfg.GitHubWebhookEvents == "" {
		t.Fatal("expected default GitHub webhook events to be set")
	}
	if cfg.DBPort != 5432 {
		t.Fatalf("expected default db port 5432, got %d", cfg.DBPort)
	}
	if cfg.DBSSLMode != "disable" {
		t.Fatalf("expected default sslmode disable, got %q", cfg.DBSSLMode)
	}
	if cfg.JWTTTL != "15m" {
		t.Fatalf("expected default jwt ttl 15m, got %q", cfg.JWTTTL)
	}
	if cfg.CookieSecure {
		t.Fatal("expected default cookie secure=false")
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	t.Setenv("CODEXK8S_PUBLIC_BASE_URL", "https://staging.example.com")
	t.Setenv("CODEXK8S_BOOTSTRAP_OWNER_EMAIL", "owner@example.com")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("CODEXK8S_JWT_SIGNING_KEY", "jwt-key")
	t.Setenv("CODEXK8S_DB_HOST", "postgres")
	t.Setenv("CODEXK8S_DB_NAME", "codex_k8s")
	t.Setenv("CODEXK8S_DB_USER", "codex")
	t.Setenv("CODEXK8S_DB_PASSWORD", "pass")
	t.Setenv("CODEXK8S_TOKEN_ENCRYPTION_KEY", "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	// CODEXK8S_GITHUB_WEBHOOK_SECRET intentionally unset

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing required webhook secret")
	}
}
