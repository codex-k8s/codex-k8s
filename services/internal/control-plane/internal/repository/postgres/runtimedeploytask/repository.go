package runtimedeploytask

import (
	"context"
	"encoding/json"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_by_run_id.sql
	queryGetByRunID string
	//go:embed sql/select_by_run_id_for_update.sql
	querySelectByRunIDForUpdate string
	//go:embed sql/insert_pending.sql
	queryInsertPending string
	//go:embed sql/reset_desired_to_pending.sql
	queryResetDesiredToPending string
	//go:embed sql/claim_next.sql
	queryClaimNext string
	//go:embed sql/mark_succeeded.sql
	queryMarkSucceeded string
	//go:embed sql/mark_failed.sql
	queryMarkFailed string
	//go:embed sql/renew_lease.sql
	queryRenewLease string
	//go:embed sql/list_recent.sql
	queryListRecent string
	//go:embed sql/append_log.sql
	queryAppendLog string
)

// Repository persists runtime_deploy_tasks state in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL runtime_deploy_tasks repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// UpsertDesired creates or updates one run-bound desired deployment state.
func (r *Repository) UpsertDesired(ctx context.Context, params domainrepo.UpsertDesiredParams) (domainrepo.Task, error) {
	normalized, err := normalizeUpsertParams(params)
	if err != nil {
		return domainrepo.Task{}, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Task{}, fmt.Errorf("begin runtime deploy upsert transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	existing, found, err := getByRunIDForUpdate(ctx, tx, normalized.RunID)
	if err != nil {
		return domainrepo.Task{}, err
	}
	if !found {
		inserted, insertErr := insertPending(ctx, tx, normalized)
		if insertErr != nil {
			return domainrepo.Task{}, insertErr
		}
		if err := tx.Commit(ctx); err != nil {
			return domainrepo.Task{}, fmt.Errorf("commit runtime deploy upsert transaction: %w", err)
		}
		return inserted, nil
	}

	if !shouldResetDesired(existing, normalized) {
		if err := tx.Commit(ctx); err != nil {
			return domainrepo.Task{}, fmt.Errorf("commit runtime deploy upsert transaction: %w", err)
		}
		return existing, nil
	}

	updated, err := resetDesiredToPending(ctx, tx, normalized)
	if err != nil {
		return domainrepo.Task{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Task{}, fmt.Errorf("commit runtime deploy upsert transaction: %w", err)
	}
	return updated, nil
}

// GetByRunID returns one runtime deploy task by run id.
func (r *Repository) GetByRunID(ctx context.Context, runID string) (domainrepo.Task, bool, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return domainrepo.Task{}, false, nil
	}
	row := r.db.QueryRow(ctx, queryGetByRunID, runID)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Task{}, false, nil
		}
		return domainrepo.Task{}, false, fmt.Errorf("query runtime deploy task by run_id=%s: %w", runID, err)
	}
	return task, true, nil
}

// ClaimNext acquires one pending/expired-running task lease.
func (r *Repository) ClaimNext(ctx context.Context, params domainrepo.ClaimParams) (domainrepo.Task, bool, error) {
	leaseOwner := strings.TrimSpace(params.LeaseOwner)
	leaseTTL := strings.TrimSpace(params.LeaseTTL)
	if leaseOwner == "" {
		return domainrepo.Task{}, false, fmt.Errorf("claim runtime deploy task: lease_owner is required")
	}
	if leaseTTL == "" {
		return domainrepo.Task{}, false, fmt.Errorf("claim runtime deploy task: lease_ttl is required")
	}

	row := r.db.QueryRow(ctx, queryClaimNext, leaseOwner, leaseTTL)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Task{}, false, nil
		}
		return domainrepo.Task{}, false, fmt.Errorf("claim runtime deploy task: %w", err)
	}
	return task, true, nil
}

