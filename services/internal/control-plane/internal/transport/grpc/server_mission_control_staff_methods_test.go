package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	missioncontroldomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestBuildMissionControlSnapshotID_EmptySnapshotStable(t *testing.T) {
	t.Parallel()

	query := missionControlSnapshotQuery{
		viewMode:     "board",
		activeFilter: "all_active",
		search:       "abc",
		limit:        50,
	}
	first := buildMissionControlSnapshotID(
		query,
		nil,
		nil,
		time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC),
	)
	second := buildMissionControlSnapshotID(
		query,
		nil,
		nil,
		time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
	)
	if first != second {
		t.Fatalf("empty snapshot ids must be stable, got %q and %q", first, second)
	}
}

func TestCollectMissionControlActiveSetRejectsIncompleteSearchCoverage(t *testing.T) {
	t.Parallel()

	entities := make([]missioncontroldomain.Entity, 0, missionControlSnapshotSearchLimit+1)
	for idx := 0; idx <= missionControlSnapshotSearchLimit; idx++ {
		entities = append(entities, missioncontroldomain.Entity{
			ID:                int64(idx + 1),
			ProjectID:         "proj-1",
			EntityKind:        enumtypes.MissionControlEntityKindWorkItem,
			EntityExternalKey: "repo#" + string(rune('a'+(idx%26))),
			Title:             "Entity",
			ProjectedAt:       time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC),
		})
	}

	server := &Server{
		missionControlDomain: stubMissionControlDomain{
			activeSet: missioncontroldomain.ActiveSet{Entities: entities},
		},
	}
	_, _, err := server.collectMissionControlActiveSet(context.Background(), []string{"proj-1"}, "all_active", 50, true)
	var failedPrecondition errs.FailedPrecondition
	if !errors.As(err, &failedPrecondition) {
		t.Fatalf("expected failed precondition, got %v", err)
	}
}

func TestMissionControlAllowedActionsForDegradedWorkItem(t *testing.T) {
	t.Parallel()

	actions := missionControlAllowedActions(missioncontroldomain.Entity{
		EntityKind: enumtypes.MissionControlEntityKindWorkItem,
		SyncStatus: enumtypes.MissionControlSyncStatusDegraded,
	})
	if len(actions) != 3 {
		t.Fatalf("allowed actions len = %d, want 3", len(actions))
	}
	if actions[0].GetBlockedReason() == "" {
		t.Fatal("expected blocked reason for degraded discussion.create")
	}
	if got, want := actions[2].ActionKind, "command.retry_sync"; got != want {
		t.Fatalf("retry action kind = %q, want %q", got, want)
	}
	if actions[2].GetBlockedReason() != "" {
		t.Fatal("retry_sync must stay available while degraded")
	}
}

