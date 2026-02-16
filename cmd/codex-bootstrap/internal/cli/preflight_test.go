package cli

import (
	"testing"

	gh "github.com/google/go-github/v82/github"
)

func TestMissingRequiredKeys(t *testing.T) {
	values := map[string]string{
		"A": "ok",
		"B": "  ",
		"C": "",
	}
	got := missingRequiredKeys(values, []string{"A", "B", "C", "D"})
	want := []string{"B", "C", "D"}
	if len(got) != len(want) {
		t.Fatalf("unexpected missing keys count: got=%d want=%d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected missing key at %d: got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestSplitRepositoryFullName(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		owner, name, err := splitRepositoryFullName(" codex-k8s / codex-k8s ")
		if err != nil {
			t.Fatalf("splitRepositoryFullName returned error: %v", err)
		}
		if owner != "codex-k8s" || name != "codex-k8s" {
			t.Fatalf("unexpected split result: owner=%q name=%q", owner, name)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, _, err := splitRepositoryFullName("codex-k8s")
		if err == nil {
			t.Fatal("expected error for invalid repository format")
		}
	})
}

func TestResolveWebhookURL(t *testing.T) {
	explicit := resolveWebhookURL(map[string]string{
		"CODEXK8S_GITHUB_WEBHOOK_URL": "https://example.org/hook",
		"CODEXK8S_PRODUCTION_DOMAIN":  "platform.codex-k8s.dev",
	})
	if explicit != "https://example.org/hook" {
		t.Fatalf("expected explicit url, got %q", explicit)
	}

	derived := resolveWebhookURL(map[string]string{
		"CODEXK8S_PRODUCTION_DOMAIN": "platform.codex-k8s.dev",
	})
	if derived != "https://platform.codex-k8s.dev/api/v1/webhooks/github" {
		t.Fatalf("unexpected derived url: %q", derived)
	}
}

func TestHasWebhookURL(t *testing.T) {
	hooks := []*gh.Hook{
		{Config: &gh.HookConfig{URL: gh.Ptr("https://platform.codex-k8s.dev/api/v1/webhooks/github")}},
	}
	if !hasWebhookURL(hooks, "https://platform.codex-k8s.dev/api/v1/webhooks/github") {
		t.Fatal("expected webhook url to be found")
	}
	if hasWebhookURL(hooks, "https://example.org/missing") {
		t.Fatal("did not expect webhook url to be found")
	}
}

func TestHasLabel(t *testing.T) {
	labels := []*gh.Label{
		{Name: gh.Ptr("run:dev")},
		{Name: gh.Ptr("run:ops")},
	}
	if !hasLabel(labels, "run:dev") {
		t.Fatal("expected label to be found")
	}
	if hasLabel(labels, "run:qa") {
		t.Fatal("did not expect label to be found")
	}
}
