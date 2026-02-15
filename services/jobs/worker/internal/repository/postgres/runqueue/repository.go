package runqueue

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	domainrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
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
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL run queue repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ClaimNextPending atomically claims one pending run and leases a slot.
func (r *Repository) ClaimNextPending(ctx context.Context, params domainrepo.ClaimParams) (domainrepo.ClaimedRun, bool, error) {
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

	projectID := projectIDRaw.String
	explicitProjectID := projectIDRaw.Valid && strings.TrimSpace(projectIDRaw.String) != ""
	if projectID == "" {
		projectID = deriveProjectID(correlationID, runPayload)
	}
	projectSlug, projectName := deriveProjectMeta(projectID, correlationID, runPayload)

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

	if _, err := tx.Exec(ctx, queryEnsureProjectSlots, projectID, params.SlotsPerProject); err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("ensure slots for project %s: %w", projectID, err)
	}
	if _, err := tx.Exec(ctx, queryReleaseExpiredSlots, projectID); err != nil {
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("release expired slots for project %s: %w", projectID, err)
	}

	leaseUntilInterval := fmt.Sprintf("%d seconds", maxInt64(1, int64(params.LeaseTTL.Seconds())))
	var (
		slotID string
		slotNo int
	)
	if err := tx.QueryRow(ctx, queryLeaseSlot, projectID, runID, leaseUntilInterval).Scan(&slotID, &slotNo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.ClaimedRun{}, false, nil
		}
		return domainrepo.ClaimedRun{}, false, fmt.Errorf("lease slot for run %s: %w", runID, err)
	}

	res, err := tx.Exec(ctx, queryMarkRunRunning, runID, projectID)
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

// ListRunning returns active runs eligible for job reconciliation.
func (r *Repository) ListRunning(ctx context.Context, limit int) ([]domainrepo.RunningRun, error) {
	rows, err := r.db.Query(ctx, queryListRunning, limit)
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
			slotID        string
			slotNo        int
			learningMode  bool
			runPayload    []byte
			startedAt     pgtype.Timestamptz
		)
		if err := rows.Scan(&runID, &correlationID, &projectID, &slotID, &slotNo, &learningMode, &runPayload, &startedAt); err != nil {
			return nil, fmt.Errorf("scan running run row: %w", err)
		}
		item := domainrepo.RunningRun{
			RunID:         runID,
			CorrelationID: correlationID,
			ProjectID:     projectID,
			SlotID:        slotID,
			SlotNo:        slotNo,
			LearningMode:  learningMode,
			RunPayload:    json.RawMessage(runPayload),
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
	if params.Status != rundomain.StatusSucceeded && params.Status != rundomain.StatusFailed && params.Status != rundomain.StatusCanceled {
		return false, fmt.Errorf("unsupported final run status %q", params.Status)
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin finish transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res, err := tx.Exec(ctx, queryMarkRunFinished, params.RunID, string(params.Status), params.FinishedAt.UTC())
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

// deriveProjectID prefers repository identity and falls back to correlation-scoped synthetic id.
func deriveProjectID(correlationID string, runPayload []byte) string {
	var payload querytypes.RunQueuePayload
	if err := json.Unmarshal(runPayload, &payload); err == nil && payload.Repository.FullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(payload.Repository.FullName))).String()
	}

	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}

// deriveProjectMeta builds stable project slug/name values from payload or synthetic fallback.
func deriveProjectMeta(projectID string, correlationID string, runPayload []byte) (slug string, name string) {
	var payload querytypes.RunQueuePayload

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

// maxInt64 returns the greater of two int64 values.
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
