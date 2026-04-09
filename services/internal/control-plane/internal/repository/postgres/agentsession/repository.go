package agentsession

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentsession"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/insert_if_absent.sql
	queryInsertIfAbsent string
	//go:embed sql/update_if_version_matches.sql
	queryUpdateIfVersionMatches string
	//go:embed sql/get_by_run_id.sql
	queryGetByRunID string
	//go:embed sql/get_latest_by_repository_branch_and_agent.sql
	queryGetLatestByRepositoryBranchAndAgent string
	//go:embed sql/set_wait_state_by_run_id.sql
	querySetWaitStateByRunID string
	//go:embed sql/cleanup_session_payloads_finished_before.sql
	queryCleanupSessionPayloadsFinishedBefore string
)

// Repository stores resumable agent sessions in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL agent session repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Upsert stores or updates run session snapshot by run_id.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.UpsertResult, error) {
	if params.ExpectedSnapshotVersion < 0 {
		return domainrepo.UpsertResult{}, fmt.Errorf("upsert agent session: snapshot_version must be >= 0")
	}

	record, err := buildUpsertRecord(normalizeUpsertParams(params), nil)
	if err != nil {
		return domainrepo.UpsertResult{}, fmt.Errorf("upsert agent session: build record: %w", err)
	}

	if record.ExpectedVersion == 0 {
		inserted, result, err := r.insertIfAbsent(ctx, record)
		if err != nil {
			return domainrepo.UpsertResult{}, err
		}
		if inserted {
			return result, nil
		}

		current, found, err := r.GetByRunID(ctx, record.RunID)
		if err != nil {
			return domainrepo.UpsertResult{}, err
		}
		if found && isIdempotentReplay(current, record.ExpectedVersion, record.SnapshotChecksum) {
			return snapshotStateFromSession(current), nil
		}
		return domainrepo.UpsertResult{}, domainrepo.SnapshotVersionConflict{
			ExpectedSnapshotVersion: record.ExpectedVersion,
			ActualSnapshotVersion:   current.SnapshotVersion,
		}
	}

	current, found, err := r.GetByRunID(ctx, record.RunID)
	if err != nil {
		return domainrepo.UpsertResult{}, err
	}
	if !found {
		return domainrepo.UpsertResult{}, domainrepo.SnapshotVersionConflict{
			ExpectedSnapshotVersion: record.ExpectedVersion,
			ActualSnapshotVersion:   0,
		}
	}

	record, err = buildUpsertRecord(normalizeUpsertParams(params), &current)
	if err != nil {
		return domainrepo.UpsertResult{}, fmt.Errorf("upsert agent session: build merged record: %w", err)
	}

	updated, result, err := r.updateIfVersionMatches(ctx, record)
	if err != nil {
		return domainrepo.UpsertResult{}, err
	}
	if updated {
		return result, nil
	}

	current, found, err = r.GetByRunID(ctx, record.RunID)
	if err != nil {
		return domainrepo.UpsertResult{}, err
	}
	if found && isIdempotentReplay(current, record.ExpectedVersion, record.SnapshotChecksum) {
		return snapshotStateFromSession(current), nil
	}
	actualVersion := int64(0)
	if found {
		actualVersion = current.SnapshotVersion
	}
	return domainrepo.UpsertResult{}, domainrepo.SnapshotVersionConflict{
		ExpectedSnapshotVersion: record.ExpectedVersion,
		ActualSnapshotVersion:   actualVersion,
	}
}

