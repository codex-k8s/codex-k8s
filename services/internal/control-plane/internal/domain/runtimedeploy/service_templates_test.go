package runtimedeploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkerReplicas(t *testing.T) {
	t.Parallel()

	assertDefaultWorkerReplicas(t, "production", "2", "3")
	assertDefaultWorkerReplicas(t, "prod", "5", "5")
	assertDefaultWorkerReplicas(t, "ai", "1", "1")
	assertDefaultWorkerReplicas(t, "ai", "", "1")
}

func TestResolveHotReloadFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		targetEnv string
		current   string
		want      string
	}{
		{
			name:      "ai overrides inherited false",
			targetEnv: "ai",
			current:   "false",
			want:      "true",
		},
		{
			name:      "ai default true",
			targetEnv: "ai",
			current:   "",
			want:      "true",
		},
		{
			name:      "production keeps explicit value",
			targetEnv: "production",
			current:   "true",
			want:      "true",
		},
		{
			name:      "production default false",
			targetEnv: "production",
			current:   "",
			want:      "false",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveHotReloadFlag(tt.targetEnv, tt.current); got != tt.want {
				t.Fatalf("resolveHotReloadFlag(%q, %q) = %q, want %q", tt.targetEnv, tt.current, got, tt.want)
			}
		})
	}
}

func TestBuildTemplateVars_AiForcesKanikoCleanupDisabled(t *testing.T) {
	t.Setenv("CODEXK8S_KANIKO_CLEANUP", "true")
	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-k8s-dev-1")
	if got, want := vars["CODEXK8S_KANIKO_CLEANUP"], "false"; got != want {
		t.Fatalf("buildTemplateVars ai CODEXK8S_KANIKO_CLEANUP=%q want %q", got, want)
	}
	if got, want := vars["CODEXK8S_HOT_RELOAD"], "true"; got != want {
		t.Fatalf("buildTemplateVars ai CODEXK8S_HOT_RELOAD=%q want %q", got, want)
	}
}

func TestBuildTemplateVars_ProductionPreservesKanikoCleanupValue(t *testing.T) {
	t.Setenv("CODEXK8S_KANIKO_CLEANUP", "true")
	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "production"}, "codex-k8s-prod")
	if got, want := vars["CODEXK8S_KANIKO_CLEANUP"], "true"; got != want {
		t.Fatalf("buildTemplateVars production CODEXK8S_KANIKO_CLEANUP=%q want %q", got, want)
	}
}

func TestResolveServicesConfigPath_PrefersRepoSnapshotWhenConfigPathIsAbsolute(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	repoServicesPath := filepath.Join(repoRoot, "services.yaml")
	if err := os.WriteFile(repoServicesPath, []byte("apiVersion: codex-k8s.dev/v1alpha1\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", repoServicesPath, err)
	}

	svc := &Service{
		cfg: Config{
			ServicesConfigPath: "/app/services.yaml",
			RepositoryRoot:     repoRoot,
		},
	}
	if got := svc.resolveServicesConfigPath(repoRoot, ""); got != repoServicesPath {
		t.Fatalf("resolveServicesConfigPath() = %q, want %q", got, repoServicesPath)
	}
}

func TestResolveServicesConfigPath_UsesAbsolutePathWhenRepoSnapshotMissing(t *testing.T) {
	t.Parallel()

	absoluteRoot := t.TempDir()
	absolutePath := filepath.Join(absoluteRoot, "services.yaml")
	if err := os.WriteFile(absolutePath, []byte("apiVersion: codex-k8s.dev/v1alpha1\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", absolutePath, err)
	}

	svc := &Service{
		cfg: Config{
			ServicesConfigPath: absolutePath,
		},
	}
	if got := svc.resolveServicesConfigPath(t.TempDir(), ""); got != absolutePath {
		t.Fatalf("resolveServicesConfigPath() = %q, want %q", got, absolutePath)
	}
}

func TestResolveServicesConfigPath_PathFromRunRelativeHasPriority(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pathFromRun := "configs/services.ai.yaml"
	fullPath := filepath.Join(repoRoot, pathFromRun)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, []byte("apiVersion: codex-k8s.dev/v1alpha1\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", fullPath, err)
	}

	svc := &Service{
		cfg: Config{
			ServicesConfigPath: "/app/services.yaml",
			RepositoryRoot:     repoRoot,
		},
	}
	if got := svc.resolveServicesConfigPath(repoRoot, pathFromRun); got != fullPath {
		t.Fatalf("resolveServicesConfigPath(%q) = %q, want %q", pathFromRun, got, fullPath)
	}
}

func assertDefaultWorkerReplicas(t *testing.T, targetEnv string, platformReplicas string, want string) {
	t.Helper()

	if got := defaultWorkerReplicas(targetEnv, platformReplicas); got != want {
		t.Fatalf("defaultWorkerReplicas(%q, %q) = %q, want %q", targetEnv, platformReplicas, got, want)
	}
}
