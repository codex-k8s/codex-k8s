package runtimedeploy

import "testing"

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
}

func TestBuildTemplateVars_ProductionPreservesKanikoCleanupValue(t *testing.T) {
	t.Setenv("CODEXK8S_KANIKO_CLEANUP", "true")
	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "production"}, "codex-k8s-prod")
	if got, want := vars["CODEXK8S_KANIKO_CLEANUP"], "true"; got != want {
		t.Fatalf("buildTemplateVars production CODEXK8S_KANIKO_CLEANUP=%q want %q", got, want)
	}
}

func assertDefaultWorkerReplicas(t *testing.T, targetEnv string, platformReplicas string, want string) {
	t.Helper()

	if got := defaultWorkerReplicas(targetEnv, platformReplicas); got != want {
		t.Fatalf("defaultWorkerReplicas(%q, %q) = %q, want %q", targetEnv, platformReplicas, got, want)
	}
}
