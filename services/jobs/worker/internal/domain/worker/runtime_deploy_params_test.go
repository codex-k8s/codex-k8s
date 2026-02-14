package worker

import (
	"encoding/json"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

func TestBuildPrepareRunEnvironmentParams_FromRuntimePayload(t *testing.T) {
	t.Parallel()

	claimed := runqueuerepo.ClaimedRun{
		RunID:      "run-123",
		SlotNo:     7,
		RunPayload: json.RawMessage(`{"project":{"services_yaml":"deploy/services.yaml"},"repository":{"full_name":"codex-k8s/test"},"runtime":{"mode":"full-env","target_env":"ai-staging","namespace":"codex-k8s-ai-staging","build_ref":"abc123","deploy_only":true}}`),
	}
	execution := valuetypes.RunExecutionContext{
		RuntimeMode: agentdomain.RuntimeModeFullEnv,
		Namespace:   "fallback-namespace",
	}

	params := buildPrepareRunEnvironmentParams(claimed, execution)

	if got, want := params.RunID, "run-123"; got != want {
		t.Fatalf("RunID mismatch: got %q want %q", got, want)
	}
	if got, want := params.RuntimeMode, "full-env"; got != want {
		t.Fatalf("RuntimeMode mismatch: got %q want %q", got, want)
	}
	if got, want := params.Namespace, "codex-k8s-ai-staging"; got != want {
		t.Fatalf("Namespace mismatch: got %q want %q", got, want)
	}
	if got, want := params.TargetEnv, "ai-staging"; got != want {
		t.Fatalf("TargetEnv mismatch: got %q want %q", got, want)
	}
	if got, want := params.RepositoryFullName, "codex-k8s/test"; got != want {
		t.Fatalf("RepositoryFullName mismatch: got %q want %q", got, want)
	}
	if got, want := params.ServicesYAMLPath, "deploy/services.yaml"; got != want {
		t.Fatalf("ServicesYAMLPath mismatch: got %q want %q", got, want)
	}
	if got, want := params.BuildRef, "abc123"; got != want {
		t.Fatalf("BuildRef mismatch: got %q want %q", got, want)
	}
	if got, want := params.SlotNo, 7; got != want {
		t.Fatalf("SlotNo mismatch: got %d want %d", got, want)
	}
	if !params.DeployOnly {
		t.Fatal("expected DeployOnly=true")
	}
}

func TestBuildPrepareRunEnvironmentParams_DefaultsSlotEnvForFullEnvRun(t *testing.T) {
	t.Parallel()

	claimed := runqueuerepo.ClaimedRun{
		RunID:      "run-999",
		SlotNo:     3,
		RunPayload: json.RawMessage(`{"repository":{"full_name":"codex-k8s/codex-k8s"},"runtime":{"mode":"full-env"}}`),
	}
	execution := valuetypes.RunExecutionContext{
		RuntimeMode: agentdomain.RuntimeModeFullEnv,
		Namespace:   "legacy-run-namespace",
	}

	params := buildPrepareRunEnvironmentParams(claimed, execution)

	if got, want := params.TargetEnv, "ai"; got != want {
		t.Fatalf("TargetEnv mismatch: got %q want %q", got, want)
	}
	if got := params.Namespace; got != "" {
		t.Fatalf("expected empty namespace to let services.yaml resolve slot namespace, got %q", got)
	}
}
