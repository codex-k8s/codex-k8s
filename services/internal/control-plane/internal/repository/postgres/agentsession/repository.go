package agentsession

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
)

var (
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/get_latest_by_repository_branch.sql
	queryGetLatestByRepositoryBranch string
)

// Repository stores resumable agent sessions in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL agent session repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Upsert stores or updates run session snapshot by run_id.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) error {
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "running"
	}

	var (
		issueNumber any
		prNumber    any
		finishedAt  any
		startedAt   any
		sessionJSON any
		codexJSON   any
	)

	if params.IssueNumber != nil {
		issueNumber = *params.IssueNumber
	}
	if params.PRNumber != nil {
		prNumber = *params.PRNumber
	}
	if params.FinishedAt != nil {
		finishedAt = params.FinishedAt.UTC()
	}
	if !params.StartedAt.IsZero() {
		startedAt = params.StartedAt.UTC()
	}
	if len(params.SessionJSON) > 0 {
		sessionJSON = []byte(params.SessionJSON)
	}
	if len(params.CodexSessionJSON) > 0 {
		codexJSON = []byte(params.CodexSessionJSON)
	}

	if _, err := r.db.ExecContext(
		ctx,
		queryUpsert,
		params.RunID,
		params.CorrelationID,
		params.ProjectID,
		strings.TrimSpace(params.RepositoryFullName),
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

// GetLatestByRepositoryBranch returns latest snapshot by repository + branch.
func (r *Repository) GetLatestByRepositoryBranch(ctx context.Context, repositoryFullName string, branchName string) (domainrepo.Session, bool, error) {
	var (
		item       domainrepo.Session
		projectID  sql.NullString
		issueNum   sql.NullInt64
		prNum      sql.NullInt64
		prURL      sql.NullString
		trigger    sql.NullString
		tplKind    sql.NullString
		tplSource  sql.NullString
		tplLocale  sql.NullString
		model      sql.NullString
		reasoning  sql.NullString
		sessionID  sql.NullString
		sessionRaw []byte
		path       sql.NullString
		codexRaw   []byte
		finishedAt sql.NullTime
	)

	err := r.db.QueryRowContext(
		ctx,
		queryGetLatestByRepositoryBranch,
		strings.TrimSpace(repositoryFullName),
		strings.TrimSpace(branchName),
	).Scan(
		&item.ID,
		&item.RunID,
		&item.CorrelationID,
		&projectID,
		&item.RepositoryFullName,
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
		&sessionID,
		&sessionRaw,
		&path,
		&codexRaw,
		&item.StartedAt,
		&finishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainrepo.Session{}, false, nil
		}
		return domainrepo.Session{}, false, fmt.Errorf("get latest agent session by repository+branch: %w", err)
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
	if finishedAt.Valid {
		item.FinishedAt = finishedAt.Time.UTC()
	}
	item.StartedAt = item.StartedAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	return item, true, nil
}
