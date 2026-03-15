package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
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

func TestBuildPromptProjectDocs_FiltersByRole(t *testing.T) {
	t.Parallel()

	stack := &servicescfg.Stack{
		Spec: servicescfg.Spec{
			ProjectDocs: []servicescfg.ProjectDocRef{
				{Path: "README.md"},
				{Repository: "service-api", Path: "docs/arch", Roles: []string{"sa", "dev"}},
				{Path: "docs/ops", Roles: []string{"sre"}},
			},
		},
	}

	docs := buildPromptProjectDocs(stack, "dev")
	if len(docs) != 2 {
		t.Fatalf("docs len=%d, want 2", len(docs))
	}
	if docs[0].Path != "README.md" || docs[1].Path != "docs/arch" {
		t.Fatalf("unexpected docs order/content: %+v", docs)
	}
	if docs[1].Repository != "service-api" {
		t.Fatalf("docs[1].repository=%q, want service-api", docs[1].Repository)
	}

	sreDocs := buildPromptProjectDocs(stack, "sre")
	if len(sreDocs) != 2 {
		t.Fatalf("sre docs len=%d, want 2", len(sreDocs))
	}
	if sreDocs[1].Path != "docs/ops" {
		t.Fatalf("unexpected sre docs: %+v", sreDocs)
	}
}

func TestBuildPromptProjectDocs_DedupByPathWithPriority(t *testing.T) {
	t.Parallel()

	stack := &servicescfg.Stack{
		Spec: servicescfg.Spec{
			ProjectDocs: []servicescfg.ProjectDocRef{
				{Repository: "service-orders", Path: "docs/architecture.md", Description: "service copy"},
				{Repository: "policy-docs", Path: "docs/architecture.md", Description: "policy copy"},
				{Repository: "orchestrator", Path: "docs/runtime.md", Description: "orchestrator"},
			},
		},
	}

	docs := buildPromptProjectDocs(stack, "dev")
	if len(docs) != 2 {
		t.Fatalf("docs len=%d, want 2", len(docs))
	}
	if docs[0].Repository != "policy-docs" {
		t.Fatalf("docs[0].repository=%q, want policy-docs", docs[0].Repository)
	}
	if docs[0].Description != "policy copy" {
		t.Fatalf("docs[0].description=%q, want policy copy", docs[0].Description)
	}
}

func TestBuildPromptRoleContext_DefaultAndKnownRole(t *testing.T) {
	t.Parallel()

	known := buildPromptRoleContext(resolvedRunContext{
		Payload: querytypes.RunPayload{
			Agent: &querytypes.RunPayloadAgent{Key: "qa"},
		},
	})
	if known.AgentKey != "qa" {
		t.Fatalf("known role key=%q, want qa", known.AgentKey)
	}
	if len(known.Capabilities) == 0 {
		t.Fatalf("known role capabilities must not be empty")
	}

	unknown := buildPromptRoleContext(resolvedRunContext{
		Payload: querytypes.RunPayload{
			Agent: &querytypes.RunPayloadAgent{Key: "custom-role"},
		},
	})
	if unknown.AgentKey != "custom-role" {
		t.Fatalf("unknown role key=%q, want custom-role", unknown.AgentKey)
	}
	if len(unknown.Capabilities) != 1 {
		t.Fatalf("unknown role capabilities len=%d, want 1", len(unknown.Capabilities))
	}
}

func TestPromptContextIncludesBuiltInUserInteractionToolsForDevRun(t *testing.T) {
	t.Parallel()

	runPayload, err := json.Marshal(querytypes.RunPayload{
		Project: querytypes.RunPayloadProject{
			ID:           "project-1",
			RepositoryID: "repo-1",
		},
		Repository: querytypes.RunPayloadRepository{
			FullName: "codex-k8s/codex-k8s",
		},
		Agent: &querytypes.RunPayloadAgent{
			Key: "dev",
		},
		Trigger: &querytypes.RunPayloadTrigger{
			Kind: "dev",
		},
	})
	if err != nil {
		t.Fatalf("marshal run payload: %v", err)
	}

	service := &Service{
		cfg: Config{
			ServerName:         "codex-k8s-control-plane-mcp",
			InternalMCPBaseURL: "http://platform-control-plane.example/mcp",
			RepositoryRoot:     t.TempDir(),
		},
		runs: &promptContextTestRunsRepository{
			run: agentrunrepo.Run{
				ID:            "run-1",
				CorrelationID: "corr-1",
				ProjectID:     "project-1",
				Status:        "running",
				RunPayload:    runPayload,
			},
		},
		repos: &promptContextTestReposRepository{
			repository: repocfgrepo.RepositoryBinding{
				ID:               "repo-1",
				ProjectID:        "project-1",
				Provider:         "github",
				Owner:            "codex-k8s",
				Name:             "codex-k8s",
				ServicesYAMLPath: "services.yaml",
			},
		},
		toolCatalog: DefaultToolCatalog(),
		now:         time.Now,
	}

	result, err := service.PromptContext(context.Background(), SessionContext{
		RunID:         "run-1",
		CorrelationID: "corr-1",
		ProjectID:     "project-1",
		RuntimeMode:   agentdomain.RuntimeModeFullEnv,
	})
	if err != nil {
		t.Fatalf("PromptContext() error = %v", err)
	}

	if !promptContextHasTool(result.Context.MCP.Tools, ToolMCPUserNotify) {
		t.Fatalf("prompt context tools do not include %q", ToolMCPUserNotify)
	}
	if !promptContextHasTool(result.Context.MCP.Tools, ToolMCPUserDecisionRequest) {
		t.Fatalf("prompt context tools do not include %q", ToolMCPUserDecisionRequest)
	}
}

