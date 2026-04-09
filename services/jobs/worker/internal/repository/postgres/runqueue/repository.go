package runqueue

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	rundomain "github.com/codex-k8s/kodex/libs/go/domain/run"
	domainrepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
	querytypes "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/types/query"
)

var (
	//go:embed sql/claim_next_pending_for_update.sql
	queryClaimNextPendingForUpdate string
	//go:embed sql/create_pending_resume_if_absent.sql
	queryCreatePendingResumeIfAbsent string
	//go:embed sql/get_run_id_by_correlation_id.sql
	queryGetRunIDByCorrelationID string
	//go:embed sql/upsert_project.sql
	queryUpsertProject string
	//go:embed sql/ensure_project_exists.sql
	queryEnsureProjectExists string
	//go:embed sql/get_project_settings.sql
	queryGetProjectSettings string
	//go:embed sql/ensure_project_slots.sql
	queryEnsureProjectSlots string
	//go:embed sql/release_expired_slots.sql
	queryReleaseExpiredSlots string
	//go:embed sql/lease_slot.sql
	queryLeaseSlot string
	//go:embed sql/mark_run_running.sql
	queryMarkRunRunning string
	//go:embed sql/claim_running.sql
	queryClaimRunning string
	//go:embed sql/release_stale_running_leases.sql
	queryReleaseStaleRunningLeases string
	//go:embed sql/release_owned_running_leases.sql
	queryReleaseOwnedRunningLeases string
	//go:embed sql/list_running.sql
	queryListRunning string
	//go:embed sql/list_non_terminal_by_run_ids.sql
	queryListNonTerminalByRunIDs string
	//go:embed sql/extend_slot_lease.sql
	queryExtendSlotLease string
	//go:embed sql/mark_run_finished.sql
	queryMarkRunFinished string
	//go:embed sql/mark_slot_releasing.sql
	queryMarkSlotReleasing string
	//go:embed sql/mark_slot_free.sql
	queryMarkSlotFree string
)

// Repository persists run queue state in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL run queue repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreatePendingResumeIfAbsent inserts one pending resume run derived from an existing source run.
func (r *Repository) CreatePendingResumeIfAbsent(ctx context.Context, params domainrepo.CreatePendingResumeParams) (bool, error) {
	sourceRunID := strings.TrimSpace(params.SourceRunID)
	if sourceRunID == "" {
		return false, fmt.Errorf("create pending resume run: source_run_id is required")
	}
	correlationID := strings.TrimSpace(params.CorrelationID)
	if correlationID == "" {
		return false, fmt.Errorf("create pending resume run: correlation_id is required")
	}

	runID := uuid.NewString()
	var insertedRunID string
	err := r.db.QueryRow(ctx, queryCreatePendingResumeIfAbsent, runID, sourceRunID, correlationID).Scan(&insertedRunID)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("insert pending resume run: %w", err)
	}

	var existingRunID string
	if err := r.db.QueryRow(ctx, queryGetRunIDByCorrelationID, correlationID).Scan(&existingRunID); err == nil {
		return false, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("lookup existing pending resume run: %w", err)
	}

	return false, fmt.Errorf("source run %s not found for pending resume", sourceRunID)
}

