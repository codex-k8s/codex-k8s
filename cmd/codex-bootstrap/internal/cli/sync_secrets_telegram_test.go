package cli

import "testing"

func TestHydrateValuesFromExistingSecrets_ConfiguresTelegramAdapter(t *testing.T) {
	values := map[string]string{
		"CODEXK8S_TELEGRAM_BOT_TOKEN": "bot-token",
	}

	if err := hydrateValuesFromExistingSecrets(values, nil, nil, nil); err != nil {
		t.Fatalf("hydrateValuesFromExistingSecrets returned error: %v", err)
	}

	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BASE_URL"], defaultTelegramInteractionAdapterBaseURL; got != want {
		t.Fatalf("telegram adapter base url = %q, want %q", got, want)
	}
	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT"], defaultTelegramInteractionAdapterTimeout; got != want {
		t.Fatalf("telegram adapter timeout = %q, want %q", got, want)
	}
	if got := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN"]; got == "" {
		t.Fatal("expected telegram interaction adapter bearer token to be generated")
	}
	if got := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET"]; got == "" {
		t.Fatal("expected telegram interaction adapter webhook secret to be generated")
	}
}

func TestBuildRuntimeSecretValues_IncludesInteractionCallbackBaseURL(t *testing.T) {
	values := map[string]string{
		"CODEXK8S_PUBLIC_BASE_URL":               "https://platform.codex-k8s.dev",
		"CODEXK8S_INTERACTION_CALLBACK_BASE_URL": "http://codex-k8s",
	}

	got := buildRuntimeSecretValues(values)

	if got["CODEXK8S_PUBLIC_BASE_URL"] != "https://platform.codex-k8s.dev" {
		t.Fatalf("public base url = %q", got["CODEXK8S_PUBLIC_BASE_URL"])
	}
	if got["CODEXK8S_INTERACTION_CALLBACK_BASE_URL"] != "http://codex-k8s" {
		t.Fatalf("interaction callback base url = %q", got["CODEXK8S_INTERACTION_CALLBACK_BASE_URL"])
	}
}

func TestHydrateValuesFromExistingSecrets_PreservesExistingTelegramSecrets(t *testing.T) {
	values := map[string]string{}
	existingRuntime := map[string][]byte{
		"CODEXK8S_TELEGRAM_BOT_TOKEN":                          []byte("existing-bot-token"),
		"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BASE_URL":       []byte("http://existing-adapter:8080"),
		"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN":   []byte("existing-bearer"),
		"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET": []byte("existing-webhook"),
		"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT":        []byte("15s"),
	}

	if err := hydrateValuesFromExistingSecrets(values, nil, existingRuntime, nil); err != nil {
		t.Fatalf("hydrateValuesFromExistingSecrets returned error: %v", err)
	}

	if got, want := values["CODEXK8S_TELEGRAM_BOT_TOKEN"], "existing-bot-token"; got != want {
		t.Fatalf("telegram bot token = %q, want %q", got, want)
	}
	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BASE_URL"], "http://existing-adapter:8080"; got != want {
		t.Fatalf("telegram adapter base url = %q, want %q", got, want)
	}
	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN"], "existing-bearer"; got != want {
		t.Fatalf("telegram adapter bearer token = %q, want %q", got, want)
	}
	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET"], "existing-webhook"; got != want {
		t.Fatalf("telegram adapter webhook secret = %q, want %q", got, want)
	}
	if got, want := values["CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT"], "15s"; got != want {
		t.Fatalf("telegram adapter timeout = %q, want %q", got, want)
	}
}
