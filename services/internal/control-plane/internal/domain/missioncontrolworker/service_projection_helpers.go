package missioncontrolworker

import (
	"fmt"
	"strings"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func runProjectionKey(runID string) string {
	return strings.TrimSpace(runID)
}

func runDisplayTitle(runID string) string {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return "Run"
	}
	return fmt.Sprintf("Run %s", runID)
}

func updateSeedLastTimelineAt(seeds map[string]projectionSeed, key string, occurredAt time.Time) {
	seed, ok := seeds[key]
	if !ok || occurredAt.IsZero() {
		return
	}
	if seed.LastTimelineAt == nil || occurredAt.After(*seed.LastTimelineAt) {
		seed.LastTimelineAt = timePointer(occurredAt)
		seeds[key] = seed
	}
}

func coverageClassForIssueState(state string) enumtypes.MissionControlCoverageClass {
	return coverageClassForGitHubState(state)
}

func coverageClassForPullRequestState(state string) enumtypes.MissionControlCoverageClass {
	return coverageClassForGitHubState(state)
}

func coverageClassForGitHubState(state string) enumtypes.MissionControlCoverageClass {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "":
		return enumtypes.MissionControlCoverageClassOutOfScope
	case "open":
		return enumtypes.MissionControlCoverageClassOpenPrimary
	default:
		return enumtypes.MissionControlCoverageClassRecentClosedContext
	}
}

func coverageClassForRun(
	workItemCoverage enumtypes.MissionControlCoverageClass,
	pullRequestCoverage enumtypes.MissionControlCoverageClass,
) enumtypes.MissionControlCoverageClass {
	switch {
	case workItemCoverage == enumtypes.MissionControlCoverageClassOpenPrimary || pullRequestCoverage == enumtypes.MissionControlCoverageClassOpenPrimary:
		return enumtypes.MissionControlCoverageClassOpenPrimary
	case workItemCoverage == enumtypes.MissionControlCoverageClassRecentClosedContext || pullRequestCoverage == enumtypes.MissionControlCoverageClassRecentClosedContext:
		return enumtypes.MissionControlCoverageClassRecentClosedContext
	default:
		return enumtypes.MissionControlCoverageClassOutOfScope
	}
}

func runContinuityStatus(
	runStatus string,
	runCoverage enumtypes.MissionControlCoverageClass,
	pullRequestEntityKey string,
	nextRunEntityKey string,
) enumtypes.MissionControlContinuityStatus {
	if runCoverage == enumtypes.MissionControlCoverageClassOutOfScope {
		return enumtypes.MissionControlContinuityStatusOutOfScope
	}
	if nextRunEntityKey != "" || !isSucceededRunStatus(runStatus) {
		return enumtypes.MissionControlContinuityStatusComplete
	}
	if strings.TrimSpace(pullRequestEntityKey) == "" {
		return enumtypes.MissionControlContinuityStatusMissingPullRequest
	}
	// Follow-up continuity needs explicit link evidence; the shadow foundation must not infer
	// a blocking missing_follow_up_issue from the absence of a newer run alone.
	return enumtypes.MissionControlContinuityStatusOutOfScope
}

func workItemContinuityStatus(runContinuity enumtypes.MissionControlContinuityStatus) enumtypes.MissionControlContinuityStatus {
	switch runContinuity {
	case enumtypes.MissionControlContinuityStatusMissingPullRequest,
		enumtypes.MissionControlContinuityStatusMissingFollowUpIssue,
		enumtypes.MissionControlContinuityStatusStaleProvider,
		enumtypes.MissionControlContinuityStatusOutOfScope:
		return runContinuity
	default:
		return enumtypes.MissionControlContinuityStatusComplete
	}
}

func pullRequestContinuityStatus(
	runContinuity enumtypes.MissionControlContinuityStatus,
	pullRequestCoverage enumtypes.MissionControlCoverageClass,
) enumtypes.MissionControlContinuityStatus {
	if pullRequestCoverage == enumtypes.MissionControlCoverageClassOutOfScope ||
		runContinuity == enumtypes.MissionControlContinuityStatusOutOfScope {
		return enumtypes.MissionControlContinuityStatusOutOfScope
	}
	if runContinuity == enumtypes.MissionControlContinuityStatusMissingFollowUpIssue {
		return enumtypes.MissionControlContinuityStatusMissingFollowUpIssue
	}
	return enumtypes.MissionControlContinuityStatusComplete
}

func isSucceededRunStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "succeeded")
}

func trackRunLineage(
	lineageState map[string]string,
	repositoryFullName string,
	run agentrunrepo.RunLookupItem,
	runEntityKey string,
) string {
	if strings.TrimSpace(runEntityKey) == "" {
		return ""
	}
	contours := runLineageContours(repositoryFullName, run)
	nextRunEntityKey := ""
	if len(contours) > 0 {
		nextRunEntityKey = strings.TrimSpace(lineageState[contours[0]])
	}
	for _, contour := range contours {
		if strings.TrimSpace(contour) == "" {
			continue
		}
		lineageState[contour] = runEntityKey
	}
	return nextRunEntityKey
}

func runLineageContours(repositoryFullName string, run agentrunrepo.RunLookupItem) []string {
	contours := make([]string, 0, 2)
	if run.IssueNumber > 0 {
		contours = append(contours, "issue::"+workItemProjectionKey(repositoryFullName, run.IssueNumber))
	}
	if run.PullRequestNumber > 0 {
		contours = append(contours, "pull_request::"+pullRequestProjectionKey(repositoryFullName, run.PullRequestNumber))
	}
	return contours
}

func buildWorkspaceWatermarks(
	projectID string,
	entitySeeds map[string]projectionSeed,
	gapSeeds map[string]continuityGapSeed,
	observedAt time.Time,
) []workspaceWatermarkSeed {
	var (
		primaryOpenCount    int
		recentClosedCount   int
		runEntityCount      int
		legacyAgentCount    int
		openGapCount        int
		providerEntityCount int
		staleProviderCount  int
		recentClosedStart   *time.Time
		recentClosedEnd     *time.Time
		providerWindowStart *time.Time
		providerWindowEnd   *time.Time
	)

	for _, seed := range entitySeeds {
		switch seed.CoverageClass {
		case enumtypes.MissionControlCoverageClassOpenPrimary:
			primaryOpenCount++
		case enumtypes.MissionControlCoverageClassRecentClosedContext:
			recentClosedCount++
			recentClosedStart = earlierTimePtr(recentClosedStart, seed.ProjectedAt)
			recentClosedEnd = laterTimePtr(recentClosedEnd, seed.ProjectedAt)
		}
		if seed.EntityKind == enumtypes.MissionControlEntityKindRun {
			runEntityCount++
		}
		if seed.EntityKind == enumtypes.MissionControlEntityKindAgent {
			legacyAgentCount++
		}
		if seed.ProviderKind != enumtypes.MissionControlProviderKindGitHub {
			continue
		}
		providerEntityCount++
		if seed.StaleAfter != nil && seed.StaleAfter.Before(observedAt) {
			staleProviderCount++
		}
		providerWindowStart = earlierTimePtr(providerWindowStart, projectionTimestamp(seed))
		providerWindowEnd = laterTimePtr(providerWindowEnd, projectionTimestamp(seed))
	}

	for range gapSeeds {
		openGapCount++
	}

	providerFreshnessStatus := enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope
	providerFreshnessSummary := "Shadow backfill не нашёл GitHub-backed узлы для оценки provider freshness."
	if providerEntityCount > 0 {
		if staleProviderCount > 0 {
			providerFreshnessStatus = enumtypes.MissionControlWorkspaceWatermarkStatusStale
			providerFreshnessSummary = fmt.Sprintf(
				"Shadow provider freshness просрочен для %d из %d GitHub-backed узлов.",
				staleProviderCount,
				providerEntityCount,
			)
		} else {
			providerFreshnessStatus = enumtypes.MissionControlWorkspaceWatermarkStatusFresh
			providerFreshnessSummary = fmt.Sprintf(
				"Shadow provider freshness находится в staleness window для %d GitHub-backed узлов.",
				providerEntityCount,
			)
		}
	}

	graphProjectionStatus := enumtypes.MissionControlWorkspaceWatermarkStatusFresh
	if openGapCount > 0 {
		graphProjectionStatus = enumtypes.MissionControlWorkspaceWatermarkStatusDegraded
	}
	graphProjectionSummary := fmt.Sprintf(
		"Graph projection warmup записал %d узлов, %d run-узлов и %d открытых continuity gaps.",
		len(entitySeeds),
		runEntityCount,
		openGapCount,
	)

	launchPolicyStatus := enumtypes.MissionControlWorkspaceWatermarkStatusFresh
	launchPolicySummary := "Launch policy остаётся platform-canonical; shadow foundation не блокирует stage.next_step.execute."
	if openGapCount > 0 {
		launchPolicyStatus = enumtypes.MissionControlWorkspaceWatermarkStatusDegraded
		launchPolicySummary = fmt.Sprintf(
			"Launch policy остаётся shadow-only, пока %d открытых continuity gaps не закрыты.",
			openGapCount,
		)
	}

	return []workspaceWatermarkSeed{
		{
			watermarkKind:   enumtypes.MissionControlWorkspaceWatermarkKindProviderFreshness,
			status:          providerFreshnessStatus,
			summary:         providerFreshnessSummary,
			windowStartedAt: providerWindowStart,
			windowEndedAt:   providerWindowEnd,
			payloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "shadow_backfill",
				Scope:             "run_evidence",
				EntityCount:       providerEntityCount,
				RunEntityCount:    runEntityCount,
				OpenGapCount:      openGapCount,
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			watermarkKind: enumtypes.MissionControlWorkspaceWatermarkKindProviderCoverage,
			status:        enumtypes.MissionControlWorkspaceWatermarkStatusOutOfScope,
			summary: fmt.Sprintf(
				"Provider mirror coverage ещё не синхронизирован; shadow backfill видит %d open-primary и %d recent-closed узлов из persisted run evidence.",
				primaryOpenCount,
				recentClosedCount,
			),
			windowStartedAt: recentClosedStart,
			windowEndedAt:   recentClosedEnd,
			payloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "shadow_backfill",
				Scope:             "bounded_recent_closed_pending_provider_mirror",
				EntityCount:       len(entitySeeds),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      openGapCount,
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			watermarkKind: enumtypes.MissionControlWorkspaceWatermarkKindGraphProjection,
			status:        graphProjectionStatus,
			summary:       graphProjectionSummary,
			payloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "shadow_backfill",
				Scope:             strings.TrimSpace(projectID),
				EntityCount:       len(entitySeeds),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      openGapCount,
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
		{
			watermarkKind: enumtypes.MissionControlWorkspaceWatermarkKindLaunchPolicy,
			status:        launchPolicyStatus,
			summary:       launchPolicySummary,
			payloadJSON: mustMarshal(valuetypes.MissionControlWorkspaceWatermarkPayload{
				Source:            "shadow_backfill",
				Scope:             "launch_policy",
				EntityCount:       len(entitySeeds),
				RunEntityCount:    runEntityCount,
				OpenGapCount:      openGapCount,
				LegacyAgentCount:  legacyAgentCount,
				RecentClosedCount: recentClosedCount,
				PrimaryOpenCount:  primaryOpenCount,
			}),
		},
	}
}

func projectionTimestamp(seed projectionSeed) time.Time {
	if seed.ProviderUpdatedAt != nil && !seed.ProviderUpdatedAt.IsZero() {
		return seed.ProviderUpdatedAt.UTC()
	}
	return seed.ProjectedAt.UTC()
}

func earlierTimePtr(current *time.Time, candidate time.Time) *time.Time {
	if candidate.IsZero() {
		return current
	}
	if current == nil || candidate.Before(*current) {
		return timePointer(candidate)
	}
	return current
}

func laterTimePtr(current *time.Time, candidate time.Time) *time.Time {
	if candidate.IsZero() {
		return current
	}
	if current == nil || candidate.After(*current) {
		return timePointer(candidate)
	}
	return current
}
