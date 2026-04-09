package githubratelimitwait

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainguard "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/githubratelimit"
	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/githubratelimitwait/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/create_wait.sql
	queryCreateWait string
	//go:embed sql/update_wait.sql
	queryUpdateWait string
	//go:embed sql/get_wait_by_id.sql
	queryGetWaitByID string
	//go:embed sql/get_wait_by_signal_id.sql
	queryGetWaitBySignalID string
	//go:embed sql/get_open_wait_by_run_and_contour.sql
	queryGetOpenWaitByRunAndContour string
	//go:embed sql/list_waits_by_run_id.sql
	queryListWaitsByRunID string
	//go:embed sql/claim_next_due_auto_resume_candidate.sql
	queryClaimNextDueAutoResumeCandidate string
	//go:embed sql/insert_evidence.sql
	queryInsertEvidence string
	//go:embed sql/lock_run_for_update.sql
	queryLockRunForUpdate string
	//go:embed sql/list_open_waits_by_run_for_update.sql
	queryListOpenWaitsByRunForUpdate string
	//go:embed sql/clear_dominant_flags_by_run.sql
	queryClearDominantFlagsByRun string
	//go:embed sql/set_dominant_flag_by_id.sql
	querySetDominantFlagByID string
	//go:embed sql/set_run_wait_context.sql
	querySetRunWaitContext string
	//go:embed sql/clear_run_wait_context.sql
	queryClearRunWaitContext string
	//go:embed sql/set_session_backpressure.sql
	querySetSessionBackpressure string
	//go:embed sql/clear_session_backpressure.sql
	queryClearSessionBackpressure string
)

// Repository persists GitHub rate-limit wait aggregate and evidence in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

type waitLifecycleValidationInput struct {
	State                  enumtypes.GitHubRateLimitWaitState
	Confidence             enumtypes.GitHubRateLimitConfidence
	RecoveryHintKind       enumtypes.GitHubRateLimitRecoveryHintKind
	ManualActionKind       enumtypes.GitHubRateLimitManualActionKind
	AutoResumeAttemptsUsed int
	MaxAutoResumeAttempts  int
	ResumeNotBefore        *time.Time
	ResolvedAt             *time.Time
}

// NewRepository constructs PostgreSQL wait repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts one new wait aggregate row.
func (r *Repository) Create(ctx context.Context, params domainrepo.CreateWaitParams) (domainrepo.Wait, error) {
	record, err := normalizeCreateParams(params)
	if err != nil {
		return domainrepo.Wait{}, err
	}

	item, err := scanWait(r.db.QueryRow(
		ctx,
		queryCreateWait,
		record.ProjectID,
		record.RunID,
		string(record.ContourKind),
		string(record.SignalOrigin),
		string(record.OperationClass),
		string(record.State),
		string(record.LimitKind),
		string(record.Confidence),
		string(record.RecoveryHintKind),
		record.SignalID,
		nullableTrimmedText(record.RequestFingerprint),
		record.CorrelationID,
		string(record.ResumeActionKind),
		jsonOrEmptyObject(record.ResumePayloadJSON),
		nullableTrimmedText(string(record.ManualActionKind)),
		record.AutoResumeAttemptsUsed,
		record.MaxAutoResumeAttempts,
		timestamptzPtrOrNil(record.ResumeNotBefore),
		timestamptzPtrOrNil(record.LastResumeAttemptAt),
		timestamptzOrNow(record.FirstDetectedAt),
		timestamptzOrNow(record.LastSignalAt),
		timestamptzPtrOrNil(record.ResolvedAt),
	))
	if err != nil {
		return domainrepo.Wait{}, fmt.Errorf("create github rate-limit wait: %w", err)
	}
	return item, nil
}

