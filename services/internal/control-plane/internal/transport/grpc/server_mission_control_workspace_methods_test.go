package grpc

import (
	"encoding/json"
	"testing"
	"time"

	missioncontroldomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestMissionControlWorkspaceSummary_AggregatesGlobalStateCounters(t *testing.T) {
	t.Parallel()

	summary := missionControlWorkspaceSummary([]missionControlWorkspaceProjectSnapshot{
		{
			projectID: "project-1",
			snapshot: missioncontroldomain.WorkspaceSnapshot{
				Summary: valuetypes.MissionControlWorkspaceSummary{
					RootCount:                  1,
					NodeCount:                  3,
					BlockingGapCount:           1,
					WarningGapCount:            2,
					RecentClosedContextCount:   1,
					WorkingCount:               1,
					WaitingCount:               1,
					BlockedCount:               0,
					ReviewCount:                1,
					RecentCriticalUpdatesCount: 0,
				},
			},
		},
		{
			projectID: "project-2",
			snapshot: missioncontroldomain.WorkspaceSnapshot{
				Summary: valuetypes.MissionControlWorkspaceSummary{
					RootCount:                  2,
					NodeCount:                  4,
					BlockingGapCount:           0,
					WarningGapCount:            1,
					RecentClosedContextCount:   2,
					WorkingCount:               2,
					WaitingCount:               0,
					BlockedCount:               1,
					ReviewCount:                0,
					RecentCriticalUpdatesCount: 1,
				},
			},
		},
	})

	if got, want := summary.GetNodeCount(), int32(7); got != want {
		t.Fatalf("node_count = %d, want %d", got, want)
	}
	if got, want := summary.GetWorkingCount(), int32(3); got != want {
		t.Fatalf("working_count = %d, want %d", got, want)
	}
	if got, want := summary.GetBlockedCount(), int32(1); got != want {
		t.Fatalf("blocked_count = %d, want %d", got, want)
	}
	if got, want := summary.GetRecentCriticalUpdatesCount(), int32(1); got != want {
		t.Fatalf("recent_critical_updates_count = %d, want %d", got, want)
	}
}

func TestMissionControlNodeDetailsToProto_RunUsesLifecycleFieldsFromProjectionPayload(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.March, 18, 15, 4, 0, 0, time.UTC)
	finishedAt := time.Date(2026, time.March, 18, 16, 9, 0, 0, time.UTC)
	raw, err := json.Marshal(valuetypes.MissionControlRunProjectionPayload{
		RunID:              "run-545",
		AgentKey:           "dev",
		LastStatus:         "succeeded",
		RuntimeMode:        "full-env",
		TriggerLabel:       "run:dev",
		BranchHead:         "2623e7f2dae8d280d4e82263ff42e98a023afd6f",
		IssueRef:           "codex-k8s/codex-k8s#545",
		PullRequestRef:     "codex-k8s/codex-k8s/pull/552",
		CandidateNamespace: "codex-issue-3278207d1cd3-i545",
		StartedAt:          &startedAt,
		FinishedAt:         &finishedAt,
		LastHeartbeatAt:    timePtr(time.Date(2026, time.March, 18, 16, 8, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("marshal run payload: %v", err)
	}

	protoDetails := missionControlNodeDetailsToProto(missioncontroldomain.EntityDetails{
		Entity: missioncontroldomain.Entity{
			EntityKind:        enumtypes.MissionControlEntityKindRun,
			EntityExternalKey: "run-545",
			DetailPayloadJSON: raw,
		},
		Node: valuetypes.MissionControlWorkspaceNode{
			NodeRef: valuetypes.MissionControlEntityRef{
				EntityKind:     enumtypes.MissionControlEntityKindRun,
				EntityPublicID: "run-545",
			},
			Title:             "Run run-545",
			VisibilityTier:    enumtypes.MissionControlWorkspaceVisibilityTierPrimary,
			ActiveState:       enumtypes.MissionControlActiveStateReview,
			ContinuityStatus:  enumtypes.MissionControlContinuityStatusComplete,
			CoverageClass:     enumtypes.MissionControlCoverageClassOpenPrimary,
			RootNodePublicID:  "codex-k8s/codex-k8s#545",
			ProjectionVersion: 11,
		},
	})

	run := protoDetails.GetRun()
	if run == nil {
		t.Fatal("expected run payload")
	}
	if got, want := run.GetCandidateNamespace(), "codex-issue-3278207d1cd3-i545"; got != want {
		t.Fatalf("candidate_namespace = %q, want %q", got, want)
	}
	if got, want := run.GetStartedAt().AsTime(), startedAt; !got.Equal(want) {
		t.Fatalf("started_at = %v, want %v", got, want)
	}
	if got, want := run.GetFinishedAt().AsTime(), finishedAt; !got.Equal(want) {
		t.Fatalf("finished_at = %v, want %v", got, want)
	}
	if got, want := len(run.GetProducedIssueRefs()), 1; got != want {
		t.Fatalf("produced_issue_refs len = %d, want %d", got, want)
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
