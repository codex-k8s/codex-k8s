package grpc

import (
	"context"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Server) ReportChangeGovernanceDraftSignal(
	ctx context.Context,
	req *controlplanev1.ReportChangeGovernanceDraftSignalRequest,
) (*controlplanev1.ReportChangeGovernanceDraftSignalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if s.changeGovernance == nil {
		return nil, status.Error(codes.FailedPrecondition, "change governance service is not configured")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != strings.TrimSpace(runSession.RunID) {
		return nil, status.Error(codes.PermissionDenied, "run_id does not match authenticated run token")
	}
	if req.GetOccurredAt() == nil || !req.GetOccurredAt().IsValid() {
		return nil, status.Error(codes.InvalidArgument, "occurred_at is required")
	}

	result, err := s.changeGovernance.ReportDraftSignal(ctx, querytypes.ChangeGovernanceDraftSignalParams{
		RunID:                runID,
		SignalID:             strings.TrimSpace(req.GetSignalId()),
		CorrelationID:        strings.TrimSpace(req.GetCorrelationId()),
		ProjectID:            strings.TrimSpace(req.GetProjectId()),
		RepositoryFullName:   strings.TrimSpace(req.GetRepositoryFullName()),
		IssueNumber:          int(req.GetIssueNumber()),
		PRNumber:             intPtrFromInt32Value(req.GetPrNumber()),
		BranchName:           strings.TrimSpace(req.GetBranchName()),
		DraftRef:             strings.TrimSpace(req.GetDraftRef()),
		DraftKind:            enumtypes.ChangeGovernanceDraftKind(strings.TrimSpace(req.GetDraftKind())),
		ChangeScopeHints:     changeGovernanceScopeHintsFromProto(req.GetChangeScopeHints()),
		CandidateRiskDrivers: changeGovernanceRiskDriversFromProto(req.GetCandidateRiskDrivers()),
		DraftChecksum:        strings.TrimSpace(req.GetDraftChecksum()),
		OccurredAt:           req.GetOccurredAt().AsTime().UTC(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.ReportChangeGovernanceDraftSignalResponse{
		PackageId:    strings.TrimSpace(result.PackageID),
		DraftState:   strings.TrimSpace(string(result.DraftState)),
		NextStepKind: strings.TrimSpace(string(result.NextStepKind)),
	}, nil
}

func (s *Server) PublishChangeGovernanceWaveMap(
	ctx context.Context,
	req *controlplanev1.PublishChangeGovernanceWaveMapRequest,
) (*controlplanev1.PublishChangeGovernanceWaveMapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if s.changeGovernance == nil {
		return nil, status.Error(codes.FailedPrecondition, "change governance service is not configured")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != strings.TrimSpace(runSession.RunID) {
		return nil, status.Error(codes.PermissionDenied, "run_id does not match authenticated run token")
	}
	if req.GetPublishedAt() == nil || !req.GetPublishedAt().IsValid() {
		return nil, status.Error(codes.InvalidArgument, "published_at is required")
	}

	result, err := s.changeGovernance.PublishWaveMap(ctx, querytypes.ChangeGovernanceWaveMapParams{
		PackageID:     strings.TrimSpace(req.GetPackageId()),
		WaveMapID:     strings.TrimSpace(req.GetWaveMapId()),
		CorrelationID: strings.TrimSpace(req.GetCorrelationId()),
		Waves:         changeGovernanceWavesFromProto(req.GetWaves()),
		PublishedAt:   req.GetPublishedAt().AsTime().UTC(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.PublishChangeGovernanceWaveMapResponse{
		PackageId:         strings.TrimSpace(result.PackageID),
		PublicationState:  strings.TrimSpace(string(result.PublicationState)),
		ProjectionVersion: result.ProjectionVersion,
	}, nil
}

func (s *Server) UpsertChangeGovernanceEvidenceSignal(
	ctx context.Context,
	req *controlplanev1.UpsertChangeGovernanceEvidenceSignalRequest,
) (*controlplanev1.UpsertChangeGovernanceEvidenceSignalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if s.changeGovernance == nil {
		return nil, status.Error(codes.FailedPrecondition, "change governance service is not configured")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != strings.TrimSpace(runSession.RunID) {
		return nil, status.Error(codes.PermissionDenied, "run_id does not match authenticated run token")
	}
	if req.GetOccurredAt() == nil || !req.GetOccurredAt().IsValid() {
		return nil, status.Error(codes.InvalidArgument, "occurred_at is required")
	}

	result, err := s.changeGovernance.UpsertEvidenceSignal(ctx, querytypes.ChangeGovernanceEvidenceSignalParams{
		PackageID:             strings.TrimSpace(req.GetPackageId()),
		SignalID:              strings.TrimSpace(req.GetSignalId()),
		CorrelationID:         strings.TrimSpace(req.GetCorrelationId()),
		ScopeKind:             enumtypes.ChangeGovernanceEvidenceScopeKind(strings.TrimSpace(req.GetScopeKind())),
		ScopeRef:              strings.TrimSpace(req.GetScopeRef()),
		BlockKind:             enumtypes.ChangeGovernanceEvidenceBlockKind(strings.TrimSpace(req.GetBlockKind())),
		ArtifactLinks:         changeGovernanceArtifactLinksFromProto(req.GetArtifactLinks()),
		VerificationStateHint: enumtypes.ChangeGovernanceVerificationMinimumState(strings.TrimSpace(req.GetVerificationStateHint())),
		RequiredByTier:        req.GetRequiredByTier(),
		OccurredAt:            req.GetOccurredAt().AsTime().UTC(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.UpsertChangeGovernanceEvidenceSignalResponse{
		PackageId:                 strings.TrimSpace(result.PackageID),
		EvidenceCompletenessState: strings.TrimSpace(string(result.EvidenceCompletenessState)),
		VerificationMinimumState:  strings.TrimSpace(string(result.VerificationMinimumState)),
		ProjectionVersion:         result.ProjectionVersion,
	}, nil
}

func changeGovernanceScopeHintsFromProto(items []*controlplanev1.ChangeGovernanceScopeHint) []querytypes.ChangeGovernanceScopeHint {
	result := make([]querytypes.ChangeGovernanceScopeHint, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, querytypes.ChangeGovernanceScopeHint{
			ContextKey:  strings.TrimSpace(item.GetContextKey()),
			SurfaceKind: enumtypes.ChangeGovernanceSurfaceKind(strings.TrimSpace(item.GetSurfaceKind())),
		})
	}
	return result
}

func changeGovernanceRiskDriversFromProto(items []string) []enumtypes.ChangeGovernanceRiskDriver {
	result := make([]enumtypes.ChangeGovernanceRiskDriver, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		result = append(result, enumtypes.ChangeGovernanceRiskDriver(value))
	}
	return result
}

func changeGovernanceWavesFromProto(items []*controlplanev1.ChangeGovernanceWaveDraft) []querytypes.ChangeGovernanceWaveDraft {
	result := make([]querytypes.ChangeGovernanceWaveDraft, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, querytypes.ChangeGovernanceWaveDraft{
			WaveKey:             strings.TrimSpace(item.GetWaveKey()),
			PublishOrder:        int(item.GetPublishOrder()),
			DominantIntent:      enumtypes.ChangeGovernanceDominantIntent(strings.TrimSpace(item.GetDominantIntent())),
			BoundedScopeKind:    enumtypes.ChangeGovernanceBoundedScopeKind(strings.TrimSpace(item.GetBoundedScopeKind())),
			Summary:             strings.TrimSpace(item.GetSummary()),
			VerificationTargets: changeGovernanceVerificationTargetsFromProto(item.GetVerificationTargets()),
		})
	}
	return result
}

func changeGovernanceVerificationTargetsFromProto(items []*controlplanev1.ChangeGovernanceVerificationTarget) []querytypes.ChangeGovernanceVerificationTarget {
	result := make([]querytypes.ChangeGovernanceVerificationTarget, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, querytypes.ChangeGovernanceVerificationTarget{
			TargetKind: enumtypes.ChangeGovernanceVerificationTargetKind(strings.TrimSpace(item.GetTargetKind())),
			TargetRef:  strings.TrimSpace(item.GetTargetRef()),
		})
	}
	return result
}

func changeGovernanceArtifactLinksFromProto(items []*controlplanev1.ChangeGovernanceArtifactLinkSeed) []querytypes.ChangeGovernanceArtifactLinkSeed {
	result := make([]querytypes.ChangeGovernanceArtifactLinkSeed, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, querytypes.ChangeGovernanceArtifactLinkSeed{
			ArtifactKind: enumtypes.ChangeGovernanceArtifactKind(strings.TrimSpace(item.GetArtifactKind())),
			ArtifactRef:  strings.TrimSpace(item.GetArtifactRef()),
			RelationKind: enumtypes.ChangeGovernanceArtifactRelationKind(strings.TrimSpace(item.GetRelationKind())),
			DisplayLabel: strings.TrimSpace(item.GetDisplayLabel()),
		})
	}
	return result
}

func intPtrFromInt32Value(value *wrapperspb.Int32Value) *int {
	if value == nil || value.GetValue() <= 0 {
		return nil
	}
	result := int(value.GetValue())
	return &result
}