func promptContextHasTool(tools []ToolCapability, name ToolName) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

type promptContextTestRunsRepository struct {
	run agentrunrepo.Run
}

func (r *promptContextTestRunsRepository) CreatePendingIfAbsent(context.Context, agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	return agentrunrepo.CreateResult{}, nil
}

func (r *promptContextTestRunsRepository) GetByID(context.Context, string) (agentrunrepo.Run, bool, error) {
	return r.run, true, nil
}

func (r *promptContextTestRunsRepository) CancelActiveByID(context.Context, string) (bool, error) {
	return false, nil
}

func (r *promptContextTestRunsRepository) ListRecentByProject(context.Context, string, string, int, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *promptContextTestRunsRepository) SearchRecentByProjectIssueOrPullRequest(context.Context, string, string, int64, int64, int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *promptContextTestRunsRepository) ListRunIDsByRepositoryIssue(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *promptContextTestRunsRepository) ListRunIDsByRepositoryPullRequest(context.Context, string, int64, int) ([]string, error) {
	return nil, nil
}

func (r *promptContextTestRunsRepository) SetWaitContext(context.Context, agentrunrepo.SetWaitContextParams) (bool, error) {
	return false, nil
}

func (r *promptContextTestRunsRepository) ClearWaitContextIfMatches(context.Context, agentrunrepo.ClearWaitContextParams) (bool, error) {
	return false, nil
}

type promptContextTestReposRepository struct {
	repository repocfgrepo.RepositoryBinding
}

func (r *promptContextTestReposRepository) ListForProject(context.Context, string, int) ([]repocfgrepo.RepositoryBinding, error) {
	return nil, nil
}

func (r *promptContextTestReposRepository) GetByID(context.Context, string) (repocfgrepo.RepositoryBinding, bool, error) {
	return r.repository, true, nil
}

func (r *promptContextTestReposRepository) Upsert(context.Context, repocfgrepo.UpsertParams) (repocfgrepo.RepositoryBinding, error) {
	return repocfgrepo.RepositoryBinding{}, nil
}

func (r *promptContextTestReposRepository) Delete(context.Context, string, string) error {
	return nil
}

func (r *promptContextTestReposRepository) FindByProviderExternalID(context.Context, string, int64) (repocfgrepo.FindResult, bool, error) {
	return repocfgrepo.FindResult{}, false, nil
}

func (r *promptContextTestReposRepository) FindByProviderOwnerName(context.Context, string, string, string) (repocfgrepo.FindResult, bool, error) {
	return repocfgrepo.FindResult{}, false, nil
}

func (r *promptContextTestReposRepository) GetTokenEncrypted(context.Context, string) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *promptContextTestReposRepository) GetBotTokenEncrypted(context.Context, string) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *promptContextTestReposRepository) UpsertBotParams(context.Context, repocfgrepo.RepositoryBotParamsUpsertParams) error {
	return nil
}

func (r *promptContextTestReposRepository) UpsertPreflightReport(context.Context, repocfgrepo.RepositoryPreflightReportUpsertParams) error {
	return nil
}

func (r *promptContextTestReposRepository) AcquirePreflightLock(context.Context, repocfgrepo.RepositoryPreflightLockAcquireParams) (string, bool, error) {
	return "", false, nil
}

func (r *promptContextTestReposRepository) ReleasePreflightLock(context.Context, string, string) error {
	return nil
}

func (r *promptContextTestReposRepository) SetTokenEncryptedForAll(context.Context, []byte) (int64, error) {
	return 0, nil
}