// Update mutates one existing wait aggregate.
func (r *Repository) Update(ctx context.Context, params domainrepo.UpdateWaitParams) (domainrepo.Wait, bool, error) {
	record, err := normalizeUpdateParams(params)
	if err != nil {
		return domainrepo.Wait{}, false, err
	}

	item, err := scanWait(r.db.QueryRow(
		ctx,
		queryUpdateWait,
		record.ID,
		string(record.SignalOrigin),
		string(record.OperationClass),
		string(record.State),
		string(record.LimitKind),
		string(record.Confidence),
		string(record.RecoveryHintKind),
		record.SignalID,
		nullableTrimmedText(record.RequestFingerprint),
		record.CorrelationID,
		string(record.ResumeActionKind),
		jsonOrEmptyObject(record.ResumePayloadJSON),
		nullableTrimmedText(string(record.ManualActionKind)),
		record.AutoResumeAttemptsUsed,
		record.MaxAutoResumeAttempts,
		timestamptzPtrOrNil(record.ResumeNotBefore),
		timestamptzPtrOrNil(record.LastResumeAttemptAt),
		timestamptzOrNow(record.LastSignalAt),
		timestamptzPtrOrNil(record.ResolvedAt),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Wait{}, false, nil
		}
		return domainrepo.Wait{}, false, fmt.Errorf("update github rate-limit wait: %w", err)
	}
	return item, true, nil
}

// GetByID returns one wait aggregate by id.
func (r *Repository) GetByID(ctx context.Context, waitID string) (domainrepo.Wait, bool, error) {
	return r.lookupWait(ctx, queryGetWaitByID, "wait id", strings.TrimSpace(waitID))
}

// GetBySignalID returns one wait aggregate by latest signal id.
func (r *Repository) GetBySignalID(ctx context.Context, signalID string) (domainrepo.Wait, bool, error) {
	return r.lookupWait(ctx, queryGetWaitBySignalID, "signal id", strings.TrimSpace(signalID))
}

// GetOpenByRunAndContour returns one open wait for run+contour when present.
func (r *Repository) GetOpenByRunAndContour(ctx context.Context, runID string, contourKind enumtypes.GitHubRateLimitContourKind) (domainrepo.Wait, bool, error) {
	return r.lookupWait(ctx, queryGetOpenWaitByRunAndContour, "run+contour open wait", strings.TrimSpace(runID), string(contourKind))
}