// ClaimNextPending atomically claims one pending run and optionally leases a slot.
func (r *Repository) ClaimNextPending(ctx context.Context, params domainrepo.ClaimParams) (domainrepo.ClaimedRun, bool, error) {
	workerID := strings.TrimSpace(params.WorkerID)
	if workerID == "" {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("claim pending run: worker_id is required")
	}
	runLeaseTTL := params.RunLeaseTTL
	if runLeaseTTL <= 0 {
		runLeaseTTL = params.LeaseTTL
	}
	if runLeaseTTL <= 0 {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("claim pending run: run_lease_ttl is required")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("begin claim transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var (
		runID         string
		correlationID string
		projectIDRaw  pgtype.Text
		learningMode  bool
		runPayload    []byte
	)

	err = tx.QueryRow(ctx, queryClaimNextPendingForUpdate).Scan(
		&runID,
		&correlationID,
		&projectIDRaw,
		&learningMode,
		&runPayload,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.ClaimedRun{}, false, nil
		}
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("select pending run for claim: %w", err)
	}

	payload := parseRunQueuePayload(runPayload)
	projectID := projectIDRaw.String
	explicitProjectID := projectIDRaw.Valid && strings.TrimSpace(projectIDRaw.String) != ""
	if projectID == "" {
		projectID = deriveProjectID(correlationID, payload)
	}
	projectSlug, projectName := deriveProjectMeta(projectID, correlationID, payload)
	requiresSlot := requiresProjectSlot(payload)

	settingsJSON, err := json.Marshal(querytypes.ProjectSettings{LearningModeDefault: params.ProjectLearningModeDefault})
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("marshal project settings: %w", err)
	}

	if explicitProjectID {
		if _, err := tx.Exec(ctx, queryEnsureProjectExists, projectID, projectSlug, projectName, settingsJSON); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("ensure project %s exists: %w", projectID, err)
		}
	} else {
		if _, err := tx.Exec(ctx, queryUpsertProject, projectID, projectSlug, projectName, settingsJSON); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("upsert project %s: %w", projectID, err)
		}
	}

	var (
		slotID string
		slotNo int
	)
	if requiresSlot {
		projectSettingsJSON, err := r.getProjectSettingsJSON(ctx, tx, projectID)
		if err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("get project settings for project %s: %w", projectID, err)
		}
		effectiveSlotsPerProject := resolveSlotsPerProject(projectSettingsJSON, params.SlotsPerProject)
		if _, err := tx.Exec(ctx, queryEnsureProjectSlots, projectID, effectiveSlotsPerProject); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("ensure slots for project %s: %w", projectID, err)
		}
		if _, err := tx.Exec(ctx, queryReleaseExpiredSlots, projectID); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("release expired slots for project %s: %w", projectID, err)
		}

		leaseUntilInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(params.LeaseTTL.Seconds())))
		if err := tx.QueryRow(ctx, queryLeaseSlot, projectID, runID, leaseUntilInterval).Scan(&slotID, &slotNo); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domainrepo.ClaimedRun{}, false, nil
			}
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("lease slot for run %s: %w", runID, err)
		}
	}

	runLeaseInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(runLeaseTTL.Seconds())))
	res, err := tx.Exec(ctx, queryMarkRunRunning, runID, projectID, workerID, runLeaseInterval)
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("mark run %s as running: %w", runID, err)
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("mark run %s as running affected 0 rows", runID)
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("commit claim transaction: %w", err)
	}

	return domainrepo.ClaimedRun{
		RunID:         runID,
		CorrelationID: correlationID,
		ProjectID:     projectID,
		LearningMode:  learningMode,
		RunPayload:    json.RawMessage(runPayload),
		SlotNo:        slotNo,
		SlotID:        slotID,
	}, true, nil
}

// ClaimRunning atomically leases running runs for one worker reconciliation tick.
func (r *Repository) ClaimRunning(ctx context.Context, params domainrepo.ClaimRunningParams) ([]domainrepo.RunningRun, error) {
	workerID := strings.TrimSpace(params.WorkerID)
	if workerID == "" {
		return nil, fmt.Errorf("claim running runs: worker_id is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	leaseTTL := params.LeaseTTL
	if leaseTTL <= 0 {
		return nil, fmt.Errorf("claim running runs: lease_ttl is required")
	}
	leaseInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(leaseTTL.Seconds())))

	rows, err := r.db.Query(ctx, queryClaimRunning, workerID, leaseInterval, limit)
	if err != nil {
		return nil, fmt.Errorf("claim running runs: %w", err)
	}
	defer rows.Close()

	items, err := scanRunningRows(rows, limit)
	if err != nil {
		return nil, fmt.Errorf("scan claimed running runs: %w", err)
	}
	return items, nil
}

// ReleaseStaleLeases clears running-run leases owned by stale worker instances.
func (r *Repository) ReleaseStaleLeases(ctx context.Context, params domainrepo.ReleaseStaleLeasesParams) ([]domainrepo.ReleasedStaleLease, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.Query(ctx, queryReleaseStaleRunningLeases, limit, params.ReleaseMissingOwners, params.ActiveWorkerIDs)
	if err != nil {
		return nil, fmt.Errorf("release stale running leases: %w", err)
	}
	defer rows.Close()

	items := make([]domainrepo.ReleasedStaleLease, 0, limit)
	for rows.Next() {
		var (
			item                  domainrepo.ReleasedStaleLease
			previousLeaseUntilRaw pgtype.Timestamptz
			workerHeartbeatAtRaw  pgtype.Timestamptz
			workerExpiresAtRaw    pgtype.Timestamptz
		)
		if err := rows.Scan(
			&item.RunID,
			&item.CorrelationID,
			&item.ProjectID,
			&item.PreviousLeaseOwner,
			&previousLeaseUntilRaw,
			&workerHeartbeatAtRaw,
			&workerExpiresAtRaw,
			&item.WorkerStatus,
		); err != nil {
			return nil, fmt.Errorf("scan released stale running lease row: %w", err)
		}
		item.PreviousLeaseUntil = timestamptzPtr(previousLeaseUntilRaw)
		item.WorkerHeartbeatAt = timestamptzPtr(workerHeartbeatAtRaw)
		item.WorkerExpiresAt = timestamptzPtr(workerExpiresAtRaw)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate released stale running leases: %w", err)
	}
	return items, nil
}

