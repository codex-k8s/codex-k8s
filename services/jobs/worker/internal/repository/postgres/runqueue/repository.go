package runqueue

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domainrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
)

var (
	//go:embed sql/claim_next_pending_for_update.sql
	queryClaimNextPendingForUpdate string
	//go:embed sql/upsert_project.sql
	queryUpsertProject string
	//go:embed sql/ensure_project_exists.sql
	queryEnsureProjectExists string
	//go:embed sql/ensure_project_slots.sql
	queryEnsureProjectSlots string
	//go:embed sql/release_expired_slots.sql
	queryReleaseExpiredSlots string
	//go:embed sql/lease_slot.sql
	queryLeaseSlot string
	//go:embed sql/mark_run_running.sql
	queryMarkRunRunning string
	//go:embed sql/list_running.sql
	queryListRunning string
	//go:embed sql/mark_run_finished.sql
	queryMarkRunFinished string
	//go:embed sql/mark_slot_releasing.sql
	queryMarkSlotReleasing string
	//go:embed sql/mark_slot_free.sql
	queryMarkSlotFree string
)

// Repository persists run queue state in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL run queue repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ClaimNextPending atomically claims one pending run and leases a slot.
func (r *Repository) ClaimNextPending(ctx context.Context, params domainrepo.ClaimParams) (domainrepo.ClaimedRun, bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("begin claim transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var (
		runID         string
		correlationID string
		projectIDRaw  sql.NullString
		learningMode  bool
		runPayload    []byte
	)

	err = tx.QueryRowContext(ctx, queryClaimNextPendingForUpdate).Scan(
		&runID,
		&correlationID,
		&projectIDRaw,
		&learningMode,
		&runPayload,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainrepo.ClaimedRun{}, false, nil
		}
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("select pending run for claim: %w", err)
	}

	projectID := projectIDRaw.String
	explicitProjectID := projectIDRaw.Valid && strings.TrimSpace(projectIDRaw.String) != ""
	if projectID == "" {
		projectID = deriveProjectID(correlationID, runPayload)
	}
	projectSlug, projectName := deriveProjectMeta(projectID, correlationID, runPayload)

	settingsJSON, err := json.Marshal(map[string]any{
		"learning_mode_default": params.ProjectLearningModeDefault,
	})
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("marshal project settings: %w", err)
	}

	if explicitProjectID {
		if _, err := tx.ExecContext(ctx, queryEnsureProjectExists, projectID, projectSlug, projectName, settingsJSON); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("ensure project %s exists: %w", projectID, err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, queryUpsertProject, projectID, projectSlug, projectName, settingsJSON); err != nil {
			return domainrepo.ClaimedRun{}, false, fmt.Errorf("upsert project %s: %w", projectID, err)
		}
	}

	if _, err := tx.ExecContext(ctx, queryEnsureProjectSlots, projectID, params.SlotsPerProject); err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("ensure slots for project %s: %w", projectID, err)
	}
	if _, err := tx.ExecContext(ctx, queryReleaseExpiredSlots, projectID); err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("release expired slots for project %s: %w", projectID, err)
	}

	leaseUntilInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(params.LeaseTTL.Seconds())))
	var (
		slotID string
		slotNo int
	)
	if err := tx.QueryRowContext(ctx, queryLeaseSlot, projectID, runID, leaseUntilInterval).Scan(&slotID, &slotNo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainrepo.ClaimedRun{}, false, nil
		}
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("lease slot for run %s: %w", runID, err)
	}

	res, err := tx.ExecContext(ctx, queryMarkRunRunning, runID, projectID)
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("mark run %s as running: %w", runID, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("read rows affected when mark run %s as running: %w", runID, err)
	}
	if rows == 0 {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("mark run %s as running affected 0 rows", runID)
	}

	if err := tx.Commit(); err != nil {
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

// ListRunning returns active runs eligible for job reconciliation.
func (r *Repository) ListRunning(ctx context.Context, limit int) ([]domainrepo.RunningRun, error) {
	rows, err := r.db.QueryContext(ctx, queryListRunning, limit)
	if err != nil {
		return nil, fmt.Errorf("list running runs: %w", err)
	}
	defer rows.Close()

	result := make([]domainrepo.RunningRun, 0, limit)
	for rows.Next() {
		var (
			runID         string
			correlationID string
			projectID     string
			learningMode  bool
			startedAt     sql.NullTime
		)
		if err := rows.Scan(&runID, &correlationID, &projectID, &learningMode, &startedAt); err != nil {
			return nil, fmt.Errorf("scan running run row: %w", err)
		}
		item := domainrepo.RunningRun{
			RunID:         runID,
			CorrelationID: correlationID,
			ProjectID:     projectID,
			LearningMode:  learningMode,
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

// FinishRun sets final status and releases leased slot.
func (r *Repository) FinishRun(ctx context.Context, params domainrepo.FinishParams) (bool, error) {
	if params.Status != "succeeded" && params.Status != "failed" && params.Status != "canceled" {
		return false, fmt.Errorf("unsupported final run status %q", params.Status)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin finish transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, queryMarkRunFinished, params.RunID, params.Status, params.FinishedAt.UTC())
	if err != nil {
		return false, fmt.Errorf("mark run %s as %s: %w", params.RunID, params.Status, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read rows affected for finish run %s: %w", params.RunID, err)
	}
	if rows == 0 {
		return false, nil
	}

	if _, err := tx.ExecContext(ctx, queryMarkSlotReleasing, params.ProjectID, params.RunID); err != nil {
		return false, fmt.Errorf("mark slot releasing for run %s: %w", params.RunID, err)
	}
	if _, err := tx.ExecContext(ctx, queryMarkSlotFree, params.ProjectID, params.RunID); err != nil {
		return false, fmt.Errorf("mark slot free for run %s: %w", params.RunID, err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit finish transaction: %w", err)
	}

	return true, nil
}

func deriveProjectID(correlationID string, runPayload []byte) string {
	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(runPayload, &payload); err == nil && payload.Repository.FullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(payload.Repository.FullName))).String()
	}

	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}

func deriveProjectMeta(projectID string, correlationID string, runPayload []byte) (slug string, name string) {
	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(runPayload, &payload); err == nil && payload.Repository.FullName != "" {
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

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
