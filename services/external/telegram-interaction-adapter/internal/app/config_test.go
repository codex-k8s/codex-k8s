package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("CODEXK8S_HTTP_ADDR", "")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STATE_PATH", "")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT", "")
	t.Setenv("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.TelegramStatePath != "/var/lib/codex-k8s-telegram-interaction-adapter/state.json" {
		t.Fatalf("TelegramStatePath = %q", cfg.TelegramStatePath)
	}
	if cfg.TelegramHTTPTimeout != "10s" {
		t.Fatalf("TelegramHTTPTimeout = %q, want 10s", cfg.TelegramHTTPTimeout)
	}
	if cfg.CallbackHTTPTimeout != "10s" {
		t.Fatalf("CallbackHTTPTimeout = %q, want 10s", cfg.CallbackHTTPTimeout)
	}
}
