package runstatus

import (
	"testing"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	nextstepdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/nextstep"
)

func TestStageDescriptorFromTriggerKind_UsesConfiguredReviseLabels(t *testing.T) {
	t.Parallel()

	labels := nextstepdomain.NewLabels(nextstepdomain.Config{
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
	})

	testCases := []struct {
		name        string
		triggerKind webhookdomain.TriggerKind
		wantStage   string
		wantRun     string
		wantRevise  string
	}{
		{name: "doc audit", triggerKind: webhookdomain.TriggerKindDocAuditRevise, wantStage: "doc-audit", wantRun: "run:docs-audit", wantRevise: "run:docs-audit:revise"},
		{name: "qa", triggerKind: webhookdomain.TriggerKindQARevise, wantStage: "qa", wantRun: "run:quality-assurance", wantRevise: "run:quality-assurance:revise"},
		{name: "release", triggerKind: webhookdomain.TriggerKindReleaseRevise, wantStage: "release", wantRun: "run:ship", wantRevise: "run:ship:revise"},
		{name: "postdeploy", triggerKind: webhookdomain.TriggerKindPostDeployRevise, wantStage: "postdeploy", wantRun: "run:post-release", wantRevise: "run:post-release:revise"},
		{name: "ops", triggerKind: webhookdomain.TriggerKindOpsRevise, wantStage: "ops", wantRun: "run:operations", wantRevise: "run:operations:revise"},
		{name: "self improve", triggerKind: webhookdomain.TriggerKindSelfImproveRevise, wantStage: "self-improve", wantRun: "run:self-patch", wantRevise: "run:self-patch:revise"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			descriptor, ok := stageDescriptorFromTriggerKind(labels, string(testCase.triggerKind))
			if !ok {
				t.Fatalf("stageDescriptorFromTriggerKind() ok = false, want true")
			}
			if descriptor.Stage != testCase.wantStage {
				t.Fatalf("stageDescriptorFromTriggerKind().Stage = %q, want %q", descriptor.Stage, testCase.wantStage)
			}
			if descriptor.RunLabel != testCase.wantRun {
				t.Fatalf("stageDescriptorFromTriggerKind().RunLabel = %q, want %q", descriptor.RunLabel, testCase.wantRun)
			}
			if descriptor.ReviseLabel != testCase.wantRevise {
				t.Fatalf("stageDescriptorFromTriggerKind().ReviseLabel = %q, want %q", descriptor.ReviseLabel, testCase.wantRevise)
			}
		})
	}
}

func TestBuildNextStepActions_SkipsDiscussionMode(t *testing.T) {
	t.Parallel()

	actions := buildNextStepActions("https://platform.codex-k8s.dev", nextstepdomain.DefaultLabels(), runContext{}, commentState{
		RepositoryFullName: "codex-k8s/codex-k8s",
		IssueNumber:        298,
		TriggerKind:        triggerKindDiscussion,
		TriggerLabel:       "mode:discussion",
		DiscussionMode:     true,
	})
	if len(actions) != 0 {
		t.Fatalf("expected no next-step actions for discussion run, got %#v", actions)
	}
}
