package cli

import (
	"testing"

	"github.com/codex-k8s/kodex/libs/go/servicescfg"
)

func TestCollectGitHubLabels(t *testing.T) {
	values := map[string]string{
		"KODEX_RUN_DEV_LABEL":   "run:dev",
		"KODEX_RUN_OPS_LABEL":   "run:ops",
		"KODEX_PUBLIC_BASE_URL": "https://platform.kodex.works",
	}

	labels := collectGitHubLabels(values)
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d: %v", len(labels), labels)
	}
	if _, ok := labels["run:dev"]; !ok {
		t.Fatalf("expected run:dev label")
	}
	if _, ok := labels["run:ops"]; !ok {
		t.Fatalf("expected run:ops label")
	}
}

func TestNormalizeGitHubEvents(t *testing.T) {
	events := normalizeGitHubEvents("push, pull_request, push, , issues")
	if len(events) != 3 {
		t.Fatalf("expected 3 unique events, got %d: %v", len(events), events)
	}
	if events[0] != "push" || events[1] != "pull_request" || events[2] != "issues" {
		t.Fatalf("unexpected events order/content: %v", events)
	}
}

func TestApplyGitHubEnvironmentOverrides(t *testing.T) {
	values := map[string]string{
		"KODEX_OPENAI_API_KEY":            "prod-key",
		"KODEX_AI_OPENAI_API_KEY":         "ai-key",
		"KODEX_PRODUCTION_OPENAI_API_KEY": "prod-override-key",
		"KODEX_AI_DOMAIN":                 "ai.platform.example.dev",
		"KODEX_AI_AI_DOMAIN":              "should-not-be-used",
	}

	keys := []string{"KODEX_OPENAI_API_KEY", "KODEX_AI_DOMAIN"}
	resolver := servicescfg.NewSecretResolver(nil)

	production := cloneStringMap(values)
	applyEnvironmentOverrides(production, "production", keys, resolver)
	if got, want := production["KODEX_OPENAI_API_KEY"], "prod-override-key"; got != want {
		t.Fatalf("production override mismatch: got %q want %q", got, want)
	}
	if got, want := production["KODEX_AI_DOMAIN"], "ai.platform.example.dev"; got != want {
		t.Fatalf("expected KODEX_AI_DOMAIN to remain unchanged: got %q want %q", got, want)
	}

	ai := cloneStringMap(values)
	applyEnvironmentOverrides(ai, "ai", keys, resolver)
	if got, want := ai["KODEX_OPENAI_API_KEY"], "ai-key"; got != want {
		t.Fatalf("ai override mismatch: got %q want %q", got, want)
	}
	if got, want := ai["KODEX_AI_DOMAIN"], "ai.platform.example.dev"; got != want {
		t.Fatalf("expected KODEX_AI_DOMAIN to remain unchanged: got %q want %q", got, want)
	}
}
