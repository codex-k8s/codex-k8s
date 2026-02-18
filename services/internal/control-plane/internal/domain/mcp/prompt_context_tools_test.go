package mcp

import (
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestBuildPromptRuntimeInventory_DefaultStrategyAndSort(t *testing.T) {
	t.Parallel()

	stack := &servicescfg.Stack{
		Spec: servicescfg.Spec{
			Services: []servicescfg.Service{
				{
					Name:        "worker",
					DeployGroup: "internal",
					DependsOn:   []string{"postgres"},
					Manifests: []servicescfg.ManifestRef{
						{Path: "deploy/base/worker.yaml.tpl"},
					},
				},
				{
					Name:               "api",
					CodeUpdateStrategy: servicescfg.CodeUpdateStrategyRestart,
				},
			},
		},
	}

	inventory := buildPromptRuntimeInventory(stack)
	if len(inventory) != 2 {
		t.Fatalf("inventory len=%d, want 2", len(inventory))
	}
	if inventory[0].Name != "api" {
		t.Fatalf("inventory[0].name=%q, want api", inventory[0].Name)
	}
	if inventory[0].CodeUpdateStrategy != string(servicescfg.CodeUpdateStrategyRestart) {
		t.Fatalf("inventory[0].code_update_strategy=%q", inventory[0].CodeUpdateStrategy)
	}
	if inventory[1].Name != "worker" {
		t.Fatalf("inventory[1].name=%q, want worker", inventory[1].Name)
	}
	if inventory[1].CodeUpdateStrategy != string(servicescfg.CodeUpdateStrategyRebuild) {
		t.Fatalf("inventory[1].code_update_strategy=%q, want rebuild", inventory[1].CodeUpdateStrategy)
	}
	if len(inventory[1].ManifestPaths) != 1 || inventory[1].ManifestPaths[0] != "deploy/base/worker.yaml.tpl" {
		t.Fatalf("unexpected manifest paths: %+v", inventory[1].ManifestPaths)
	}
}

func TestResolvePromptTargetEnv_ForDevTrigger(t *testing.T) {
	t.Parallel()

	runCtx := resolvedRunContext{
		Session: SessionContext{RuntimeMode: agentdomain.RuntimeModeFullEnv, Namespace: "codex-k8s-dev-3"},
		Payload: querytypes.RunPayload{
			Trigger: &querytypes.RunPayloadTrigger{Kind: "dev"},
		},
	}

	env := resolvePromptTargetEnv(runCtx, "production")
	if env != "ai" {
		t.Fatalf("target env=%q, want ai", env)
	}
}

func TestResolvePromptServicesConfigPath_ResolvesRepoSnapshot(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	snapshotPath := filepath.Join(repoRoot, "github", "acme", "demo", "feature-test")
	if err := os.MkdirAll(snapshotPath, 0o755); err != nil {
		t.Fatalf("mkdir snapshot path: %v", err)
	}
	servicesPath := filepath.Join(snapshotPath, "services.yaml")
	if err := os.WriteFile(servicesPath, []byte("apiVersion: codex-k8s.dev/v1alpha1\n"), 0o644); err != nil {
		t.Fatalf("write services.yaml: %v", err)
	}

	svc := &Service{cfg: Config{RepositoryRoot: repoRoot}}
	runCtx := resolvedRunContext{
		Repository: entitytypes.RepositoryBinding{
			Owner:            "acme",
			Name:             "demo",
			ServicesYAMLPath: "services.yaml",
		},
		Payload: querytypes.RunPayload{
			PullRequest: &querytypes.RunPayloadPullRequest{HeadRef: "feature/test"},
		},
	}

	resolved, err := svc.resolvePromptServicesConfigPath(runCtx, "services.yaml")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if resolved != servicesPath {
		t.Fatalf("resolved path=%q, want %q", resolved, servicesPath)
	}
}
