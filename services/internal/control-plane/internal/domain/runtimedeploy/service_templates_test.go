package runtimedeploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codex-k8s/kodex/libs/go/manifesttpl"
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
	t.Setenv("KODEX_KANIKO_CLEANUP", "true")
	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "kodex-dev-1")
	if got, want := vars["KODEX_KANIKO_CLEANUP"], "false"; got != want {
		t.Fatalf("buildTemplateVars ai KODEX_KANIKO_CLEANUP=%q want %q", got, want)
	}
	if got, want := vars["KODEX_HOT_RELOAD"], "true"; got != want {
		t.Fatalf("buildTemplateVars ai KODEX_HOT_RELOAD=%q want %q", got, want)
	}
}

func TestBuildTemplateVars_ProductionPreservesKanikoCleanupValue(t *testing.T) {
	t.Setenv("KODEX_KANIKO_CLEANUP", "true")
	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "production"}, "kodex-prod")
	if got, want := vars["KODEX_KANIKO_CLEANUP"], "true"; got != want {
		t.Fatalf("buildTemplateVars production KODEX_KANIKO_CLEANUP=%q want %q", got, want)
	}
}

func TestBuildTemplateVars_PreservesExplicitPlatformControlPlaneEndpoints(t *testing.T) {
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "kodex-control-plane.kodex-prod.svc.cluster.local:9090")
	t.Setenv("KODEX_CONTROL_PLANE_MCP_BASE_URL", "http://kodex-control-plane.kodex-prod.svc.cluster.local:8081/mcp")

	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-issue-503")
	if got, want := vars["KODEX_CONTROL_PLANE_GRPC_TARGET"], "kodex-control-plane.kodex-prod.svc.cluster.local:9090"; got != want {
		t.Fatalf("grpc target = %q, want %q", got, want)
	}
	if got, want := vars["KODEX_CONTROL_PLANE_MCP_BASE_URL"], "http://kodex-control-plane.kodex-prod.svc.cluster.local:8081/mcp"; got != want {
		t.Fatalf("mcp base url = %q, want %q", got, want)
	}
}

func TestBuildTemplateVars_DoesNotInventNamespaceLocalControlPlaneEndpoints(t *testing.T) {
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "")
	t.Setenv("KODEX_CONTROL_PLANE_MCP_BASE_URL", "")

	svc := &Service{}
	vars := svc.buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-issue-503")
	if got := vars["KODEX_CONTROL_PLANE_GRPC_TARGET"]; got != "" {
		t.Fatalf("grpc target = %q, want empty value", got)
	}
	if got := vars["KODEX_CONTROL_PLANE_MCP_BASE_URL"]; got != "" {
		t.Fatalf("mcp base url = %q, want empty value", got)
	}
}

func TestRenderAppTemplate_UsesLocalControlPlaneDefaultsWhenVarsMissing(t *testing.T) {
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "")
	t.Setenv("KODEX_CONTROL_PLANE_MCP_BASE_URL", "")

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "..", "deploy", "base", "kodex", "app.yaml.tpl"))
	if err != nil {
		t.Fatalf("read app template: %v", err)
	}

	rendered, err := manifesttpl.Render("app", raw, (&Service{}).buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-issue-503"))
	if err != nil {
		t.Fatalf("render app template: %v", err)
	}

	output := string(rendered)
	if !strings.Contains(output, "value: 'kodex-control-plane:9090'") {
		t.Fatalf("rendered app template does not contain local grpc target default:\n%s", output)
	}
	if !strings.Contains(output, "value: 'http://kodex-control-plane:8081/mcp'") {
		t.Fatalf("rendered app template does not contain local mcp base url default:\n%s", output)
	}
}

func TestRenderTelegramAdapterTemplate_UsesLocalControlPlaneDefaultWhenVarsMissing(t *testing.T) {
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "")

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "..", "deploy", "base", "telegram-interaction-adapter", "telegram-interaction-adapter.yaml.tpl"))
	if err != nil {
		t.Fatalf("read telegram adapter template: %v", err)
	}

	rendered, err := manifesttpl.Render("telegram-adapter", raw, (&Service{}).buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-issue-503"))
	if err != nil {
		t.Fatalf("render telegram adapter template: %v", err)
	}

	output := string(rendered)
	if !strings.Contains(output, "value: 'kodex-control-plane:9090'") {
		t.Fatalf("rendered telegram adapter template does not contain local grpc target default:\n%s", output)
	}
	if !strings.Contains(output, "name: KODEX_ENV") {
		t.Fatalf("rendered telegram adapter template does not contain KODEX_ENV:\n%s", output)
	}
}

func TestRenderIngressTemplate_OmitsTelegramWebhookIngressForAI(t *testing.T) {
	t.Setenv("KODEX_CONTROL_PLANE_GRPC_TARGET", "")

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "..", "deploy", "base", "kodex", "ingress.yaml.tpl"))
	if err != nil {
		t.Fatalf("read ingress template: %v", err)
	}

	rendered, err := manifesttpl.Render("ingress", raw, (&Service{}).buildTemplateVars(PrepareParams{TargetEnv: "ai"}, "codex-issue-503"))
	if err != nil {
		t.Fatalf("render ingress template: %v", err)
	}

	output := string(rendered)
	if strings.Contains(output, "name: kodex-telegram-webhook") {
		t.Fatalf("rendered ai ingress must not contain telegram webhook ingress:\n%s", output)
	}
}

func TestRenderIngressTemplate_ContainsTelegramWebhookIngressForProduction(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "..", "deploy", "base", "kodex", "ingress.yaml.tpl"))
	if err != nil {
		t.Fatalf("read ingress template: %v", err)
	}

	rendered, err := manifesttpl.Render("ingress", raw, (&Service{}).buildTemplateVars(PrepareParams{TargetEnv: "production"}, "kodex-prod"))
	if err != nil {
		t.Fatalf("render ingress template: %v", err)
	}

	output := string(rendered)
	if !strings.Contains(output, "name: kodex-telegram-webhook") {
		t.Fatalf("rendered production ingress must contain telegram webhook ingress:\n%s", output)
	}
}

func TestResolveServicesConfigPath_PrefersRepoSnapshotWhenConfigPathIsAbsolute(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	repoServicesPath := filepath.Join(repoRoot, "services.yaml")
	if err := os.WriteFile(repoServicesPath, []byte("apiVersion: kodex.works/v1alpha1\n"), 0o644); err != nil {
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
	if err := os.WriteFile(absolutePath, []byte("apiVersion: kodex.works/v1alpha1\n"), 0o644); err != nil {
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
	if err := os.WriteFile(fullPath, []byte("apiVersion: kodex.works/v1alpha1\n"), 0o644); err != nil {
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
