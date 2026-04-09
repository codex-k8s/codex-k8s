package missioncontrol

import (
	"strings"

	missioncontrolrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const (
	warmupReconcileGatingBlockingGaps    = "open_blocking_continuity_gaps"
	warmupTransportGatingFreshness       = "provider_freshness_not_ready"
	warmupTransportGatingCoverage        = "provider_coverage_not_ready"
	warmupTransportGatingGraphProjection = "graph_projection_not_ready"
	warmupTransportGatingLaunchPolicy    = "launch_policy_not_ready"
)

func enrichWarmupSummary(summary valuetypes.MissionControlWarmupSummary, watermarks []missioncontrolrepo.WorkspaceWatermark) valuetypes.MissionControlWarmupSummary {
	summary.ProviderFreshnessStatus = enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope
	summary.ProviderCoverageStatus = enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope
	summary.GraphProjectionStatus = enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope
	summary.LaunchPolicyStatus = enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope

	for _, watermark := range watermarks {
		switch watermark.WatermarkKind {
		case enumtypes.MissionControlWorkspaceWatermarkKindProviderFreshness:
			summary.ProviderFreshnessStatus = watermark.Status
		case enumtypes.MissionControlWorkspaceWatermarkKindProviderCoverage:
			summary.ProviderCoverageStatus = watermark.Status
		case enumtypes.MissionControlWorkspaceWatermarkKindGraphProjection:
			summary.GraphProjectionStatus = watermark.Status
		case enumtypes.MissionControlWorkspaceWatermarkKindLaunchPolicy:
			summary.LaunchPolicyStatus = watermark.Status
		}
	}

	summary.ReadyForReconcile = summary.BlockingGapCount == 0
	summary.ReconcileGatingReason = ""
	if !summary.ReadyForReconcile {
		summary.ReconcileGatingReason = warmupReconcileGatingBlockingGaps
	}

	summary.ReadyForTransport = summary.ReadyForReconcile &&
		warmupStatusReady(summary.ProviderFreshnessStatus) &&
		warmupStatusReady(summary.ProviderCoverageStatus) &&
		warmupStatusReady(summary.GraphProjectionStatus) &&
		warmupStatusReady(summary.LaunchPolicyStatus)
	summary.TransportGatingReason = resolveWarmupTransportGatingReason(summary)

	return summary
}

func warmupStatusReady(status enumtypes.MissionControlWorkspaceWatermarkStatus) bool {
	return status == enumtypes.MissionControlWorkspaceWatermarkStatusFresh
}

func resolveWarmupTransportGatingReason(summary valuetypes.MissionControlWarmupSummary) string {
	switch {
	case !summary.ReadyForReconcile:
		return summary.ReconcileGatingReason
	case !warmupStatusReady(summary.ProviderFreshnessStatus):
		return warmupTransportGatingFreshness
	case !warmupStatusReady(summary.ProviderCoverageStatus):
		return warmupTransportGatingCoverage
	case !warmupStatusReady(summary.GraphProjectionStatus):
		return warmupTransportGatingGraphProjection
	case !warmupStatusReady(summary.LaunchPolicyStatus):
		return warmupTransportGatingLaunchPolicy
	default:
		return ""
	}
}

func normalizeWarmupGatingReason(value string) string {
	return strings.TrimSpace(value)
}
