package staff

import (
	"context"
	"testing"

	nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

type stubNextStepGitHubMgmt struct {
	labels        []string
	addedLabels   [][]string
	removedLabels []string
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

func (s *stubNextStepGitHubMgmt) AddIssueLabels(_ context.Context, _ string, _ string, _ string, _ int, labels []string) ([]string, error) {
	copiedLabels := append([]string(nil), labels...)
	s.addedLabels = append(s.addedLabels, copiedLabels)
	s.labels = append(s.labels, copiedLabels...)
	return append([]string(nil), s.labels...), nil
}

func (s *stubNextStepGitHubMgmt) RemoveIssueLabel(_ context.Context, _ string, _ string, _ string, _ int, label string) error {
	s.removedLabels = append(s.removedLabels, label)
	filtered := make([]string, 0, len(s.labels))
	for _, item := range s.labels {
		if item == label {
			continue
		}
		filtered = append(filtered, item)
	}
	s.labels = filtered
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

			result, err := service.previewOrExecuteIssueStageTransition(context.Background(), "token", "kodex", "kodex", 255, testCase.targetLabel, false)
			if err != nil {
				t.Fatalf("previewOrExecuteIssueStageTransition() error = %v", err)
			}
			if len(result.AddedLabels) != 1 || result.AddedLabels[0] != testCase.targetLabel {
				t.Fatalf("previewOrExecuteIssueStageTransition().AddedLabels = %#v, want %q", result.AddedLabels, testCase.targetLabel)
			}
		})
	}
}

func TestPreviewOrExecuteIssueStageTransitionWithCASDetectsRunLabelDrift(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{NextStepLabels: nextstepdomain.DefaultLabels()},
		githubMgmt: &stubNextStepGitHubMgmt{
			labels: []string{"run:qa", "state:in-review"},
		},
	}

	_, err := service.previewOrExecuteIssueStageTransitionWithCAS(
		context.Background(),
		"token",
		"kodex",
		"kodex",
		427,
		"run:release",
		[]string{"run:dev:revise"},
		false,
	)
	if err == nil {
		t.Fatal("expected conflict when current run labels drifted")
	}
}
