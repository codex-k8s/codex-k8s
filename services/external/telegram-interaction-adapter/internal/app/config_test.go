package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("CODEXK8S_HTTP_ADDR", "")
	t.Setenv("CODEXK8S_CONTROL_PLANE_GRPC_TARGET", "codex-k8s-control-plane:9090")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT", "")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STT_MODEL", "")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.ControlPlaneGRPCTarget != "codex-k8s-control-plane:9090" {
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
