package runtimedeploy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestEvaluateRuntimeReuse_AllowsFastPathWhenFingerprintMatches(t *testing.T) {
	t.Parallel()

	svc, params, k8s := newRuntimeReuseTestService(t)
	fingerprint, err := svc.buildRuntimeFingerprint(context.Background(), params)
	if err != nil {
		t.Fatalf("buildRuntimeFingerprint() error = %v", err)
	}
	if err := svc.persistRuntimeFingerprint(context.Background(), params.Namespace, fingerprint); err != nil {
		t.Fatalf("persistRuntimeFingerprint() error = %v", err)
	}

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if !result.Reusable {
		t.Fatalf("expected reusable=true, got %#v", result)
	}
	if got, want := result.FingerprintHash, fingerprint.Hash; got != want {
		t.Fatalf("expected fingerprint hash %q, got %q", want, got)
	}
	if got := k8s.namespaces[params.Namespace].Annotations[runtimeFingerprintJSONAnnotationKey]; got == "" {
		t.Fatal("expected persisted fingerprint annotation to be stored")
	}
}

func TestEvaluateRuntimeReuse_ReturnsFingerprintMissingWhenNamespaceHasNoFingerprint(t *testing.T) {
	t.Parallel()

	svc, params, _ := newRuntimeReuseTestService(t)

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if result.Reusable {
		t.Fatalf("expected reusable=false, got %#v", result)
	}
	if got, want := result.Reason, runtimeReuseReasonFingerprintMissing; got != want {
		t.Fatalf("expected reason %q, got %q", want, got)
	}
}

