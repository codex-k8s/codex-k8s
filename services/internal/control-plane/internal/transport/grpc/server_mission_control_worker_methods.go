package grpc

import (
	"context"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	missioncontroldomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/missioncontrol"
	missioncontrolworkerdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/missioncontrolworker"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListMissionControlWarmupProjects(ctx context.Context, req *controlplanev1.ListMissionControlWarmupProjectsRequest) (*controlplanev1.ListMissionControlWarmupProjectsResponse, error) {
	if s.missionControl == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	items, err := s.missionControl.ListWarmupProjects(ctx, clampLimit(req.GetLimit(), 50))
	if err != nil {
		return nil, toStatus(err)
	}

	out := make([]*controlplanev1.MissionControlWarmupProject, 0, len(items))
	for _, item := range items {
		out = append(out, &controlplanev1.MissionControlWarmupProject{
			ProjectId:          strings.TrimSpace(item.ProjectID),
			ProjectName:        strings.TrimSpace(item.ProjectName),
			RepositoryFullName: strings.TrimSpace(item.RepositoryFullName),
		})
	}
	return &controlplanev1.ListMissionControlWarmupProjectsResponse{Items: out}, nil
}

func (s *Server) RunMissionControlWarmup(ctx context.Context, req *controlplanev1.RunMissionControlWarmupRequest) (*controlplanev1.RunMissionControlWarmupResponse, error) {
	if s.missionControl == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.GetProjectId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	result, err := s.missionControl.RunWarmup(ctx, missioncontrolworkerdomain.WarmupRequest{
		ProjectID:     strings.TrimSpace(req.GetProjectId()),
		RequestedBy:   strings.TrimSpace(req.GetRequestedBy()),
		CorrelationID: strings.TrimSpace(req.GetCorrelationId()),
		ForceRebuild:  req.GetForceRebuild(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.RunMissionControlWarmupResponse{
		ProjectId:                    strings.TrimSpace(result.Summary.ProjectID),
		EntityCount:                  result.Summary.EntityCount,
		RelationCount:                result.Summary.RelationCount,
		TimelineEntryCount:           result.Summary.TimelineEntryCount,
		CommandCount:                 result.Summary.CommandCount,
		MaxProjectionVersion:         result.Summary.MaxProjectionVersion,
		RunEntityCount:               result.Summary.RunEntityCount,
		LegacyAgentCount:             result.Summary.LegacyAgentCount,
		ContinuityGapCount:           result.Summary.ContinuityGapCount,
		OpenContinuityGapCount:       result.Summary.OpenContinuityGapCount,
		BlockingGapCount:             result.Summary.BlockingGapCount,
		MissingPullRequestGapCount:   result.Summary.MissingPullRequestGapCount,
		MissingFollowUpIssueGapCount: result.Summary.MissingFollowUpIssueGapCount,
		WatermarkCount:               result.Summary.WatermarkCount,
		ReadyForReconcile:            result.Summary.ReadyForReconcile,
		ReconcileGatingReason:        strings.TrimSpace(result.Summary.ReconcileGatingReason),
		ReadyForTransport:            result.Summary.ReadyForTransport,
		TransportGatingReason:        strings.TrimSpace(result.Summary.TransportGatingReason),
		ProviderFreshnessStatus:      strings.TrimSpace(string(result.Summary.ProviderFreshnessStatus)),
		ProviderCoverageStatus:       strings.TrimSpace(string(result.Summary.ProviderCoverageStatus)),
		GraphProjectionStatus:        strings.TrimSpace(string(result.Summary.GraphProjectionStatus)),
		LaunchPolicyStatus:           strings.TrimSpace(string(result.Summary.LaunchPolicyStatus)),
		BackfilledEntities:           int32(result.BackfilledEntities),
		BackfilledRelations:          int32(result.BackfilledRelations),
		BackfilledTimelines:          int32(result.BackfilledTimelines),
	}, nil
}

func (s *Server) ClaimMissionControlPendingCommands(ctx context.Context, req *controlplanev1.ClaimMissionControlPendingCommandsRequest) (*controlplanev1.ClaimMissionControlPendingCommandsResponse, error) {
	if s.missionControl == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	leaseTTL := time.Duration(0)
	if req.GetLeaseTtl() != nil {
		leaseTTL = req.GetLeaseTtl().AsDuration()
	}

	items, err := s.missionControl.ClaimPendingCommands(
		ctx,
		strings.TrimSpace(req.GetWorkerId()),
		leaseTTL,
		clampLimit(req.GetLimit(), 100),
	)
	if err != nil {
		return nil, toStatus(err)
	}

	out := make([]*controlplanev1.MissionControlPendingCommand, 0, len(items))
	for _, item := range items {
		protoItem := &controlplanev1.MissionControlPendingCommand{
			ProjectId:            strings.TrimSpace(item.ProjectID),
			CommandId:            strings.TrimSpace(item.CommandID),
			CommandKind:          strings.TrimSpace(string(item.CommandKind)),
			EffectiveCommandKind: strings.TrimSpace(string(item.EffectiveCommandKind)),
			Status:               strings.TrimSpace(string(item.Status)),
			CorrelationId:        strings.TrimSpace(item.CorrelationID),
			BusinessIntentKey:    strings.TrimSpace(item.BusinessIntentKey),
			RepositoryFullName:   strings.TrimSpace(item.RepositoryFullName),
			RequestedAt:          timestamppb.New(item.RequestedAt.UTC()),
			UpdatedAt:            timestamppb.New(item.UpdatedAt.UTC()),
		}
		if strings.TrimSpace(item.RetryTargetCommandID) != "" {
			protoItem.RetryTargetCommandId = stringPtrOrNil(strings.TrimSpace(item.RetryTargetCommandID))
		}
		if item.StageNextStep != nil {
			protoItem.StageNextStep = &controlplanev1.MissionControlStageNextStepPayload{
				ThreadKind:   strings.TrimSpace(item.StageNextStep.ThreadKind),
				ThreadNumber: int32(item.StageNextStep.ThreadNo),
				TargetLabel:  strings.TrimSpace(item.StageNextStep.TargetLabel),
			}
		}
		out = append(out, protoItem)
	}
	return &controlplanev1.ClaimMissionControlPendingCommandsResponse{Items: out}, nil
}

func (s *Server) QueueMissionControlCommand(ctx context.Context, req *controlplanev1.QueueMissionControlCommandRequest) (*controlplanev1.MissionControlCommandState, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	command, err := s.missionControlDomain.QueueCommand(ctx, missioncontroldomain.CommandQueueParams{
		ProjectID:     strings.TrimSpace(req.GetProjectId()),
		CommandID:     strings.TrimSpace(req.GetCommandId()),
		StatusMessage: strings.TrimSpace(req.GetStatusMessage()),
		UpdatedAt:     tsToTime(req.GetUpdatedAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return s.commandStateToProto(ctx, command)
}

func (s *Server) MarkMissionControlCommandPendingSync(ctx context.Context, req *controlplanev1.MarkMissionControlCommandPendingSyncRequest) (*controlplanev1.MissionControlCommandState, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	command, err := s.missionControlDomain.MarkCommandPendingSync(ctx, missioncontroldomain.CommandSyncProgressParams{
		ProjectID:           strings.TrimSpace(req.GetProjectId()),
		CommandID:           strings.TrimSpace(req.GetCommandId()),
		StatusMessage:       strings.TrimSpace(req.GetStatusMessage()),
		ProviderDeliveryIDs: trimStringSlice(req.GetProviderDeliveryIds()),
		UpdatedAt:           tsToTime(req.GetUpdatedAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return s.commandStateToProto(ctx, command)
}

func (s *Server) MarkMissionControlCommandReconciled(ctx context.Context, req *controlplanev1.MarkMissionControlCommandReconciledRequest) (*controlplanev1.MissionControlCommandState, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	command, err := s.missionControlDomain.MarkCommandReconciled(ctx, missioncontroldomain.CommandReconcileParams{
		ProjectID:           strings.TrimSpace(req.GetProjectId()),
		CommandID:           strings.TrimSpace(req.GetCommandId()),
		StatusMessage:       strings.TrimSpace(req.GetStatusMessage()),
		ProviderDeliveryIDs: trimStringSlice(req.GetProviderDeliveryIds()),
		UpdatedAt:           tsToTime(req.GetUpdatedAt()),
		ReconciledAt:        tsToTime(req.GetReconciledAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return s.commandStateToProto(ctx, command)
}

func (s *Server) MarkMissionControlCommandFailed(ctx context.Context, req *controlplanev1.MarkMissionControlCommandFailedRequest) (*controlplanev1.MissionControlCommandState, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control worker service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	command, err := s.missionControlDomain.MarkCommandFailed(ctx, missioncontroldomain.CommandFailureParams{
		ProjectID:           strings.TrimSpace(req.GetProjectId()),
		CommandID:           strings.TrimSpace(req.GetCommandId()),
		FailureReason:       enumtypes.MissionControlCommandFailureReason(strings.TrimSpace(req.GetFailureReason())),
		StatusMessage:       strings.TrimSpace(req.GetStatusMessage()),
		ProviderDeliveryIDs: trimStringSlice(req.GetProviderDeliveryIds()),
		UpdatedAt:           tsToTime(req.GetUpdatedAt()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return s.commandStateToProto(ctx, command)
}

func (s *Server) commandStateToProto(ctx context.Context, command missioncontroldomain.Command) (*controlplanev1.MissionControlCommandState, error) {
	if s.missionControlDomain == nil {
		return nil, status.Error(codes.FailedPrecondition, "mission control domain service is not configured")
	}
	statusView, err := s.missionControlDomain.GetCommandStatus(ctx, command.ProjectID, command.ID)
	if err != nil {
		return nil, toStatus(err)
	}
	return missionControlCommandStatusViewToProto(statusView), nil
}

func trimStringSlice(values []string) []string {
	var out []string
	for i := range values {
		if trimmed := strings.TrimSpace(values[i]); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
