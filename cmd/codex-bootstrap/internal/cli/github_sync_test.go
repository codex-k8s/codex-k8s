package cli

import (
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

func TestCollectGitHubVariableKeys(t *testing.T) {
	values := map[string]string{
		"CODEXK8S_PRODUCTION_NAMESPACE": "codex-k8s-prod",
		"CODEXK8S_GITHUB_PAT":           "secret",
		"CODEXK8S_PUBLIC_BASE_URL":      "https://platform.codex-k8s.dev",
		"TARGET_HOST":                   "example.org",
	}

	keys := collectGitHubVariableKeys(values)
	if len(keys) != 2 {
		t.Fatalf("expected 2 variable keys, got %d: %v", len(keys), keys)
	}
	if keys[0] != "CODEXK8S_PRODUCTION_NAMESPACE" || keys[1] != "CODEXK8S_PUBLIC_BASE_URL" {
		t.Fatalf("unexpected variable keys: %v", keys)
	}
}

func TestCollectGitHubLabels(t *testing.T) {
	values := map[string]string{
		"CODEXK8S_RUN_DEV_LABEL":   "run:dev",
		"CODEXK8S_RUN_OPS_LABEL":   "run:ops",
		"CODEXK8S_PUBLIC_BASE_URL": "https://platform.codex-k8s.dev",
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
		"CODEXK8S_OPENAI_API_KEY":            "prod-key",
		"CODEXK8S_AI_OPENAI_API_KEY":         "ai-key",
		"CODEXK8S_PRODUCTION_OPENAI_API_KEY": "prod-override-key",
		"CODEXK8S_AI_DOMAIN":                 "ai.platform.example.dev",
		"CODEXK8S_AI_AI_DOMAIN":              "should-not-be-used",
	}

	keys := append(append([]string(nil), githubRepoSecretKeys...), githubEnvVariableKeys...)
	resolver := servicescfg.NewSecretResolver(nil)

	production := cloneStringMap(values)
	applyEnvironmentOverrides(production, "production", keys, resolver)
	if got, want := production["CODEXK8S_OPENAI_API_KEY"], "prod-override-key"; got != want {
		t.Fatalf("production override mismatch: got %q want %q", got, want)
	}
	if got, want := production["CODEXK8S_AI_DOMAIN"], "ai.platform.example.dev"; got != want {
		t.Fatalf("expected CODEXK8S_AI_DOMAIN to remain unchanged: got %q want %q", got, want)
	}

	ai := cloneStringMap(values)
	applyEnvironmentOverrides(ai, "ai", keys, resolver)
	if got, want := ai["CODEXK8S_OPENAI_API_KEY"], "ai-key"; got != want {
		t.Fatalf("ai override mismatch: got %q want %q", got, want)
	}
	if got, want := ai["CODEXK8S_AI_DOMAIN"], "ai.platform.example.dev"; got != want {
		t.Fatalf("expected CODEXK8S_AI_DOMAIN to remain unchanged: got %q want %q", got, want)
	}
}