// ReleaseOwnedLeases clears running-run leases currently owned by one worker during graceful shutdown.
func (r *Repository) ReleaseOwnedLeases(ctx context.Context, params domainrepo.ReleaseOwnedLeasesParams) ([]domainrepo.ReleasedStaleLease, error) {
	workerID := strings.TrimSpace(params.WorkerID)
	if workerID == "" {
		return nil, fmt.Errorf("release owned running leases: worker_id is required")
	}

	rows, err := r.db.Query(ctx, queryReleaseOwnedRunningLeases, workerID)
	if err != nil {
		return nil, fmt.Errorf("release owned running leases: %w", err)
	}
	defer rows.Close()

	items := make([]domainrepo.ReleasedStaleLease, 0, 8)
	for rows.Next() {
		var (
			item                  domainrepo.ReleasedStaleLease
			previousLeaseUntilRaw pgtype.Timestamptz
		)
		if err := rows.Scan(
			&item.RunID,
			&item.CorrelationID,
			&item.ProjectID,
			&item.PreviousLeaseOwner,
			&previousLeaseUntilRaw,
			&item.WorkerStatus,
		); err != nil {
			return nil, fmt.Errorf("scan released owned running lease row: %w", err)
		}
		item.PreviousLeaseUntil = timestamptzPtr(previousLeaseUntilRaw)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate released owned running leases: %w", err)
	}
	return items, nil
}

// ListRunning returns active runs for diagnostics/peer checks.
func (r *Repository) ListRunning(ctx context.Context, limit int) ([]domainrepo.RunningRun, error) {
	rows, err := r.db.Query(ctx, queryListRunning, limit)
	if err != nil {
		return nil, fmt.Errorf("list running runs: %w", err)
	}
	defer rows.Close()

	items, err := scanRunningRows(rows, limit)
	if err != nil {
		return nil, fmt.Errorf("scan running runs: %w", err)
	}
	return items, nil
}

// ListNonTerminalByRunIDs returns non-terminal runs referenced by managed namespace labels.
func (r *Repository) ListNonTerminalByRunIDs(ctx context.Context, runIDs []string) ([]domainrepo.NonTerminalRun, error) {
	normalized := normalizeRunIDs(runIDs)
	if len(normalized) == 0 {
		return nil, nil
	}

	rows, err := r.db.Query(ctx, queryListNonTerminalByRunIDs, normalized)
	if err != nil {
		return nil, fmt.Errorf("list non-terminal runs by ids: %w", err)
	}
	defer rows.Close()

	items := make([]domainrepo.NonTerminalRun, 0, len(normalized))
	for rows.Next() {
		var item domainrepo.NonTerminalRun
		if err := rows.Scan(&item.RunID, &item.Status); err != nil {
			return nil, fmt.Errorf("scan non-terminal run row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate non-terminal runs by ids: %w", err)
	}
	return items, nil
}

// ExtendLease refreshes slot lease ownership for one running run.
func (r *Repository) ExtendLease(ctx context.Context, params domainrepo.ExtendLeaseParams) (bool, error) {
	projectID := strings.TrimSpace(params.ProjectID)
	runID := strings.TrimSpace(params.RunID)
	if projectID == "" || runID == "" {
		return false, nil
	}

	leaseUntilInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(params.LeaseTTL.Seconds())))
	res, err := r.db.Exec(ctx, queryExtendSlotLease, projectID, runID, leaseUntilInterval)
	if err != nil {
		return false, fmt.Errorf("extend slot lease for run %s: %w", runID, err)
	}

	return res.RowsAffected() > 0, nil
}

func normalizeRunIDs(runIDs []string) []string {
	if len(runIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(runIDs))
	result := make([]string, 0, len(runIDs))
	for _, raw := range runIDs {
		runID := strings.TrimSpace(raw)
		if runID == "" {
			continue
		}
		if _, ok := seen[runID]; ok {
			continue
		}
		seen[runID] = struct{}{}
		result = append(result, runID)
	}
	return result
}

