package runtimedeploy

import (
	"testing"

	"github.com/codex-k8s/kodex/libs/go/servicescfg"
)

func TestResolveTelegramRuntimeSecretValues_GeneratesRequiredSecrets(t *testing.T) {
	t.Parallel()

	resolver := servicescfg.NewSecretResolver(nil)
	values := map[string]string{
		"KODEX_TELEGRAM_BOT_TOKEN": "bot-token",
	}

	got, err := resolveTelegramRuntimeSecretValues(resolver, "ai", values, nil, nil)
	if err != nil {
		t.Fatalf("resolveTelegramRuntimeSecretValues returned error: %v", err)
	}

	if got.BaseURL != defaultTelegramInteractionAdapterBaseURL {
		t.Fatalf("base url = %q, want %q", got.BaseURL, defaultTelegramInteractionAdapterBaseURL)
	}
	if got.Timeout != defaultTelegramInteractionAdapterTimeout {
		t.Fatalf("timeout = %q, want %q", got.Timeout, defaultTelegramInteractionAdapterTimeout)
	}
	if got.BotToken != "bot-token" {
		t.Fatalf("bot token = %q, want %q", got.BotToken, "bot-token")
	}
	if got.BearerToken == "" {
		t.Fatal("expected bearer token to be generated")
	}
	if got.WebhookSecret == "" {
		t.Fatal("expected webhook secret to be generated")
	}
}

func TestResolveTelegramRuntimeSecretValues_PreservesExistingValues(t *testing.T) {
	t.Parallel()

	resolver := servicescfg.NewSecretResolver(nil)
	existing := map[string][]byte{
		"KODEX_TELEGRAM_BOT_TOKEN":                          []byte("existing-bot-token"),
		"KODEX_TELEGRAM_CHAT_ID":                            []byte("chat-1"),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_BASE_URL":       []byte("http://existing-adapter:8080"),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN":   []byte("existing-bearer"),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET": []byte("existing-webhook"),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT":        []byte("15s"),
	}

	got, err := resolveTelegramRuntimeSecretValues(resolver, "production", map[string]string{}, existing, nil)
	if err != nil {
		t.Fatalf("resolveTelegramRuntimeSecretValues returned error: %v", err)
	}

	if got.BotToken != "existing-bot-token" {
		t.Fatalf("bot token = %q, want %q", got.BotToken, "existing-bot-token")
	}
	if got.ChatID != "chat-1" {
		t.Fatalf("chat id = %q, want %q", got.ChatID, "chat-1")
	}
	if got.BaseURL != "http://existing-adapter:8080" {
		t.Fatalf("base url = %q, want %q", got.BaseURL, "http://existing-adapter:8080")
	}
	if got.BearerToken != "existing-bearer" {
		t.Fatalf("bearer token = %q, want %q", got.BearerToken, "existing-bearer")
	}
	if got.WebhookSecret != "existing-webhook" {
		t.Fatalf("webhook secret = %q, want %q", got.WebhookSecret, "existing-webhook")
	}
	if got.Timeout != "15s" {
		t.Fatalf("timeout = %q, want %q", got.Timeout, "15s")
	}
}