// ListByRunID returns all waits for one run ordered by newest update first.
func (r *Repository) ListByRunID(ctx context.Context, runID string) ([]domainrepo.Wait, error) {
	rows, err := r.db.Query(ctx, queryListWaitsByRunID, strings.TrimSpace(runID))
	if err != nil {
		return nil, fmt.Errorf("list github rate-limit waits by run id: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.WaitRow])
	if err != nil {
		return nil, fmt.Errorf("collect github rate-limit waits by run id: %w", err)
	}

	result := make([]domainrepo.Wait, 0, len(items))
	for _, item := range items {
		result = append(result, waitFromDBModel(item))
	}
	return result, nil
}

// ClaimNextDueAutoResume moves one due wait into auto_resume_in_progress and returns it for worker processing.
func (r *Repository) ClaimNextDueAutoResume(ctx context.Context, dueBefore time.Time, staleInProgressBefore time.Time) (domainrepo.Wait, bool, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Wait{}, false, fmt.Errorf("begin claim github rate-limit auto-resume tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	candidate, err := scanWait(tx.QueryRow(
		ctx,
		queryClaimNextDueAutoResumeCandidate,
		dueBefore.UTC(),
		staleInProgressBefore.UTC(),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Wait{}, false, nil
		}
		return domainrepo.Wait{}, false, fmt.Errorf("claim due github rate-limit wait: %w", err)
	}

	attemptsUsed := candidate.AutoResumeAttemptsUsed
	if candidate.State != enumtypes.GitHubRateLimitWaitStateAutoResumeInProgress {
		attemptsUsed++
	}

	claimed, err := scanWait(tx.QueryRow(
		ctx,
		queryUpdateWait,
		candidate.ID,
		string(enumtypes.GitHubRateLimitSignalOriginWorker),
		string(candidate.OperationClass),
		string(enumtypes.GitHubRateLimitWaitStateAutoResumeInProgress),
		string(candidate.LimitKind),
		string(candidate.Confidence),
		string(candidate.RecoveryHintKind),
		candidate.SignalID,
		nullableTrimmedText(candidate.RequestFingerprint),
		candidate.CorrelationID,
		string(candidate.ResumeActionKind),
		jsonOrEmptyObject(candidate.ResumePayloadJSON),
		nullableTrimmedText(string(candidate.ManualActionKind)),
		attemptsUsed,
		candidate.MaxAutoResumeAttempts,
		timestamptzPtrOrNil(candidate.ResumeNotBefore),
		timestamptzOrNow(dueBefore.UTC()),
		timestamptzOrNow(candidate.LastSignalAt),
		timestamptzPtrOrNil(candidate.ResolvedAt),
	))
	if err != nil {
		return domainrepo.Wait{}, false, fmt.Errorf("mark github rate-limit wait %s in progress: %w", candidate.ID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Wait{}, false, fmt.Errorf("commit claim github rate-limit auto-resume tx: %w", err)
	}
	return claimed, true, nil
}

// AppendEvidence inserts one append-only evidence row.
func (r *Repository) AppendEvidence(ctx context.Context, params domainrepo.CreateEvidenceParams) (domainrepo.Evidence, error) {
	record, err := normalizeEvidenceParams(params)
	if err != nil {
		return domainrepo.Evidence{}, err
	}

	item, err := scanEvidence(r.db.QueryRow(
		ctx,
		queryInsertEvidence,
		record.WaitID,
		string(record.EventKind),
		nullableTrimmedText(record.SignalID),
		nullableTrimmedText(string(record.SignalOrigin)),
		intPtrToPGInt4(record.ProviderStatusCode),
		intPtrToPGInt4(record.RetryAfterSeconds),
		intPtrToPGInt4(record.RateLimitLimit),
		intPtrToPGInt4(record.RateLimitRemaining),
		intPtrToPGInt4(record.RateLimitUsed),
		timestamptzPtrOrNil(record.RateLimitResetAt),
		nullableTrimmedText(record.RateLimitResource),
		nullableTrimmedText(record.GitHubRequestID),
		nullableTrimmedText(record.DocumentationURL),
		nullableTrimmedText(record.MessageExcerpt),
		nullableTrimmedText(record.StderrExcerpt),
		jsonOrEmptyObject(record.PayloadJSON),
		timestamptzOrNow(record.ObservedAt),
	))
	if err != nil {
		return domainrepo.Evidence{}, fmt.Errorf("append github rate-limit evidence: %w", err)
	}
	return item, nil
}

// RefreshRunProjection elects dominant wait and synchronizes typed wait linkage on run/session rows.
func (r *Repository) RefreshRunProjection(ctx context.Context, runID string) (valuetypes.GitHubRateLimitProjectionRefreshResult, error) {
	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("run_id is required")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("begin refresh github rate-limit projection tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := lockRun(ctx, tx, trimmedRunID); err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, err
	}

	openWaits, err := listOpenWaitsForRun(ctx, tx, trimmedRunID)
	if err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, err
	}
	dominant, found := domainguard.ElectDominantWait(openWaits)

	if _, err := tx.Exec(ctx, queryClearDominantFlagsByRun, trimmedRunID); err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("clear github rate-limit dominant flags: %w", err)
	}

	result := valuetypes.GitHubRateLimitProjectionRefreshResult{
		RunID:         trimmedRunID,
		OpenWaitCount: len(openWaits),
		SyncState:     enumtypes.GitHubRateLimitProjectionSyncStateCleared,
	}

	if !found {
		if _, err := tx.Exec(
			ctx,
			queryClearRunWaitContext,
			trimmedRunID,
			string(enumtypes.AgentRunWaitReasonGitHubRateLimit),
			string(enumtypes.AgentRunWaitTargetKindGitHubRateLimitWait),
		); err != nil {
			return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("clear github rate-limit run linkage: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			queryClearSessionBackpressure,
			trimmedRunID,
			string(enumtypes.AgentSessionWaitStateBackpressure),
		); err != nil {
			return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("clear github rate-limit session linkage: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("commit refresh github rate-limit projection tx: %w", err)
		}
		return result, nil
	}

	runLinkRows, err := execRowsAffected(
		ctx,
		tx,
		"set github rate-limit run linkage",
		querySetRunWaitContext,
		trimmedRunID,
		string(enumtypes.AgentRunWaitReasonGitHubRateLimit),
		string(enumtypes.AgentRunWaitTargetKindGitHubRateLimitWait),
		dominant.ID,
		timestamptzPtrOrNil(dominant.ResumeNotBefore),
	)
	if err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, err
	}
	if runLinkRows == 0 {
		// Another coarse wait already owns run linkage, so keep the previous projection untouched.
		result.SyncState = enumtypes.GitHubRateLimitProjectionSyncStateBlockedByRunWaitContext
		return result, nil
	}

	sessionLinkRows, err := execRowsAffected(
		ctx,
		tx,
		"set github rate-limit session linkage",
		querySetSessionBackpressure,
		trimmedRunID,
		string(enumtypes.AgentSessionWaitStateBackpressure),
	)
	if err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, err
	}
	if sessionLinkRows == 0 {
		// Keep run/session linkage atomic: if the session snapshot cannot move to backpressure, rollback both mutations.
		result.SyncState = enumtypes.GitHubRateLimitProjectionSyncStateBlockedBySessionWaitState
		return result, nil
	}

	dominantFlagRows, err := execRowsAffected(ctx, tx, "set github rate-limit dominant wait", querySetDominantFlagByID, dominant.ID)
	if err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, err
	}
	if dominantFlagRows != 1 {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("set github rate-limit dominant wait: expected 1 affected row, got %d", dominantFlagRows)
	}

	result.DominantWaitID = dominant.ID
	result.WaitDeadlineAt = dominant.ResumeNotBefore
	result.SyncState = enumtypes.GitHubRateLimitProjectionSyncStateApplied

	if err := tx.Commit(ctx); err != nil {
		return valuetypes.GitHubRateLimitProjectionRefreshResult{}, fmt.Errorf("commit refresh github rate-limit projection tx: %w", err)
	}
	return result, nil
}

