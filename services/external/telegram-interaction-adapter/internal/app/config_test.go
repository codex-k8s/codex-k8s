package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("KODEX_HTTP_ADDR", "")
	t.Setenv("KODEX_ENV", "")
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "kodex-control-plane:9090")
	t.Setenv("KODEX_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT", "")
	t.Setenv("KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_MODEL", "")
	t.Setenv("KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.Environment != "production" {
		t.Fatalf("Environment = %q, want production", cfg.Environment)
	}
	if cfg.ControlPlaneGRPCTarget != "kodex-control-plane:9090" {
		t.Fatalf("ControlPlaneGRPCTarget = %q", cfg.ControlPlaneGRPCTarget)
	}
	if cfg.TelegramHTTPTimeout != "10s" {
		t.Fatalf("TelegramHTTPTimeout = %q, want 10s", cfg.TelegramHTTPTimeout)
	}
	if cfg.TelegramSTTModel != "gpt-4o-mini-transcribe" {
		t.Fatalf("TelegramSTTModel = %q, want gpt-4o-mini-transcribe", cfg.TelegramSTTModel)
	}
	if cfg.TelegramSTTTimeout != "30s" {
		t.Fatalf("TelegramSTTTimeout = %q, want 30s", cfg.TelegramSTTTimeout)
	}
}

func TestTelegramWebhookSyncEnabled(t *testing.T) {
	t.Parallel()

	if !telegramWebhookSyncEnabled("production") {
		t.Fatal("production environment must own telegram webhook sync")
	}
	if telegramWebhookSyncEnabled("ai") {
		t.Fatal("ai environment must not own telegram webhook sync")
	}
	if !telegramWebhookSyncEnabled("") {
		t.Fatal("empty environment should keep production-compatible behavior")
	}
}
