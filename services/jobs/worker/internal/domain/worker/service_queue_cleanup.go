package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

type namespaceCleanupSource string

const (
	namespaceCleanupSourceWorkerTick namespaceCleanupSource = "worker_tick"
	namespaceCleanupSourceCronJob    namespaceCleanupSource = "cronjob"
	namespaceCleanupLeaseSourceLabel                        = "lease_source="
)

type expiredManagedNamespace struct {
	State                   ManagedNamespaceState
	EffectiveLeaseTTL       time.Duration
	EffectiveLeaseExpiresAt time.Time
	LeaseSource             string
}

func (s *Service) RunNamespaceCleanupOnce(ctx context.Context) error {
	return s.runNamespaceCleanupSweep(ctx, namespaceCleanupSourceCronJob)
}

func (s *Service) cleanupExpiredNamespaces(ctx context.Context) error {
	return s.runNamespaceCleanupSweep(ctx, namespaceCleanupSourceWorkerTick)
}

func (s *Service) runNamespaceCleanupSweep(ctx context.Context, source namespaceCleanupSource) error {
	if !s.cfg.RunNamespaceCleanupEnabled {
		s.logger.Info("skip namespace cleanup sweep: disabled by config", "source", source)
		return nil
	}

	managedNamespaces, err := s.launcher.ListManagedRunNamespaces(ctx, ManagedNamespaceListParams{
		NamespacePrefix: s.cfg.RunNamespacePrefix,
	})
	if err != nil {
		return fmt.Errorf("list managed run namespaces: %w", err)
	}

	now := s.now().UTC()
	expired := s.collectExpiredManagedNamespaces(managedNamespaces, now)
	if len(expired) == 0 {
		s.logger.Debug("namespace cleanup sweep found no expired namespaces", "source", source)
		return nil
	}

	nonTerminalRuns, err := s.runs.ListNonTerminalByRunIDs(ctx, cleanupCandidateRunIDs(expired))
	if err != nil {
		return fmt.Errorf("list non-terminal runs for namespace cleanup: %w", err)
	}
	nonTerminalByRunID := make(map[string]string, len(nonTerminalRuns))
	for _, item := range nonTerminalRuns {
		nonTerminalByRunID[strings.TrimSpace(item.RunID)] = strings.TrimSpace(item.Status)
	}

	deletedCount := 0
	skippedCount := 0
	failedCount := 0

	for _, candidate := range expired {
		execution := valuetypes.RunExecutionContext{
			RuntimeMode: candidate.State.RuntimeMode,
			Namespace:   candidate.State.Namespace,
		}
		leaseDetails := cleanupLeaseDetails(candidate)
		runID := strings.TrimSpace(candidate.State.RunID)

		if runID == "" {
			skippedCount++
			s.logNamespaceCleanupSkip(source, candidate, namespaceCleanupReasonMissingRunIDLabel, leaseDetails)
			s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
				candidate: candidate,
				execution: execution,
				eventType: floweventdomain.EventTypeRunNamespaceCleanupSkipped,
				reason:    namespaceCleanupReasonMissingRunIDLabel,
				source:    source,
				details:   leaseDetails,
			})
			continue
		}

		if status, ok := nonTerminalByRunID[runID]; ok {
			skippedCount++
			details := append(append([]string{}, leaseDetails...), "run_status="+status)
			s.logNamespaceCleanupSkip(source, candidate, namespaceCleanupReasonActiveRunInDB, details)
			s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
				candidate: candidate,
				execution: execution,
				eventType: floweventdomain.EventTypeRunNamespaceCleanupSkipped,
				reason:    namespaceCleanupReasonActiveRunInDB,
				source:    source,
				details:   details,
			})
			continue
		}

		workloads, err := s.launcher.InspectNamespaceWorkloads(ctx, candidate.State.Namespace)
		if err != nil {
			failedCount++
			s.logNamespaceCleanupFailure(
				ctx,
				source,
				candidate,
				execution,
				namespaceCleanupReasonInspectFailed,
				leaseDetails,
				"inspect namespace workloads for cleanup failed",
				err,
			)
			continue
		}

		if workloads.HasActiveWorkloads() {
			skippedCount++
			details := append(append([]string{}, leaseDetails...), workloads.Details()...)
			s.logNamespaceCleanupSkip(source, candidate, namespaceCleanupReasonActiveWorkloads, details)
			s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
				candidate: candidate,
				execution: execution,
				eventType: floweventdomain.EventTypeRunNamespaceCleanupSkipped,
				reason:    namespaceCleanupReasonActiveWorkloads,
				source:    source,
				details:   details,
			})
			continue
		}

		deleted, err := s.launcher.DeleteManagedNamespace(ctx, candidate.State.Namespace)
		if err != nil {
			failedCount++
			s.logNamespaceCleanupFailure(
				ctx,
				source,
				candidate,
				execution,
				namespaceCleanupReasonDeleteFailed,
				leaseDetails,
				"delete expired run namespace failed",
				err,
			)
			continue
		}
		if !deleted {
			skippedCount++
			s.logNamespaceCleanupSkip(source, candidate, namespaceCleanupReasonAlreadyDeleted, leaseDetails)
			s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
				candidate: candidate,
				execution: execution,
				eventType: floweventdomain.EventTypeRunNamespaceCleanupSkipped,
				reason:    namespaceCleanupReasonAlreadyDeleted,
				source:    source,
				details:   leaseDetails,
			})
			continue
		}

		deletedCount++
		s.logger.Info(
			"cleaned expired run namespace",
			"source", source,
			"namespace", candidate.State.Namespace,
			"run_id", runID,
			"expires_at", candidate.EffectiveLeaseExpiresAt.Format(time.RFC3339),
			"details", leaseDetails,
		)
		s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
			candidate: candidate,
			execution: execution,
			eventType: floweventdomain.EventTypeRunNamespaceCleaned,
			reason:    namespaceCleanupReasonTTLExpired,
			source:    source,
			details:   leaseDetails,
		})

		if _, upsertErr := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:       runID,
			Phase:       RunStatusPhaseNamespaceDeleted,
			RuntimeMode: string(candidate.State.RuntimeMode),
			Namespace:   candidate.State.Namespace,
			Deleted:     true,
		}); upsertErr != nil {
			s.logger.Warn("upsert run status comment (namespace cleanup) failed", "run_id", runID, "namespace", candidate.State.Namespace, "err", upsertErr)
		}
	}

	s.logger.Info(
		"namespace cleanup sweep completed",
		"source", source,
		"expired_candidates", len(expired),
		"deleted", deletedCount,
		"skipped", skippedCount,
		"failed", failedCount,
	)
	return nil
}