func (r *Repository) lookupWait(ctx context.Context, query string, op string, args ...any) (domainrepo.Wait, bool, error) {
	item, err := scanWait(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Wait{}, false, nil
		}
		return domainrepo.Wait{}, false, fmt.Errorf("lookup github rate-limit wait by %s: %w", op, err)
	}
	return item, true, nil
}

func lockRun(ctx context.Context, tx pgx.Tx, runID string) error {
	var lockedRunID string
	if err := tx.QueryRow(ctx, queryLockRunForUpdate, runID).Scan(&lockedRunID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("agent run %s not found", runID)
		}
		return fmt.Errorf("lock agent run for github rate-limit projection: %w", err)
	}
	return nil
}

func listOpenWaitsForRun(ctx context.Context, tx pgx.Tx, runID string) ([]domainrepo.Wait, error) {
	rows, err := tx.Query(ctx, queryListOpenWaitsByRunForUpdate, runID)
	if err != nil {
		return nil, fmt.Errorf("list open github rate-limit waits for projection refresh: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.WaitRow])
	if err != nil {
		return nil, fmt.Errorf("collect open github rate-limit waits for projection refresh: %w", err)
	}

	result := make([]domainrepo.Wait, 0, len(items))
	for _, item := range items {
		result = append(result, waitFromDBModel(item))
	}
	return result, nil
}

func execRowsAffected(ctx context.Context, tx pgx.Tx, op string, query string, args ...any) (int64, error) {
	commandTag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return commandTag.RowsAffected(), nil
}

func scanWait(row pgx.Row) (domainrepo.Wait, error) {
	var item dbmodel.WaitRow
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.RunID,
		&item.ContourKind,
		&item.SignalOrigin,
		&item.OperationClass,
		&item.State,
		&item.LimitKind,
		&item.Confidence,
		&item.RecoveryHintKind,
		&item.DominantForRun,
		&item.SignalID,
		&item.RequestFingerprint,
		&item.CorrelationID,
		&item.ResumeActionKind,
		&item.ResumePayloadJSON,
		&item.ManualActionKind,
		&item.AutoResumeAttemptsUsed,
		&item.MaxAutoResumeAttempts,
		&item.ResumeNotBefore,
		&item.LastResumeAttemptAt,
		&item.FirstDetectedAt,
		&item.LastSignalAt,
		&item.ResolvedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domainrepo.Wait{}, err
	}
	return waitFromDBModel(item), nil
}

func scanEvidence(row pgx.Row) (domainrepo.Evidence, error) {
	var item dbmodel.EvidenceRow
	err := row.Scan(
		&item.ID,
		&item.WaitID,
		&item.EventKind,
		&item.SignalID,
		&item.SignalOrigin,
		&item.ProviderStatusCode,
		&item.RetryAfterSeconds,
		&item.RateLimitLimit,
		&item.RateLimitRemaining,
		&item.RateLimitUsed,
		&item.RateLimitResetAt,
		&item.RateLimitResource,
		&item.GitHubRequestID,
		&item.DocumentationURL,
		&item.MessageExcerpt,
		&item.StderrExcerpt,
		&item.PayloadJSON,
		&item.ObservedAt,
		&item.CreatedAt,
	)
	if err != nil {
		return domainrepo.Evidence{}, err
	}
	return evidenceFromDBModel(item), nil
}

func normalizeCreateParams(params domainrepo.CreateWaitParams) (domainrepo.CreateWaitParams, error) {
	record := params
	record.ProjectID = strings.TrimSpace(record.ProjectID)
	record.RunID = strings.TrimSpace(record.RunID)
	record.SignalID = strings.TrimSpace(record.SignalID)
	record.CorrelationID = strings.TrimSpace(record.CorrelationID)
	if record.ProjectID == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("project_id is required")
	}
	if record.RunID == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(string(record.ContourKind)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("contour_kind is required")
	}
	if strings.TrimSpace(string(record.SignalOrigin)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("signal_origin is required")
	}
	if strings.TrimSpace(string(record.OperationClass)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("operation_class is required")
	}
	if strings.TrimSpace(string(record.LimitKind)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("limit_kind is required")
	}
	if strings.TrimSpace(string(record.RecoveryHintKind)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("recovery_hint_kind is required")
	}
	if record.SignalID == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("signal_id is required")
	}
	if record.CorrelationID == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("correlation_id is required")
	}
	if strings.TrimSpace(string(record.ResumeActionKind)) == "" {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("resume_action_kind is required")
	}
	if strings.TrimSpace(string(record.State)) == "" {
		record.State = enumtypes.GitHubRateLimitWaitStateOpen
	}
	if strings.TrimSpace(string(record.Confidence)) == "" {
		record.Confidence = enumtypes.GitHubRateLimitConfidenceDeterministic
	}
	if record.AutoResumeAttemptsUsed < 0 {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("auto_resume_attempts_used must be >= 0")
	}
	if record.MaxAutoResumeAttempts < 0 {
		return domainrepo.CreateWaitParams{}, fmt.Errorf("max_auto_resume_attempts must be >= 0")
	}
	if err := validateWaitMutation(
		record.State,
		record.Confidence,
		record.RecoveryHintKind,
		record.ManualActionKind,
		record.AutoResumeAttemptsUsed,
		record.MaxAutoResumeAttempts,
		record.ResumeNotBefore,
		record.ResolvedAt,
	); err != nil {
		return domainrepo.CreateWaitParams{}, err
	}
	if record.FirstDetectedAt.IsZero() {
		record.FirstDetectedAt = time.Now().UTC()
	}
	if record.LastSignalAt.IsZero() {
		record.LastSignalAt = time.Now().UTC()
	}
	return record, nil
}

