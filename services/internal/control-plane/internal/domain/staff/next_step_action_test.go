package staff

import (
	"context"
	"testing"

	nextstepdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/nextstep"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type stubNextStepGitHubMgmt struct {
	labels []string
}

func (s *stubNextStepGitHubMgmt) Preflight(context.Context, valuetypes.GitHubPreflightParams) (valuetypes.GitHubPreflightReport, error) {
	return valuetypes.GitHubPreflightReport{}, nil
}

func (s *stubNextStepGitHubMgmt) GetDefaultBranch(context.Context, string, string, string) (string, error) {
	return "main", nil
}

func (s *stubNextStepGitHubMgmt) GetFile(context.Context, string, string, string, string, string) ([]byte, bool, error) {
	return nil, false, nil
}

func (s *stubNextStepGitHubMgmt) CreatePullRequestWithFiles(context.Context, string, string, string, string, string, string, string, map[string][]byte) (int, string, error) {
	return 0, "", nil
}

func (s *stubNextStepGitHubMgmt) ListIssueLabels(context.Context, string, string, string, int) ([]string, error) {
	return append([]string(nil), s.labels...), nil
}

func (s *stubNextStepGitHubMgmt) AddIssueLabels(context.Context, string, string, string, int, []string) ([]string, error) {
	return nil, nil
}

func (s *stubNextStepGitHubMgmt) RemoveIssueLabel(context.Context, string, string, string, int, string) error {
	return nil
}

var _ githubManagementClient = (*stubNextStepGitHubMgmt)(nil)

func TestPreviewOrExecuteIssueStageTransition_AcceptsConfiguredReviseLabels(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			NextStepLabels: nextstepdomain.NewLabels(nextstepdomain.Config{
				RunDocAudit:          "run:docs-audit",
				RunDocAuditRevise:    "run:docs-audit:revise",
				RunQA:                "run:quality-assurance",
				RunQARevise:          "run:quality-assurance:revise",
				RunRelease:           "run:ship",
				RunReleaseRevise:     "run:ship:revise",
				RunPostDeploy:        "run:post-release",
				RunPostDeployRevise:  "run:post-release:revise",
				RunOps:               "run:operations",
				RunOpsRevise:         "run:operations:revise",
				RunSelfImprove:       "run:self-patch",
				RunSelfImproveRevise: "run:self-patch:revise",
			}),
		},
		githubMgmt: &stubNextStepGitHubMgmt{
			labels: []string{"state:in-review"},
		},
	}

	testCases := []struct {
		name        string
		targetLabel string
	}{
		{name: "doc audit", targetLabel: "run:docs-audit:revise"},
		{name: "qa", targetLabel: "run:quality-assurance:revise"},
		{name: "release", targetLabel: "run:ship:revise"},
		{name: "postdeploy", targetLabel: "run:post-release:revise"},
		{name: "ops", targetLabel: "run:operations:revise"},
		{name: "self improve", targetLabel: "run:self-patch:revise"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.previewOrExecuteIssueStageTransition(context.Background(), "token", "codex-k8s", "codex-k8s", 255, testCase.targetLabel, false)
			if err != nil {
				t.Fatalf("previewOrExecuteIssueStageTransition() error = %v", err)
			}
			if len(result.AddedLabels) != 1 || result.AddedLabels[0] != testCase.targetLabel {
				t.Fatalf("previewOrExecuteIssueStageTransition().AddedLabels = %#v, want %q", result.AddedLabels, testCase.targetLabel)
			}
		})
	}
}