func TestEvaluateRuntimeReuse_ReturnsFingerprintMismatchWhenRenderedManifestsDrift(t *testing.T) {
	t.Parallel()

	svc, params, _ := newRuntimeReuseTestService(t)
	fingerprint, err := svc.buildRuntimeFingerprint(context.Background(), params)
	if err != nil {
		t.Fatalf("buildRuntimeFingerprint() error = %v", err)
	}
	if err := svc.persistRuntimeFingerprint(context.Background(), params.Namespace, fingerprint); err != nil {
		t.Fatalf("persistRuntimeFingerprint() error = %v", err)
	}

	manifestPath := filepath.Join(svc.repoSnapshotPath(params.TargetEnv, "codex-k8s", "codex-k8s", params.BuildRef), "deploy", "base", "app.yaml")
	if err := os.WriteFile(manifestPath, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
spec:
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
        - name: app
          image: example.com/app:drifted
`), 0o644); err != nil {
		t.Fatalf("write drifted manifest: %v", err)
	}

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if result.Reusable {
		t.Fatalf("expected reusable=false, got %#v", result)
	}
	if got, want := result.Reason, runtimeReuseReasonFingerprintMismatch; got != want {
		t.Fatalf("expected reason %q, got %q", want, got)
	}
}

func TestEvaluateRuntimeReuse_RejectsTerminatingNamespace(t *testing.T) {
	t.Parallel()

	svc, params, k8s := newRuntimeReuseTestService(t)
	state := k8s.namespaces[params.Namespace]
	state.Terminating = true
	k8s.namespaces[params.Namespace] = state

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if got, want := result.Reason, runtimeReuseReasonNamespaceTerminating; got != want {
		t.Fatalf("expected reason %q, got %q", want, got)
	}
}

func TestEvaluateRuntimeReuse_RejectsActiveRuntimeDeployTask(t *testing.T) {
	t.Parallel()

	svc, params, _ := newRuntimeReuseTestService(t)
	svc.tasks = &fakeRuntimeReuseTasksRepo{
		active: runtimedeploytaskrepo.Task{
			RunID:     "run-active",
			Namespace: params.Namespace,
			Status:    entitytypes.RuntimeDeployTaskStatusRunning,
			BuildRef:  "89abcdef0123456789abcdef0123456789abcdef",
		},
		activeOK: true,
	}

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if got, want := result.Reason, runtimeReuseReasonActiveTask; got != want {
		t.Fatalf("expected reason %q, got %q", want, got)
	}
	if got, want := result.EffectiveBuildRef, "89abcdef0123456789abcdef0123456789abcdef"; got != want {
		t.Fatalf("expected effective build ref %q, got %q", want, got)
	}
}

func TestEvaluateRuntimeReuse_RejectsMutableBuildRef(t *testing.T) {
	t.Parallel()

	svc, params, _ := newRuntimeReuseTestService(t)
	params.BuildRef = "main"

	result, err := svc.EvaluateRuntimeReuse(context.Background(), params)
	if err != nil {
		t.Fatalf("EvaluateRuntimeReuse() error = %v", err)
	}
	if got, want := result.Reason, runtimeReuseReasonBuildRefNotImmutable; got != want {
		t.Fatalf("expected reason %q, got %q", want, got)
	}
}

func TestBuildRuntimeFingerprint_EnrichesMissingScopeFromRunPayload(t *testing.T) {
	t.Parallel()

	svc, params, _ := newRuntimeReuseTestService(t)
	svc.runs = &fakeRuntimeReuseRunReader{
		run: agentrunrepo.Run{
			ID:        params.RunID,
			ProjectID: params.ProjectID,
			RunPayload: json.RawMessage(`{
				"agent":{"key":"dev"},
				"issue":{"number":312}
			}`),
		},
	}

	expected, err := svc.buildRuntimeFingerprint(context.Background(), params)
	if err != nil {
		t.Fatalf("buildRuntimeFingerprint() with explicit scope error = %v", err)
	}

	params.ProjectID = ""
	params.IssueNumber = 0
	params.AgentKey = ""

	fingerprint, err := svc.buildRuntimeFingerprint(context.Background(), params)
	if err != nil {
		t.Fatalf("buildRuntimeFingerprint() with missing scope error = %v", err)
	}
	if got, want := fingerprint.ProjectID, expected.ProjectID; got != want {
		t.Fatalf("expected project_id %q, got %q", want, got)
	}
	if got, want := fingerprint.IssueNumber, expected.IssueNumber; got != want {
		t.Fatalf("expected issue_number %d, got %d", want, got)
	}
	if got, want := fingerprint.AgentKey, expected.AgentKey; got != want {
		t.Fatalf("expected agent_key %q, got %q", want, got)
	}
	if got, want := fingerprint.Hash, expected.Hash; got != want {
		t.Fatalf("expected hash %q, got %q", want, got)
	}
}

func newRuntimeReuseTestService(t *testing.T) (*Service, EvaluateReuseParams, *fakeRuntimeReuseKubernetesClient) {
	t.Helper()

	repoRoot := filepath.Join(t.TempDir(), "repo-cache")
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	snapshotRoot := filepath.Join(repoRoot, "github", "codex-k8s", "codex-k8s", commitSHA)
	mustWriteRuntimeReuseTestRepo(t, snapshotRoot, commitSHA)

	namespace := "codex-k8s-dev-1"
	k8s := &fakeRuntimeReuseKubernetesClient{
		namespaces: map[string]RuntimeNamespaceState{
			namespace: {
				Name:        namespace,
				Labels:      map[string]string{"codex-k8s.dev/managed-by": "codex-k8s-worker"},
				Annotations: map[string]string{},
			},
		},
	}
	svc := &Service{
		cfg: Config{
			RepositoryRoot: repoRoot,
		},
		k8s:   k8s,
		tasks: &fakeRuntimeReuseTasksRepo{},
	}

	params := EvaluateReuseParams{
		RunID:              "run-reuse",
		ProjectID:          "project-1",
		IssueNumber:        312,
		AgentKey:           "dev",
		RuntimeMode:        "full-env",
		Namespace:          namespace,
		TargetEnv:          "ai",
		SlotNo:             1,
		RepositoryFullName: "codex-k8s/codex-k8s",
		ServicesYAMLPath:   "services.yaml",
		BuildRef:           commitSHA,
	}
	return svc, params, k8s
}

func mustWriteRuntimeReuseTestRepo(t *testing.T, repoRoot string, commitSHA string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".git", "HEAD"), []byte(commitSHA+"\n"), 0o644); err != nil {
		t.Fatalf("write .git/HEAD: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "deploy", "base"), 0o755); err != nil {
		t.Fatalf("mkdir deploy/base: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "services.yaml"), []byte(`apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: runtime-reuse-test
spec:
  environments:
    ai:
      namespaceTemplate: "{{ .Namespace }}"
  infrastructure:
    - name: baseline
      manifests:
        - path: deploy/base/infra.yaml
  services:
    - name: app
      manifests:
        - path: deploy/base/app.yaml
`), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "deploy", "base", "infra.yaml"), []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: infra
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
data:
  env: {{ envOr "CODEXK8S_ENV" "" }}
`), 0o644); err != nil {
		t.Fatalf("write infra manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "deploy", "base", "app.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
spec:
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
        - name: app
          image: example.com/app:{{ envOr "CODEXK8S_BUILD_REF" "" }}
`), 0o644); err != nil {
		t.Fatalf("write app manifest: %v", err)
	}
}

type fakeRuntimeReuseKubernetesClient struct {
	namespaces map[string]RuntimeNamespaceState
}

func (f *fakeRuntimeReuseKubernetesClient) EnsureNamespace(_ context.Context, _ string) error {
	return nil
}

func (f *fakeRuntimeReuseKubernetesClient) GetManagedRunNamespace(_ context.Context, namespace string) (RuntimeNamespaceState, bool, error) {
	state, ok := f.namespaces[namespace]
	return state, ok, nil
}

func (f *fakeRuntimeReuseKubernetesClient) UpsertNamespaceAnnotations(_ context.Context, namespace string, annotations map[string]string) error {
	state, ok := f.namespaces[namespace]
	if !ok {
		state = RuntimeNamespaceState{Name: namespace, Labels: map[string]string{}, Annotations: map[string]string{}}
	}
	if state.Annotations == nil {
		state.Annotations = map[string]string{}
	}
	for key, value := range annotations {
		state.Annotations[key] = value
	}
	f.namespaces[namespace] = state
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) UpsertSecret(_ context.Context, _ string, _ string, _ map[string][]byte) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) UpsertTLSSecret(_ context.Context, _ string, _ string, _ map[string][]byte) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) UpsertConfigMap(_ context.Context, _ string, _ string, _ map[string]string) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) GetSecretData(_ context.Context, _ string, _ string) (map[string][]byte, bool, error) {
	return nil, false, nil
}

func (*fakeRuntimeReuseKubernetesClient) DeleteJobIfExists(_ context.Context, _ string, _ string) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) WaitForJobComplete(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) GetJobLogs(_ context.Context, _ string, _ string, _ int64) (string, error) {
	return "", nil
}

func (*fakeRuntimeReuseKubernetesClient) WaitForDeploymentReady(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) WaitForStatefulSetReady(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) WaitForDaemonSetReady(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (*fakeRuntimeReuseKubernetesClient) ApplyManifest(_ context.Context, _ []byte, _ string, _ string) ([]AppliedResourceRef, error) {
	return nil, nil
}

type fakeRuntimeReuseTasksRepo struct {
	active   runtimedeploytaskrepo.Task
	activeOK bool
}

type fakeRuntimeReuseRunReader struct {
	run agentrunrepo.Run
}

func (*fakeRuntimeReuseTasksRepo) UpsertDesired(_ context.Context, _ runtimedeploytaskrepo.UpsertDesiredParams) (runtimedeploytaskrepo.Task, error) {
	return runtimedeploytaskrepo.Task{}, nil
}

func (*fakeRuntimeReuseTasksRepo) GetByRunID(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (f *fakeRuntimeReuseTasksRepo) FindActiveByNamespace(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	return f.active, f.activeOK, nil
}

func (*fakeRuntimeReuseTasksRepo) ClaimNext(_ context.Context, _ runtimedeploytaskrepo.ClaimParams) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, nil
}

func (*fakeRuntimeReuseTasksRepo) MarkSucceeded(_ context.Context, _ runtimedeploytaskrepo.MarkSucceededParams) (bool, error) {
	return false, nil
}

func (*fakeRuntimeReuseTasksRepo) MarkFailed(_ context.Context, _ runtimedeploytaskrepo.MarkFailedParams) (bool, error) {
	return false, nil
}

func (*fakeRuntimeReuseTasksRepo) RenewLease(_ context.Context, _ runtimedeploytaskrepo.RenewLeaseParams) (bool, error) {
	return false, nil
}

func (*fakeRuntimeReuseTasksRepo) Requeue(_ context.Context, _ runtimedeploytaskrepo.RequeueParams) (bool, error) {
	return false, nil
}

func (*fakeRuntimeReuseTasksRepo) RequestAction(_ context.Context, _ runtimedeploytaskrepo.RequestActionParams) (runtimedeploytaskrepo.RequestActionResult, error) {
	return runtimedeploytaskrepo.RequestActionResult{}, nil
}

func (*fakeRuntimeReuseTasksRepo) ListRecent(_ context.Context, _ runtimedeploytaskrepo.ListFilter) ([]runtimedeploytaskrepo.Task, int, error) {
	return nil, 0, nil
}

func (*fakeRuntimeReuseTasksRepo) AppendLog(_ context.Context, _ runtimedeploytaskrepo.AppendLogParams) error {
	return nil
}

func (*fakeRuntimeReuseTasksRepo) CleanupTaskLogsUpdatedBefore(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeRuntimeReuseRunReader) GetByID(_ context.Context, _ string) (agentrunrepo.Run, bool, error) {
	return f.run, true, nil
}