func normalizeUpdateParams(params domainrepo.UpdateWaitParams) (domainrepo.UpdateWaitParams, error) {
	record := params
	record.ID = strings.TrimSpace(record.ID)
	record.SignalID = strings.TrimSpace(record.SignalID)
	record.CorrelationID = strings.TrimSpace(record.CorrelationID)
	if record.ID == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("id is required")
	}
	if strings.TrimSpace(string(record.SignalOrigin)) == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("signal_origin is required")
	}
	if strings.TrimSpace(string(record.OperationClass)) == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("operation_class is required")
	}
	if strings.TrimSpace(string(record.State)) == "" {
		record.State = enumtypes.GitHubRateLimitWaitStateOpen
	}
	if strings.TrimSpace(string(record.LimitKind)) == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("limit_kind is required")
	}
	if strings.TrimSpace(string(record.Confidence)) == "" {
		record.Confidence = enumtypes.GitHubRateLimitConfidenceDeterministic
	}
	if strings.TrimSpace(string(record.RecoveryHintKind)) == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("recovery_hint_kind is required")
	}
	if record.SignalID == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("signal_id is required")
	}
	if record.CorrelationID == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("correlation_id is required")
	}
	if strings.TrimSpace(string(record.ResumeActionKind)) == "" {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("resume_action_kind is required")
	}
	if record.AutoResumeAttemptsUsed < 0 {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("auto_resume_attempts_used must be >= 0")
	}
	if record.MaxAutoResumeAttempts < 0 {
		return domainrepo.UpdateWaitParams{}, fmt.Errorf("max_auto_resume_attempts must be >= 0")
	}
	if err := validateWaitMutation(
		record.State,
		record.Confidence,
		record.RecoveryHintKind,
		record.ManualActionKind,
		record.AutoResumeAttemptsUsed,
		record.MaxAutoResumeAttempts,
		record.ResumeNotBefore,
		record.ResolvedAt,
	); err != nil {
		return domainrepo.UpdateWaitParams{}, err
	}
	if record.LastSignalAt.IsZero() {
		record.LastSignalAt = time.Now().UTC()
	}
	return record, nil
}