// MarkSucceeded sets successful terminal state for one leased task.
func (r *Repository) MarkSucceeded(ctx context.Context, params domainrepo.MarkSucceededParams) (bool, error) {
	runID := strings.TrimSpace(params.RunID)
	leaseOwner := strings.TrimSpace(params.LeaseOwner)
	if runID == "" {
		return false, fmt.Errorf("mark runtime deploy task succeeded: run_id is required")
	}
	if leaseOwner == "" {
		return false, fmt.Errorf("mark runtime deploy task succeeded: lease_owner is required")
	}

	var returnedRunID string
	err := r.db.QueryRow(ctx, queryMarkSucceeded, runID, leaseOwner, strings.TrimSpace(params.ResultNamespace), strings.TrimSpace(params.ResultTargetEnv)).Scan(&returnedRunID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("mark runtime deploy task %s succeeded: %w", runID, err)
	}
	return strings.TrimSpace(returnedRunID) != "", nil
}

// MarkFailed sets failed terminal state for one leased task.
func (r *Repository) MarkFailed(ctx context.Context, params domainrepo.MarkFailedParams) (bool, error) {
	runID := strings.TrimSpace(params.RunID)
	leaseOwner := strings.TrimSpace(params.LeaseOwner)
	if runID == "" {
		return false, fmt.Errorf("mark runtime deploy task failed: run_id is required")
	}
	if leaseOwner == "" {
		return false, fmt.Errorf("mark runtime deploy task failed: lease_owner is required")
	}

	var returnedRunID string
	err := r.db.QueryRow(ctx, queryMarkFailed, runID, leaseOwner, strings.TrimSpace(params.LastError)).Scan(&returnedRunID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("mark runtime deploy task %s failed: %w", runID, err)
	}
	return strings.TrimSpace(returnedRunID) != "", nil
}

// RenewLease extends running task lease for current owner.
func (r *Repository) RenewLease(ctx context.Context, params domainrepo.RenewLeaseParams) (bool, error) {
	runID := strings.TrimSpace(params.RunID)
	leaseOwner := strings.TrimSpace(params.LeaseOwner)
	leaseTTL := strings.TrimSpace(params.LeaseTTL)
	if runID == "" {
		return false, fmt.Errorf("renew runtime deploy task lease: run_id is required")
	}
	if leaseOwner == "" {
		return false, fmt.Errorf("renew runtime deploy task lease: lease_owner is required")
	}
	if leaseTTL == "" {
		return false, fmt.Errorf("renew runtime deploy task lease: lease_ttl is required")
	}

	var returnedRunID string
	err := r.db.QueryRow(ctx, queryRenewLease, runID, leaseOwner, leaseTTL).Scan(&returnedRunID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("renew runtime deploy task lease for run %s: %w", runID, err)
	}
	return strings.TrimSpace(returnedRunID) != "", nil
}

// ListRecent returns runtime deploy tasks ordered by updated_at desc.
func (r *Repository) ListRecent(ctx context.Context, filter domainrepo.ListFilter) ([]domainrepo.Task, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	rows, err := r.db.Query(
		ctx,
		queryListRecent,
		strings.TrimSpace(filter.Status),
		strings.TrimSpace(filter.TargetEnv),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list runtime deploy tasks: %w", err)
	}
	defer rows.Close()

	items := make([]domainrepo.Task, 0, limit)
	for rows.Next() {
		item, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan runtime deploy task list item: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runtime deploy task list rows: %w", err)
	}
	return items, nil
}

