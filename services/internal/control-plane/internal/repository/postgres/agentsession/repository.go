package agentsession

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/upsert.sql
	queryUpsert string
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
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) error {
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "running"
	}

	issueNumber := pgtype.Int8{}
	if params.IssueNumber != nil {
		issueNumber = pgtype.Int8{Int64: int64(*params.IssueNumber), Valid: true}
	}

	prNumber := pgtype.Int8{}
	if params.PRNumber != nil {
		prNumber = pgtype.Int8{Int64: int64(*params.PRNumber), Valid: true}
	}

	finishedAt := pgtype.Timestamptz{}
	if params.FinishedAt != nil {
		finishedAt = pgtype.Timestamptz{Time: params.FinishedAt.UTC(), Valid: true}
	}

	startedAt := pgtype.Timestamptz{}
	if !params.StartedAt.IsZero() {
		startedAt = pgtype.Timestamptz{Time: params.StartedAt.UTC(), Valid: true}
	}

	var sessionJSON []byte
	if len(params.SessionJSON) > 0 {
		sessionJSON = []byte(params.SessionJSON)
	}

	var codexJSON []byte
	if len(params.CodexSessionJSON) > 0 {
		codexJSON = []byte(params.CodexSessionJSON)
	}

	if _, err := r.db.Exec(
		ctx,
		queryUpsert,
		params.RunID,
		params.CorrelationID,
		strings.TrimSpace(params.ProjectID),
		strings.TrimSpace(params.RepositoryFullName),
		strings.TrimSpace(params.AgentKey),
		issueNumber,
		strings.TrimSpace(params.BranchName),
		prNumber,
		strings.TrimSpace(params.PRURL),
		strings.TrimSpace(params.TriggerKind),
		strings.TrimSpace(params.TemplateKind),
		strings.TrimSpace(params.TemplateSource),
		strings.TrimSpace(params.TemplateLocale),
		strings.TrimSpace(params.Model),
		strings.TrimSpace(params.ReasoningEffort),
		status,
		strings.TrimSpace(params.SessionID),
		sessionJSON,
		strings.TrimSpace(params.CodexSessionPath),
		codexJSON,
		startedAt,
		finishedAt,
	); err != nil {
		return fmt.Errorf("upsert agent session: %w", err)
	}

	return nil
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

func (r *Repository) queryOneSession(ctx context.Context, query string, operationLabel string, args ...any) (domainrepo.Session, bool, error) {
	var (
		item       domainrepo.Session
		projectID  pgtype.Text
		issueNum   pgtype.Int8
		prNum      pgtype.Int8
		prURL      pgtype.Text
		trigger    pgtype.Text
		tplKind    pgtype.Text
		tplSource  pgtype.Text
		tplLocale  pgtype.Text
		model      pgtype.Text
		reasoning  pgtype.Text
		waitState  pgtype.Text
		heartbeat  pgtype.Timestamptz
		sessionID  pgtype.Text
		sessionRaw []byte
		path       pgtype.Text
		codexRaw   []byte
		guardOff   bool
		startedAt  pgtype.Timestamptz
		finishedAt pgtype.Timestamptz
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