func TestMissionControlEntityDetailsToProtoPopulatesProjectionPayloads(t *testing.T) {
	t.Parallel()

	providerSyncAt := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	lastHeartbeat := time.Date(2026, 3, 14, 10, 5, 0, 0, time.UTC)

	t.Run("work item", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(valuetypes.MissionControlWorkItemProjectionPayload{
			RepositoryFullName: "codex-k8s/codex-k8s",
			IssueNumber:        372,
			IssueURL:           "https://github.com/codex-k8s/codex-k8s/issues/372",
			LastRunID:          "run-1",
			LastStatus:         "running",
			TriggerKind:        "dev_revise",
			WorkItemType:       "issue",
			StageLabel:         "run:dev:revise",
			Labels:             []string{"run:dev:revise", "state:in-review"},
			Owner:              "ai-da-stas",
			LastProviderSyncAt: &providerSyncAt,
		})
		if err != nil {
			t.Fatalf("marshal work item payload: %v", err)
		}

		details := missioncontroldomain.EntityDetails{
			Entity: missioncontroldomain.Entity{
				EntityKind:        enumtypes.MissionControlEntityKindWorkItem,
				EntityExternalKey: "codex-k8s/codex-k8s#372",
				DetailPayloadJSON: raw,
				ProviderUpdatedAt: &providerSyncAt,
			},
		}
		protoDetails := missionControlEntityDetailsToProto(details)
		workItem := protoDetails.GetWorkItem()
		if workItem == nil {
			t.Fatal("expected work item payload")
		}
		if got, want := workItem.GetStageLabel(), "run:dev:revise"; got != want {
			t.Fatalf("stage label = %q, want %q", got, want)
		}
		if got, want := workItem.GetOwner(), "ai-da-stas"; got != want {
			t.Fatalf("owner = %q, want %q", got, want)
		}
		if got, want := len(workItem.GetLabels()), 2; got != want {
			t.Fatalf("labels len = %d, want %d", got, want)
		}
		if workItem.LastProviderSyncAt == nil {
			t.Fatal("expected last provider sync at")
		}
	})

	t.Run("pull request", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(valuetypes.MissionControlPullRequestProjectionPayload{
			RepositoryFullName: "codex-k8s/codex-k8s",
			PullRequestNumber:  442,
			PullRequestURL:     "https://github.com/codex-k8s/codex-k8s/pull/442",
			LastRunID:          "run-2",
			LastStatus:         "running",
			BranchHead:         "codex/issue-372",
			BranchBase:         "main",
		})
		if err != nil {
			t.Fatalf("marshal pull request payload: %v", err)
		}

		details := missioncontroldomain.EntityDetails{
			Entity: missioncontroldomain.Entity{
				EntityKind:        enumtypes.MissionControlEntityKindPullRequest,
				EntityExternalKey: "codex-k8s/codex-k8s/pull/442",
				DetailPayloadJSON: raw,
			},
			Relations: []valuetypes.MissionControlRelationView{{
				RelationKind: enumtypes.MissionControlRelationKindLinkedTo,
				SourceEntityRef: valuetypes.MissionControlEntityRef{
					EntityKind:     enumtypes.MissionControlEntityKindWorkItem,
					EntityPublicID: "codex-k8s/codex-k8s#372",
				},
				TargetEntityRef: valuetypes.MissionControlEntityRef{
					EntityKind:     enumtypes.MissionControlEntityKindPullRequest,
					EntityPublicID: "codex-k8s/codex-k8s/pull/442",
				},
			}},
		}
		protoDetails := missionControlEntityDetailsToProto(details)
		pullRequest := protoDetails.GetPullRequest()
		if pullRequest == nil {
			t.Fatal("expected pull request payload")
		}
		if got, want := pullRequest.GetBranchHead(), "codex/issue-372"; got != want {
			t.Fatalf("branch head = %q, want %q", got, want)
		}
		if got, want := len(pullRequest.GetLinkedIssueRefs()), 1; got != want {
			t.Fatalf("linked issue refs len = %d, want %d", got, want)
		}
	})

	t.Run("agent", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(valuetypes.MissionControlAgentProjectionPayload{
			AgentKey:          "dev",
			LastRunID:         "run-3",
			LastStatus:        "waiting",
			RuntimeMode:       "full-env",
			WaitingReason:     "owner_review",
			LastHeartbeatAt:   &lastHeartbeat,
			LastRunRepository: "codex-k8s/codex-k8s",
		})
		if err != nil {
			t.Fatalf("marshal agent payload: %v", err)
		}

		details := missioncontroldomain.EntityDetails{
			Entity: missioncontroldomain.Entity{
				EntityKind:        enumtypes.MissionControlEntityKindAgent,
				EntityExternalKey: "agent/dev",
				DetailPayloadJSON: raw,
			},
		}
		protoDetails := missionControlEntityDetailsToProto(details)
		agent := protoDetails.GetAgent()
		if agent == nil {
			t.Fatal("expected agent payload")
		}
		if got, want := agent.GetRuntimeMode(), "full-env"; got != want {
			t.Fatalf("runtime mode = %q, want %q", got, want)
		}
		if got, want := agent.GetWaitingReason(), "owner_review"; got != want {
			t.Fatalf("waiting reason = %q, want %q", got, want)
		}
		if agent.LastHeartbeatAt == nil {
			t.Fatal("expected last heartbeat")
		}
	})
}