// AppendLog appends one task log line.
func (r *Repository) AppendLog(ctx context.Context, params domainrepo.AppendLogParams) error {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return fmt.Errorf("append runtime deploy task log: run_id is required")
	}
	stage := strings.TrimSpace(params.Stage)
	if stage == "" {
		stage = "deploy"
	}
	level := strings.TrimSpace(params.Level)
	if level == "" {
		level = "info"
	}
	message := strings.TrimSpace(params.Message)
	if message == "" {
		return nil
	}
	maxLines := params.MaxLines
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 5000 {
		maxLines = 5000
	}

	tag, err := r.db.Exec(ctx, queryAppendLog, runID, stage, level, message, maxLines)
	if err != nil {
		return fmt.Errorf("append runtime deploy task log for run %s: %w", runID, err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}

type taskRowScanner interface {
	Scan(dest ...any) error
}

func getByRunIDForUpdate(ctx context.Context, tx pgx.Tx, runID string) (domainrepo.Task, bool, error) {
	row := tx.QueryRow(ctx, querySelectByRunIDForUpdate, runID)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Task{}, false, nil
		}
		return domainrepo.Task{}, false, fmt.Errorf("select runtime deploy task run_id=%s for update: %w", runID, err)
	}
	return task, true, nil
}

func insertPending(ctx context.Context, tx pgx.Tx, params domainrepo.UpsertDesiredParams) (domainrepo.Task, error) {
	return applyDesiredStateMutation(ctx, tx, queryInsertPending, params, "insert runtime deploy task")
}

func resetDesiredToPending(ctx context.Context, tx pgx.Tx, params domainrepo.UpsertDesiredParams) (domainrepo.Task, error) {
	return applyDesiredStateMutation(ctx, tx, queryResetDesiredToPending, params, "reset runtime deploy task to pending")
}

func applyDesiredStateMutation(ctx context.Context, tx pgx.Tx, sqlQuery string, params domainrepo.UpsertDesiredParams, action string) (domainrepo.Task, error) {
	row := tx.QueryRow(
		ctx,
		sqlQuery,
		params.RunID,
		params.RuntimeMode,
		params.Namespace,
		params.TargetEnv,
		params.SlotNo,
		params.RepositoryFullName,
		params.ServicesYAMLPath,
		params.BuildRef,
		params.DeployOnly,
	)
	task, err := scanTask(row)
	if err != nil {
		return domainrepo.Task{}, fmt.Errorf("%s run_id=%s: %w", action, params.RunID, err)
	}
	return task, nil
}

func scanTask(row taskRowScanner) (domainrepo.Task, error) {
	var (
		task            domainrepo.Task
		statusRaw       string
		leaseUntil      pgtype.Timestamptz
		createdAt       time.Time
		updatedAt       time.Time
		startedAt       pgtype.Timestamptz
		finishedAt      pgtype.Timestamptz
		logsRaw         []byte
		leaseOwner      string
		lastError       string
		resultNamespace string
		resultTargetEnv string
	)

	err := row.Scan(
		&task.RunID,
		&task.RuntimeMode,
		&task.Namespace,
		&task.TargetEnv,
		&task.SlotNo,
		&task.RepositoryFullName,
		&task.ServicesYAMLPath,
		&task.BuildRef,
		&task.DeployOnly,
		&statusRaw,
		&leaseOwner,
		&leaseUntil,
		&task.Attempts,
		&lastError,
		&resultNamespace,
		&resultTargetEnv,
		&createdAt,
		&updatedAt,
		&startedAt,
		&finishedAt,
		&logsRaw,
	)
	if err != nil {
		return domainrepo.Task{}, err
	}

	status, err := parseRuntimeDeployStatus(statusRaw)
	if err != nil {
		return domainrepo.Task{}, err
	}
	task.Status = status
	task.LeaseOwner = strings.TrimSpace(leaseOwner)
	if leaseUntil.Valid {
		task.LeaseUntil = leaseUntil.Time.UTC()
	}
	task.LastError = strings.TrimSpace(lastError)
	task.ResultNamespace = strings.TrimSpace(resultNamespace)
	task.ResultTargetEnv = strings.TrimSpace(resultTargetEnv)
	task.CreatedAt = createdAt.UTC()
	task.UpdatedAt = updatedAt.UTC()
	if startedAt.Valid {
		task.StartedAt = startedAt.Time.UTC()
	}
	if finishedAt.Valid {
		task.FinishedAt = finishedAt.Time.UTC()
	}
	task.Logs = parseTaskLogs(logsRaw)

	return task, nil
}