func normalizeEvidenceParams(params domainrepo.CreateEvidenceParams) (domainrepo.CreateEvidenceParams, error) {
	record := params
	record.WaitID = strings.TrimSpace(record.WaitID)
	record.SignalID = strings.TrimSpace(record.SignalID)
	if record.WaitID == "" {
		return domainrepo.CreateEvidenceParams{}, fmt.Errorf("wait_id is required")
	}
	if strings.TrimSpace(string(record.EventKind)) == "" {
		return domainrepo.CreateEvidenceParams{}, fmt.Errorf("event_kind is required")
	}
	if record.ObservedAt.IsZero() {
		record.ObservedAt = time.Now().UTC()
	}
	return record, nil
}

func nullableTrimmedText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func jsonOrEmptyObject(raw json.RawMessage) []byte {
	if len(raw) == 0 || !json.Valid(raw) {
		return []byte(`{}`)
	}
	return raw
}

func timestamptzPtrOrNil(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timestamptzOrNow(value time.Time) pgtype.Timestamptz {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func intPtrToPGInt4(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func validateWaitMutation(
	state enumtypes.GitHubRateLimitWaitState,
	confidence enumtypes.GitHubRateLimitConfidence,
	recoveryHintKind enumtypes.GitHubRateLimitRecoveryHintKind,
	manualActionKind enumtypes.GitHubRateLimitManualActionKind,
	autoResumeAttemptsUsed int,
	maxAutoResumeAttempts int,
	resumeNotBefore *time.Time,
	resolvedAt *time.Time,
) error {
	return validateWaitLifecycle(waitLifecycleValidationInput{
		State:                  state,
		Confidence:             confidence,
		RecoveryHintKind:       recoveryHintKind,
		ManualActionKind:       manualActionKind,
		AutoResumeAttemptsUsed: autoResumeAttemptsUsed,
		MaxAutoResumeAttempts:  maxAutoResumeAttempts,
		ResumeNotBefore:        resumeNotBefore,
		ResolvedAt:             resolvedAt,
	})
}

func validateWaitLifecycle(input waitLifecycleValidationInput) error {
	if input.AutoResumeAttemptsUsed > input.MaxAutoResumeAttempts {
		return fmt.Errorf("auto_resume_attempts_used must be <= max_auto_resume_attempts")
	}
	if input.State == enumtypes.GitHubRateLimitWaitStateResolved && input.ResolvedAt == nil {
		return fmt.Errorf("resolved_at is required when state=resolved")
	}
	if input.State == enumtypes.GitHubRateLimitWaitStateManualActionRequired {
		if strings.TrimSpace(string(input.ManualActionKind)) == "" {
			return fmt.Errorf("manual_action_kind is required when state=manual_action_required")
		}
		if input.AutoResumeAttemptsUsed != input.MaxAutoResumeAttempts {
			return fmt.Errorf("manual_action_required requires exhausted auto-resume budget")
		}
	}
	if input.ResumeNotBefore == nil &&
		(input.RecoveryHintKind != enumtypes.GitHubRateLimitRecoveryHintKindManualOnly ||
			input.Confidence != enumtypes.GitHubRateLimitConfidenceProviderUnclear ||
			input.AutoResumeAttemptsUsed != input.MaxAutoResumeAttempts) {
		return fmt.Errorf("resume_not_before is required outside exhausted manual-only provider-uncertain path")
	}
	return nil
}