// SetWaitStateByRunID updates wait-state and timeout guard fields for run session.
func (r *Repository) SetWaitStateByRunID(ctx context.Context, params domainrepo.SetWaitStateParams) (bool, error) {
	lastHeartbeatAt := pgtype.Timestamptz{}
	if params.LastHeartbeatAt != nil {
		lastHeartbeatAt = pgtype.Timestamptz{Time: params.LastHeartbeatAt.UTC(), Valid: true}
	}

	waitState := nullableTrimmedText(params.WaitState)
	res, err := r.db.Exec(
		ctx,
		querySetWaitStateByRunID,
		strings.TrimSpace(params.RunID),
		waitState,
		params.TimeoutGuardDisabled,
		lastHeartbeatAt,
	)
	if err != nil {
		return false, fmt.Errorf("set wait state by run id: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

// GetByRunID returns latest session snapshot for one run id.
func (r *Repository) GetByRunID(ctx context.Context, runID string) (domainrepo.Session, bool, error) {
	return r.queryOneSession(
		ctx,
		queryGetByRunID,
		"run id",
		strings.TrimSpace(runID),
	)
}

// GetLatestByRepositoryBranchAndAgent returns latest snapshot by repository + branch + agent key.
func (r *Repository) GetLatestByRepositoryBranchAndAgent(ctx context.Context, repositoryFullName string, branchName string, agentKey string) (domainrepo.Session, bool, error) {
	return r.queryOneSession(
		ctx,
		queryGetLatestByRepositoryBranchAndAgent,
		"repository+branch+agent",
		strings.TrimSpace(repositoryFullName),
		strings.TrimSpace(branchName),
		strings.TrimSpace(agentKey),
	)
}

// CleanupSessionPayloadsFinishedBefore clears heavy session payloads for finished runs older than cutoff.
func (r *Repository) CleanupSessionPayloadsFinishedBefore(ctx context.Context, finishedBefore time.Time) (int64, error) {
	cutoff := finishedBefore.UTC()
	if cutoff.IsZero() {
		return 0, fmt.Errorf("finished_before is required")
	}

	res, err := r.db.Exec(ctx, queryCleanupSessionPayloadsFinishedBefore, cutoff)
	if err != nil {
		return 0, fmt.Errorf("cleanup agent session payloads before %s: %w", cutoff.Format(time.RFC3339), err)
	}
	affected := res.RowsAffected()
	return affected, nil
}

func (r *Repository) insertIfAbsent(ctx context.Context, record upsertRecord) (bool, domainrepo.UpsertResult, error) {
	row := r.db.QueryRow(
		ctx,
		queryInsertIfAbsent,
		record.RunID,
		record.CorrelationID,
		nullableTrimmedUUID(record.ProjectID),
		record.RepositoryFullName,
		record.AgentKey,
		intPtrToPGType(record.IssueNumber),
		nullableTrimmedText(record.BranchName),
		intPtrToPGType(record.PRNumber),
		nullableTrimmedText(record.PRURL),
		nullableTrimmedText(record.TriggerKind),
		nullableTrimmedText(record.TemplateKind),
		nullableTrimmedText(record.TemplateSource),
		nullableTrimmedText(record.TemplateLocale),
		nullableTrimmedText(record.Model),
		nullableTrimmedText(record.ReasoningEffort),
		record.Status,
		nullableTrimmedText(record.SessionID),
		record.SessionJSON,
		nullableTrimmedText(record.CodexSessionPath),
		bytesOrNil(record.CodexSessionJSON),
		timestamptzOrNil(record.StartedAt),
		timestamptzPtrOrNil(record.FinishedAt),
		record.SnapshotChecksum,
		record.SnapshotUpdatedAt,
	)

	result, found, err := scanUpsertResult(row)
	if err != nil {
		return false, domainrepo.UpsertResult{}, fmt.Errorf("insert agent session if absent: %w", err)
	}
	return found, result, nil
}

func (r *Repository) updateIfVersionMatches(ctx context.Context, record upsertRecord) (bool, domainrepo.UpsertResult, error) {
	row := r.db.QueryRow(
		ctx,
		queryUpdateIfVersionMatches,
		record.RunID,
		record.CorrelationID,
		nullableTrimmedUUID(record.ProjectID),
		record.RepositoryFullName,
		record.AgentKey,
		intPtrToPGType(record.IssueNumber),
		nullableTrimmedText(record.BranchName),
		intPtrToPGType(record.PRNumber),
		nullableTrimmedText(record.PRURL),
		nullableTrimmedText(record.TriggerKind),
		nullableTrimmedText(record.TemplateKind),
		nullableTrimmedText(record.TemplateSource),
		nullableTrimmedText(record.TemplateLocale),
		nullableTrimmedText(record.Model),
		nullableTrimmedText(record.ReasoningEffort),
		record.Status,
		nullableTrimmedText(record.SessionID),
		record.SessionJSON,
		nullableTrimmedText(record.CodexSessionPath),
		bytesOrNil(record.CodexSessionJSON),
		timestamptzOrNil(record.StartedAt),
		timestamptzPtrOrNil(record.FinishedAt),
		record.ExpectedVersion,
		record.SnapshotChecksum,
		record.SnapshotUpdatedAt,
	)

	result, found, err := scanUpsertResult(row)
	if err != nil {
		return false, domainrepo.UpsertResult{}, fmt.Errorf("update agent session by snapshot version: %w", err)
	}
	return found, result, nil
}

func (r *Repository) queryOneSession(ctx context.Context, query string, operationLabel string, args ...any) (domainrepo.Session, bool, error) {
	var (
		item              domainrepo.Session
		projectID         pgtype.Text
		issueNum          pgtype.Int8
		prNum             pgtype.Int8
		prURL             pgtype.Text
		trigger           pgtype.Text
		tplKind           pgtype.Text
		tplSource         pgtype.Text
		tplLocale         pgtype.Text
		model             pgtype.Text
		reasoning         pgtype.Text
		waitState         pgtype.Text
		heartbeat         pgtype.Timestamptz
		sessionID         pgtype.Text
		sessionRaw        []byte
		path              pgtype.Text
		codexRaw          []byte
		guardOff          bool
		startedAt         pgtype.Timestamptz
		finishedAt        pgtype.Timestamptz
		snapshotChecksum  pgtype.Text
		snapshotUpdatedAt pgtype.Timestamptz
	)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&item.ID,
		&item.RunID,
		&item.CorrelationID,
		&projectID,
		&item.RepositoryFullName,
		&item.AgentKey,
		&issueNum,
		&item.BranchName,
		&prNum,
		&prURL,
		&trigger,
		&tplKind,
		&tplSource,
		&tplLocale,
		&model,
		&reasoning,
		&item.Status,
		&waitState,
		&guardOff,
		&heartbeat,
		&sessionID,
		&sessionRaw,
		&path,
		&codexRaw,
		&item.SnapshotVersion,
		&snapshotChecksum,
		&snapshotUpdatedAt,
		&startedAt,
		&finishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Session{}, false, nil
		}
		return domainrepo.Session{}, false, fmt.Errorf("get latest agent session by %s: %w", operationLabel, err)
	}

	if projectID.Valid {
		item.ProjectID = projectID.String
	}
	if issueNum.Valid {
		item.IssueNumber = int(issueNum.Int64)
	}
	if prNum.Valid {
		item.PRNumber = int(prNum.Int64)
	}
	if prURL.Valid {
		item.PRURL = prURL.String
	}
	if trigger.Valid {
		item.TriggerKind = trigger.String
	}
	if tplKind.Valid {
		item.TemplateKind = tplKind.String
	}
	if tplSource.Valid {
		item.TemplateSource = tplSource.String
	}
	if tplLocale.Valid {
		item.TemplateLocale = tplLocale.String
	}
	if model.Valid {
		item.Model = model.String
	}
	if reasoning.Valid {
		item.ReasoningEffort = reasoning.String
	}
	if waitState.Valid {
		item.WaitState = waitState.String
	}
	item.TimeoutGuardDisabled = guardOff
	if heartbeat.Valid {
		item.LastHeartbeatAt = heartbeat.Time.UTC()
	}
	if sessionID.Valid {
		item.SessionID = sessionID.String
	}
	item.SessionJSON = json.RawMessage(sessionRaw)
	if path.Valid {
		item.CodexSessionPath = path.String
	}
	if len(codexRaw) > 0 {
		item.CodexSessionJSON = json.RawMessage(codexRaw)
	}
	if snapshotChecksum.Valid {
		item.SnapshotChecksum = snapshotChecksum.String
	}
	if snapshotUpdatedAt.Valid {
		item.SnapshotUpdatedAt = snapshotUpdatedAt.Time.UTC()
	}
	if startedAt.Valid {
		item.StartedAt = startedAt.Time.UTC()
	}
	if finishedAt.Valid {
		item.FinishedAt = finishedAt.Time.UTC()
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	return item, true, nil
}

func nullableTrimmedText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableTrimmedUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func intPtrToPGType(value *int) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*value), Valid: true}
}

func bytesOrNil(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func timestamptzOrNil(value time.Time) pgtype.Timestamptz {
	if value.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timestamptzPtrOrNil(value *time.Time) pgtype.Timestamptz {
	if value == nil || value.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func scanUpsertResult(row pgx.Row) (domainrepo.UpsertResult, bool, error) {
	var (
		result domainrepo.UpsertResult
	)
	err := row.Scan(&result.SnapshotVersion, &result.SnapshotChecksum, &result.SnapshotUpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.UpsertResult{}, false, nil
		}
		return domainrepo.UpsertResult{}, false, err
	}
	result.SnapshotUpdatedAt = result.SnapshotUpdatedAt.UTC()
	return result, true, nil
}

func normalizeUpsertParams(params domainrepo.UpsertParams) domainrepo.UpsertParams {
	if strings.TrimSpace(params.Status) == "" {
		params.Status = "running"
	}
	return params
}
