package worker

import (
	"encoding/json"
	"strings"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

func TestResolveRunExecutionContext_FullEnvForDevTrigger(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"trigger":{"kind":"dev"},"issue":{"number":42}}`)
	ctx := resolveRunExecutionContext(
		"run-abc-123",
		"550e8400-e29b-41d4-a716-446655440000",
		payload,
		"codex-issue",
	)

	if ctx.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected full-env runtime mode, got %q", ctx.RuntimeMode)
	}
	if ctx.Namespace == "" {
		t.Fatal("expected non-empty namespace for full-env run")
	}
	if !strings.Contains(ctx.Namespace, "-i42-") {
		t.Fatalf("expected namespace to include issue number, got %q", ctx.Namespace)
	}
}

func TestResolveRunExecutionContext_FullEnvForStageTrigger(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"trigger":{"kind":"vision"},"issue":{"number":43}}`)
	ctx := resolveRunExecutionContext(
		"run-vision-123",
		"550e8400-e29b-41d4-a716-446655440001",
		payload,
		"codex-issue",
	)

	if ctx.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected full-env runtime mode, got %q", ctx.RuntimeMode)
	}
	if ctx.Namespace == "" {
		t.Fatal("expected non-empty namespace for full-env stage run")
	}
}

func TestResolveRunExecutionContext_CodeOnlyWithoutTrigger(t *testing.T) {
	t.Parallel()

	ctx := resolveRunExecutionContext(
		"run-no-trigger",
		"project-1",
		json.RawMessage(`{"event_type":"push"}`),
		"codex-issue",
	)

	assertCodeOnlyWithoutNamespace(t, ctx)
}

func TestResolveRunExecutionContext_RuntimeModeFromPayloadHasPriority(t *testing.T) {
	t.Parallel()

	ctx := resolveRunExecutionContext(
		"run-dev-code-only",
		"project-1",
		json.RawMessage(`{"runtime":{"mode":"code-only"},"trigger":{"kind":"dev"},"issue":{"number":99}}`),
		"codex-issue",
	)

	assertCodeOnlyWithoutNamespace(t, ctx)
}

func TestResolveRunExecutionContext_CodeOnlyKeepsExplicitNamespace(t *testing.T) {
	t.Parallel()

	ctx := resolveRunExecutionContext(
		"run-ai-repair",
		"project-1",
		json.RawMessage(`{"runtime":{"mode":"code-only","namespace":"codex-k8s-prod"},"trigger":{"kind":"ai_repair"},"issue":{"number":45}}`),
		"codex-issue",
	)

	if ctx.RuntimeMode != agentdomain.RuntimeModeCodeOnly {
		t.Fatalf("expected code-only runtime mode, got %q", ctx.RuntimeMode)
	}
	if got, want := ctx.Namespace, "codex-k8s-prod"; got != want {
		t.Fatalf("expected namespace override %q, got %q", want, got)
	}
}

func assertCodeOnlyWithoutNamespace(t *testing.T, ctx valuetypes.RunExecutionContext) {
	t.Helper()

	if ctx.RuntimeMode != agentdomain.RuntimeModeCodeOnly {
		t.Fatalf("expected code-only runtime mode, got %q", ctx.RuntimeMode)
	}
	if ctx.Namespace != "" {
		t.Fatalf("expected empty namespace for code-only runtime, got %q", ctx.Namespace)
	}
}

func TestResolveRunExecutionContext_UsesRuntimeNamespaceOverride(t *testing.T) {
	t.Parallel()

	ctx := resolveRunExecutionContext(
		"run-deploy",
		"project-1",
		json.RawMessage(`{"runtime":{"mode":"full-env","namespace":"codex-k8s-prod","deploy_only":true}}`),
		"codex-issue",
	)

	if ctx.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("expected full-env runtime mode, got %q", ctx.RuntimeMode)
	}
	if got, want := ctx.Namespace, "codex-k8s-prod"; got != want {
		t.Fatalf("expected namespace override %q, got %q", want, got)
	}
}

func TestBuildRunNamespace_LengthAndSanitize(t *testing.T) {
	t.Parallel()

	namespace := buildRunNamespace(
		"CoDeX_Issue",
		"550e8400-e29b-41d4-a716-446655440000",
		"run.with.invalid_chars_and_very_long_identifier_1234567890",
		128,
	)
	if namespace == "" {
		t.Fatal("expected namespace to be generated")
	}
	if len(namespace) > 63 {
		t.Fatalf("expected namespace length <=63, got %d (%q)", len(namespace), namespace)
	}
	if strings.Contains(namespace, "_") || strings.Contains(namespace, ".") {
		t.Fatalf("expected sanitized namespace, got %q", namespace)
	}
}