// FinishRun sets final status and releases leased slot.
func (r *Repository) FinishRun(ctx context.Context, params domainrepo.FinishParams) (bool, error) {
	if params.Status != rundomain.StatusSucceeded && params.Status != rundomain.StatusFailed && params.Status != rundomain.StatusCanceled {
		return false, fmt.Errorf("unsupported final run status %q", params.Status)
	}
	leaseOwner := strings.TrimSpace(params.LeaseOwner)
	if leaseOwner == "" {
		return false, fmt.Errorf("finish run %s: lease_owner is required", params.RunID)
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin finish transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res, err := tx.Exec(ctx, queryMarkRunFinished, params.RunID, string(params.Status), params.FinishedAt.UTC(), leaseOwner)
	if err != nil {
		return false, fmt.Errorf("mark run %s as %s: %w", params.RunID, params.Status, err)
	}
	rows := res.RowsAffected()
	if rows == 0 {
		return false, nil
	}

	if _, err := tx.Exec(ctx, queryMarkSlotReleasing, params.ProjectID, params.RunID); err != nil {
		return false, fmt.Errorf("mark slot releasing for run %s: %w", params.RunID, err)
	}
	if _, err := tx.Exec(ctx, queryMarkSlotFree, params.ProjectID, params.RunID); err != nil {
		return false, fmt.Errorf("mark slot free for run %s: %w", params.RunID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit finish transaction: %w", err)
	}

	return true, nil
}

func scanRunningRows(rows pgx.Rows, limit int) ([]domainrepo.RunningRun, error) {
	if limit <= 0 {
		limit = 100
	}
	result := make([]domainrepo.RunningRun, 0, limit)
	for rows.Next() {
		var (
			runID                    string
			correlationID            string
			projectID                string
			slotID                   string
			slotNo                   int
			learningMode             bool
			runPayload               []byte
			startedAt                pgtype.Timestamptz
			reclaimedAfterStaleLease bool
		)
		if err := rows.Scan(&runID, &correlationID, &projectID, &slotID, &slotNo, &learningMode, &runPayload, &startedAt, &reclaimedAfterStaleLease); err != nil {
			return nil, fmt.Errorf("scan running run row: %w", err)
		}
		item := domainrepo.RunningRun{
			RunID:                    runID,
			CorrelationID:            correlationID,
			ProjectID:                projectID,
			SlotID:                   slotID,
			SlotNo:                   slotNo,
			LearningMode:             learningMode,
			RunPayload:               json.RawMessage(runPayload),
			ReclaimedAfterStaleLease: reclaimedAfterStaleLease,
		}
		if startedAt.Valid {
			item.StartedAt = startedAt.Time.UTC()
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate running runs: %w", err)
	}
	return result, nil
}

func timestamptzPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

// parseRunQueuePayload unmarshals only fields required by runqueue repository logic.
func parseRunQueuePayload(raw []byte) querytypes.RunQueuePayload {
	if len(raw) == 0 {
		return querytypes.RunQueuePayload{}
	}

	var payload querytypes.RunQueuePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return querytypes.RunQueuePayload{}
	}

	return payload
}

func (r *Repository) getProjectSettingsJSON(ctx context.Context, tx pgx.Tx, projectID string) ([]byte, error) {
	var settingsRaw []byte
	err := tx.QueryRow(ctx, queryGetProjectSettings, projectID).Scan(&settingsRaw)
	if err == nil {
		return settingsRaw, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return nil, err
}

func resolveSlotsPerProject(projectSettingsJSON []byte, fallback int) int {
	if fallback <= 0 {
		fallback = 1
	}
	if len(projectSettingsJSON) == 0 {
		return fallback
	}

	var settings querytypes.ProjectSettings
	if err := json.Unmarshal(projectSettingsJSON, &settings); err != nil {
		return fallback
	}
	if settings.SlotsPerProject > 0 {
		return settings.SlotsPerProject
	}

	return fallback
}

func isDeployOnlyRun(payload querytypes.RunQueuePayload) bool {
	return payload.Runtime != nil && payload.Runtime.DeployOnly
}

func requiresProjectSlot(payload querytypes.RunQueuePayload) bool {
	if isDeployOnlyRun(payload) {
		return false
	}
	if payload.Runtime == nil {
		return true
	}
	return agentdomain.ParseRuntimeMode(payload.Runtime.Mode) != agentdomain.RuntimeModeCodeOnly
}

// deriveProjectID prefers repository identity and falls back to correlation-scoped synthetic id.
func deriveProjectID(correlationID string, payload querytypes.RunQueuePayload) string {
	if payload.Repository.FullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(payload.Repository.FullName))).String()
	}

	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}

// deriveProjectMeta builds stable project slug/name values from payload or synthetic fallback.
func deriveProjectMeta(projectID string, correlationID string, payload querytypes.RunQueuePayload) (slug string, name string) {
	if payload.Repository.FullName != "" {
		slug = strings.ToLower(strings.TrimSpace(payload.Repository.FullName))
		name = slug
		if strings.TrimSpace(payload.Repository.Name) != "" {
			// Preserve full_name as stable display name; repo name alone is not unique.
			name = slug
		}
		return slug, name
	}

	// Fallback for synthetic/unknown correlation-driven projects.
	slug = "project-" + strings.ToLower(strings.ReplaceAll(projectID, "-", ""))[:8]
	name = slug
	if correlationID != "" {
		name = "project-" + correlationID
	}
	return slug, name
}

// maxInt64 returns the greater of two int64 values.
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