type namespaceCleanupEventArgs struct {
	candidate expiredManagedNamespace
	execution valuetypes.RunExecutionContext
	eventType floweventdomain.EventType
	reason    namespaceCleanupSkipReason
	source    namespaceCleanupSource
	details   []string
	err       error
}

func (s *Service) insertNamespaceCleanupEventBestEffort(ctx context.Context, args namespaceCleanupEventArgs) {
	extra := namespaceLifecycleEventExtra{
		Reason:                  args.reason,
		GuardrailDetails:        append([]string(nil), args.details...),
		CleanupCommand:          string(args.source),
		NamespaceLeaseTTL:       args.candidate.EffectiveLeaseTTL,
		NamespaceLeaseExpiresAt: args.candidate.EffectiveLeaseExpiresAt,
	}
	if args.err != nil {
		extra.Error = args.err.Error()
	}
	if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
		CorrelationID: args.candidate.State.CorrelationID,
		EventType:     args.eventType,
		RunID:         args.candidate.State.RunID,
		ProjectID:     args.candidate.State.ProjectID,
		Execution:     args.execution,
		Extra:         extra,
	}); err != nil {
		s.logger.Warn(
			"insert namespace cleanup flow event failed",
			"source", args.source,
			"namespace", args.candidate.State.Namespace,
			"run_id", args.candidate.State.RunID,
			"event_type", args.eventType,
			"err", err,
		)
	}
}

func (s *Service) collectExpiredManagedNamespaces(states []ManagedNamespaceState, now time.Time) []expiredManagedNamespace {
	limit := s.cfg.NamespaceLeaseSweepLimit
	if limit <= 0 {
		limit = 200
	}

	expired := make([]expiredManagedNamespace, 0, minInt(limit, len(states)))
	for _, state := range states {
		if len(expired) >= limit {
			break
		}

		effectiveTTL := state.LeaseTTL
		effectiveExpiresAt := state.LeaseExpiresAt.UTC()
		leaseSource := "annotation"

		if effectiveExpiresAt.IsZero() {
			if effectiveTTL <= 0 {
				effectiveTTL = s.cfg.DefaultNamespaceTTL
			}
			if effectiveTTL <= 0 || state.CreatedAt.IsZero() {
				continue
			}
			effectiveExpiresAt = state.CreatedAt.UTC().Add(effectiveTTL)
			leaseSource = "created_at_fallback"
		}
		if effectiveExpiresAt.After(now) {
			continue
		}

		expired = append(expired, expiredManagedNamespace{
			State:                   state,
			EffectiveLeaseTTL:       effectiveTTL,
			EffectiveLeaseExpiresAt: effectiveExpiresAt,
			LeaseSource:             leaseSource,
		})
	}
	return expired
}

func cleanupCandidateRunIDs(candidates []expiredManagedNamespace) []string {
	runIDs := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		runID := strings.TrimSpace(candidate.State.RunID)
		if runID == "" {
			continue
		}
		if _, ok := seen[runID]; ok {
			continue
		}
		seen[runID] = struct{}{}
		runIDs = append(runIDs, runID)
	}
	return runIDs
}

func cleanupLeaseDetails(candidate expiredManagedNamespace) []string {
	details := []string{
		namespaceCleanupLeaseSourceLabel + candidate.LeaseSource,
	}
	if !candidate.EffectiveLeaseExpiresAt.IsZero() {
		details = append(details, "lease_expires_at="+candidate.EffectiveLeaseExpiresAt.Format(time.RFC3339))
	}
	if candidate.EffectiveLeaseTTL > 0 {
		details = append(details, "lease_ttl="+candidate.EffectiveLeaseTTL.String())
	}
	return details
}

func (s *Service) logNamespaceCleanupSkip(source namespaceCleanupSource, candidate expiredManagedNamespace, reason namespaceCleanupSkipReason, details []string) {
	s.logger.Info(
		"skip expired run namespace cleanup",
		"source", source,
		"namespace", candidate.State.Namespace,
		"run_id", candidate.State.RunID,
		"reason", reason,
		"details", details,
	)
}

func (s *Service) logNamespaceCleanupFailure(
	ctx context.Context,
	source namespaceCleanupSource,
	candidate expiredManagedNamespace,
	execution valuetypes.RunExecutionContext,
	reason namespaceCleanupSkipReason,
	details []string,
	message string,
	err error,
) {
	s.logger.Error(
		message,
		"source", source,
		"namespace", candidate.State.Namespace,
		"run_id", candidate.State.RunID,
		"err", err,
	)
	s.insertNamespaceCleanupEventBestEffort(ctx, namespaceCleanupEventArgs{
		candidate: candidate,
		execution: execution,
		eventType: floweventdomain.EventTypeRunNamespaceCleanupFailed,
		reason:    reason,
		source:    source,
		details:   details,
		err:       err,
	})
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