func parseTaskLogs(raw []byte) []entitytypes.RuntimeDeployTaskLogEntry {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return []entitytypes.RuntimeDeployTaskLogEntry{}
	}
	type dbLogEntry struct {
		Stage     string    `json:"stage"`
		Level     string    `json:"level"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"created_at"`
	}
	parsed := make([]dbLogEntry, 0)
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return []entitytypes.RuntimeDeployTaskLogEntry{}
	}
	out := make([]entitytypes.RuntimeDeployTaskLogEntry, 0, len(parsed))
	for _, entry := range parsed {
		out = append(out, entitytypes.RuntimeDeployTaskLogEntry{
			Stage:     strings.TrimSpace(entry.Stage),
			Level:     strings.TrimSpace(entry.Level),
			Message:   strings.TrimSpace(entry.Message),
			CreatedAt: entry.CreatedAt.UTC(),
		})
	}
	return out
}

func parseRuntimeDeployStatus(raw string) (entitytypes.RuntimeDeployTaskStatus, error) {
	status := entitytypes.RuntimeDeployTaskStatus(strings.TrimSpace(raw))
	switch status {
	case entitytypes.RuntimeDeployTaskStatusPending,
		entitytypes.RuntimeDeployTaskStatusRunning,
		entitytypes.RuntimeDeployTaskStatusSucceeded,
		entitytypes.RuntimeDeployTaskStatusFailed:
		return status, nil
	default:
		return "", fmt.Errorf("unknown runtime deploy task status %q", raw)
	}
}

func normalizeUpsertParams(params domainrepo.UpsertDesiredParams) (domainrepo.UpsertDesiredParams, error) {
	params.RunID = strings.TrimSpace(params.RunID)
	if params.RunID == "" {
		return domainrepo.UpsertDesiredParams{}, fmt.Errorf("upsert runtime deploy task: run_id is required")
	}
	params.RuntimeMode = strings.TrimSpace(params.RuntimeMode)
	if params.RuntimeMode == "" {
		params.RuntimeMode = "full-env"
	}
	params.Namespace = strings.TrimSpace(params.Namespace)
	params.TargetEnv = strings.TrimSpace(params.TargetEnv)
	if params.TargetEnv == "" {
		params.TargetEnv = "ai"
	}
	if params.SlotNo < 0 {
		params.SlotNo = 0
	}
	params.RepositoryFullName = strings.TrimSpace(params.RepositoryFullName)
	params.ServicesYAMLPath = strings.TrimSpace(params.ServicesYAMLPath)
	params.BuildRef = strings.TrimSpace(params.BuildRef)
	return params, nil
}

func shouldResetDesired(existing domainrepo.Task, params domainrepo.UpsertDesiredParams) bool {
	if existing.Status == entitytypes.RuntimeDeployTaskStatusFailed {
		return true
	}
	return !sameDesired(existing, params)
}

func sameDesired(existing domainrepo.Task, params domainrepo.UpsertDesiredParams) bool {
	if strings.TrimSpace(existing.RuntimeMode) != strings.TrimSpace(params.RuntimeMode) {
		return false
	}
	if strings.TrimSpace(existing.Namespace) != strings.TrimSpace(params.Namespace) {
		return false
	}
	if strings.TrimSpace(existing.TargetEnv) != strings.TrimSpace(params.TargetEnv) {
		return false
	}
	if existing.SlotNo != params.SlotNo {
		return false
	}
	if strings.TrimSpace(existing.RepositoryFullName) != strings.TrimSpace(params.RepositoryFullName) {
		return false
	}
	if strings.TrimSpace(existing.ServicesYAMLPath) != strings.TrimSpace(params.ServicesYAMLPath) {
		return false
	}
	if strings.TrimSpace(existing.BuildRef) != strings.TrimSpace(params.BuildRef) {
		return false
	}
	if existing.DeployOnly != params.DeployOnly {
		return false
	}
	return true
}