type stubMissionControlDomain struct {
	activeSet missioncontroldomain.ActiveSet
}

func (s stubMissionControlDomain) RunWarmup(context.Context, missioncontroldomain.WarmupRequest) (missioncontroldomain.WarmupSummary, error) {
	return missioncontroldomain.WarmupSummary{}, nil
}

func (stubMissionControlDomain) RefreshWorkspaceProjection(context.Context, missioncontroldomain.WorkspaceRefreshParams) (missioncontroldomain.WorkspaceProjectionSummary, error) {
	return missioncontroldomain.WorkspaceProjectionSummary{}, nil
}

func (stubMissionControlDomain) GetWorkspace(context.Context, missioncontroldomain.WorkspaceQuery) (missioncontroldomain.WorkspaceSnapshot, error) {
	return missioncontroldomain.WorkspaceSnapshot{}, nil
}

func (stubMissionControlDomain) PreviewLaunch(context.Context, missioncontroldomain.LaunchPreviewParams) (missioncontroldomain.LaunchPreview, error) {
	return missioncontroldomain.LaunchPreview{}, nil
}

func (s stubMissionControlDomain) ListActiveSet(context.Context, missioncontroldomain.ActiveSetQuery) (missioncontroldomain.ActiveSet, error) {
	return s.activeSet, nil
}

func (stubMissionControlDomain) GetEntityDetails(context.Context, missioncontroldomain.EntityDetailsQuery) (missioncontroldomain.EntityDetails, error) {
	return missioncontroldomain.EntityDetails{}, nil
}

func (stubMissionControlDomain) UpsertEntity(context.Context, missioncontroldomain.UpsertEntityParams, string) (missioncontroldomain.Entity, error) {
	return missioncontroldomain.Entity{}, nil
}

func (stubMissionControlDomain) UpdateEntityProjection(context.Context, missioncontroldomain.UpdateEntityParams, string) (missioncontroldomain.Entity, error) {
	return missioncontroldomain.Entity{}, nil
}

func (stubMissionControlDomain) ReplaceRelationsForSource(context.Context, missioncontroldomain.ReplaceRelationsParams, string) error {
	return nil
}

func (stubMissionControlDomain) UpsertTimelineEntry(context.Context, missioncontroldomain.UpsertTimelineEntryParams, string) (missioncontroldomain.TimelineEntry, error) {
	return missioncontroldomain.TimelineEntry{}, nil
}

func (stubMissionControlDomain) SubmitCommand(context.Context, missioncontroldomain.SubmitCommandParams) (missioncontroldomain.CommandAdmission, error) {
	return missioncontroldomain.CommandAdmission{}, nil
}

func (stubMissionControlDomain) GetCommandStatus(context.Context, string, string) (missioncontroldomain.CommandStatusView, error) {
	return missioncontroldomain.CommandStatusView{}, nil
}

func (stubMissionControlDomain) QueueCommand(context.Context, missioncontroldomain.CommandQueueParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}

func (stubMissionControlDomain) MarkCommandPendingSync(context.Context, missioncontroldomain.CommandSyncProgressParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}

func (stubMissionControlDomain) MarkCommandReconciled(context.Context, missioncontroldomain.CommandReconcileParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}

func (stubMissionControlDomain) MarkCommandFailed(context.Context, missioncontroldomain.CommandFailureParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}

func (stubMissionControlDomain) CancelCommand(context.Context, missioncontroldomain.CommandCancelParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}

func (stubMissionControlDomain) ApplyApprovalDecision(context.Context, missioncontroldomain.ApprovalDecisionParams) (missioncontroldomain.Command, error) {
	return missioncontroldomain.Command{}, nil
}
