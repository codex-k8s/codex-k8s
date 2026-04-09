package changegovernance

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	"github.com/codex-k8s/kodex/libs/go/errs"
	domainrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/changegovernance"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/changegovernance/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/get_package_by_id.sql
var queryGetPackageByID string

//go:embed sql/get_package_by_key_for_update.sql
var queryGetPackageByKeyForUpdate string

//go:embed sql/get_package_by_id_for_update.sql
var queryGetPackageByIDForUpdate string

//go:embed sql/insert_package.sql
var queryInsertPackage string

//go:embed sql/update_package_summary.sql
var queryUpdatePackageSummary string

//go:embed sql/insert_draft_if_absent.sql
var queryInsertDraftIfAbsent string

//go:embed sql/get_draft_by_signal_id.sql
var queryGetDraftBySignalID string

//go:embed sql/upsert_wave.sql
var queryUpsertWave string

//go:embed sql/stage_wave_publish_orders.sql
var queryStageWavePublishOrders string

//go:embed sql/supersede_missing_waves.sql
var querySupersedeMissingWaves string

//go:embed sql/get_wave_by_package_and_key.sql
var queryGetWaveByPackageAndKey string

//go:embed sql/get_evidence_block_by_scope.sql
var queryGetEvidenceBlockByScope string

//go:embed sql/insert_evidence_block.sql
var queryInsertEvidenceBlock string

//go:embed sql/update_evidence_block.sql
var queryUpdateEvidenceBlock string

//go:embed sql/update_wave_summary.sql
var queryUpdateWaveSummary string

//go:embed sql/upsert_artifact_link.sql
var queryUpsertArtifactLink string

//go:embed sql/list_drafts.sql
var queryListDrafts string

//go:embed sql/list_waves.sql
var queryListWaves string

//go:embed sql/list_evidence_blocks.sql
var queryListEvidenceBlocks string

//go:embed sql/list_decision_records.sql
var queryListDecisionRecords string

//go:embed sql/list_feedback_records.sql
var queryListFeedbackRecords string

//go:embed sql/list_artifact_links.sql
var queryListArtifactLinks string

//go:embed sql/list_current_projection_snapshots.sql
var queryListCurrentProjectionSnapshots string

//go:embed sql/deactivate_current_projections.sql
var queryDeactivateCurrentProjections string

//go:embed sql/insert_projection_snapshot.sql
var queryInsertProjectionSnapshot string

//go:embed sql/insert_flow_event.sql
var queryInsertFlowEvent string

// Repository persists change-governance aggregate state in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL change-governance repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// RecordDraftSignal records one hidden-draft ledger row and refreshes projections.
func (r *Repository) RecordDraftSignal(ctx context.Context, params querytypes.ChangeGovernanceDraftSignalParams) (domainrepo.Aggregate, bool, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Aggregate{}, false, fmt.Errorf("begin record change-governance draft tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	packageKey := buildPackageKey(params.RepositoryFullName, params.IssueNumber)
	pkg, found, err := r.getPackageByKeyForUpdate(ctx, tx, params.ProjectID, packageKey)
	if err != nil {
		return domainrepo.Aggregate{}, false, err
	}
	if !found {
		pkg, err = r.insertPackage(ctx, tx, params, packageKey)
		if err != nil {
			return domainrepo.Aggregate{}, false, err
		}
	}

	draft, inserted, err := r.insertDraftIfAbsent(ctx, tx, pkg.ID, params)
	if err != nil {
		return domainrepo.Aggregate{}, false, err
	}
	if !inserted {
		draft, err = r.getDraftBySignalID(ctx, tx, params.SignalID)
		if err != nil {
			return domainrepo.Aggregate{}, false, err
		}
	}

	if err := r.upsertPrimaryArtifactLinks(ctx, tx, pkg.ID, params.RepositoryFullName, params.IssueNumber, params.PRNumber, params.RunID, draft.DraftRef); err != nil {
		return domainrepo.Aggregate{}, false, err
	}

	aggregate, err := r.buildAggregate(ctx, tx, pkg.ID)
	if err != nil {
		return domainrepo.Aggregate{}, false, err
	}

	derived := derivePackageState(aggregate, params.CorrelationID, params.PRNumber)
	pkg, err = r.updatePackageSummary(ctx, tx, aggregate.Package, derived, params.OccurredAt)
	if err != nil {
		return domainrepo.Aggregate{}, false, err
	}
	aggregate.Package = pkg

	if err := r.refreshCurrentProjections(ctx, tx, aggregate); err != nil {
		return domainrepo.Aggregate{}, false, err
	}
	if err := r.insertFlowEvent(ctx, tx, params.CorrelationID, floweventdomain.ActorIDControlPlane, floweventdomain.EventTypeQualityGovernancePackageUpserted, draftPackageUpsertedEventPayload{
		PackageID:        pkg.ID,
		PackageKey:       pkg.PackageKey,
		DraftSignalID:    params.SignalID,
		PublicationState: pkg.PublicationState,
	}); err != nil {
		return domainrepo.Aggregate{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Aggregate{}, false, fmt.Errorf("commit record change-governance draft tx: %w", err)
	}
	return aggregate, !inserted, nil
}

// PublishWaveMap records semantic-wave lineage and refreshes projections.
func (r *Repository) PublishWaveMap(ctx context.Context, params querytypes.ChangeGovernanceWaveMapParams) (domainrepo.Aggregate, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Aggregate{}, fmt.Errorf("begin publish change-governance wave-map tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pkg, found, err := r.getPackageByIDForUpdate(ctx, tx, params.PackageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	if !found {
		return domainrepo.Aggregate{}, errs.NotFound{Msg: fmt.Sprintf("change governance package %q not found", params.PackageID)}
	}
	if strings.TrimSpace(params.ExpectedProjectID) != "" && pkg.ProjectID != strings.TrimSpace(params.ExpectedProjectID) {
		return domainrepo.Aggregate{}, errs.Forbidden{Msg: "change governance package is outside authenticated project scope"}
	}

	sortedWaves := append([]querytypes.ChangeGovernanceWaveDraft(nil), params.Waves...)
	slices.SortFunc(sortedWaves, func(left querytypes.ChangeGovernanceWaveDraft, right querytypes.ChangeGovernanceWaveDraft) int {
		switch {
		case left.PublishOrder < right.PublishOrder:
			return -1
		case left.PublishOrder > right.PublishOrder:
			return 1
		default:
			return strings.Compare(left.WaveKey, right.WaveKey)
		}
	})
	if err := r.stageWavePublishOrders(ctx, tx, pkg.ID); err != nil {
		return domainrepo.Aggregate{}, err
	}
	activeWaveKeys := make([]string, 0, len(sortedWaves))
	for _, wave := range sortedWaves {
		activeWaveKeys = append(activeWaveKeys, strings.TrimSpace(wave.WaveKey))
		if _, err := r.upsertWave(ctx, tx, pkg.ID, wave); err != nil {
			return domainrepo.Aggregate{}, err
		}
	}
	if err := r.supersedeMissingWaves(ctx, tx, pkg.ID, activeWaveKeys); err != nil {
		return domainrepo.Aggregate{}, err
	}
	if err := r.refreshWaveSummaries(ctx, tx, pkg.ID); err != nil {
		return domainrepo.Aggregate{}, err
	}

	aggregate, err := r.buildAggregate(ctx, tx, pkg.ID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	derived := derivePackageState(aggregate, params.CorrelationID, pkg.PRNumber)
	pkg, err = r.updatePackageSummary(ctx, tx, aggregate.Package, derived, params.PublishedAt)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	aggregate.Package = pkg

	if err := r.refreshCurrentProjections(ctx, tx, aggregate); err != nil {
		return domainrepo.Aggregate{}, err
	}
	if err := r.insertFlowEvent(ctx, tx, params.CorrelationID, floweventdomain.ActorIDControlPlane, floweventdomain.EventTypeQualityGovernanceWaveMapPublished, waveMapPublishedEventPayload{
		PackageID:         pkg.ID,
		WaveMapID:         params.WaveMapID,
		PublicationState:  pkg.PublicationState,
		ProjectionVersion: pkg.ActiveProjectionVersion,
	}); err != nil {
		return domainrepo.Aggregate{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Aggregate{}, fmt.Errorf("commit publish change-governance wave-map tx: %w", err)
	}
	return aggregate, nil
}

// UpsertEvidenceSignal records one evidence block and refreshes projections.
func (r *Repository) UpsertEvidenceSignal(ctx context.Context, params querytypes.ChangeGovernanceEvidenceSignalParams) (domainrepo.Aggregate, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.Aggregate{}, fmt.Errorf("begin upsert change-governance evidence tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pkg, found, err := r.getPackageByIDForUpdate(ctx, tx, params.PackageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	if !found {
		return domainrepo.Aggregate{}, errs.NotFound{Msg: fmt.Sprintf("change governance package %q not found", params.PackageID)}
	}
	if strings.TrimSpace(params.ExpectedProjectID) != "" && pkg.ProjectID != strings.TrimSpace(params.ExpectedProjectID) {
		return domainrepo.Aggregate{}, errs.Forbidden{Msg: "change governance package is outside authenticated project scope"}
	}

	var waveID string
	if params.ScopeKind == enumtypes.ChangeGovernanceEvidenceScopeKindWave {
		wave, waveFound, waveErr := r.getWaveByPackageAndKey(ctx, tx, pkg.ID, params.ScopeRef)
		if waveErr != nil {
			return domainrepo.Aggregate{}, waveErr
		}
		if !waveFound {
			return domainrepo.Aggregate{}, errs.NotFound{Msg: fmt.Sprintf("change governance wave %q not found for package %q", params.ScopeRef, pkg.ID)}
		}
		waveID = wave.ID
	}

	if err := r.upsertEvidenceBlock(ctx, tx, pkg.ID, waveID, params); err != nil {
		return domainrepo.Aggregate{}, err
	}
	if err := r.upsertArtifactLinks(ctx, tx, pkg.ID, params.ArtifactLinks); err != nil {
		return domainrepo.Aggregate{}, err
	}
	if err := r.refreshWaveSummaries(ctx, tx, pkg.ID); err != nil {
		return domainrepo.Aggregate{}, err
	}

	aggregate, err := r.buildAggregate(ctx, tx, pkg.ID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	derived := derivePackageState(aggregate, params.CorrelationID, pkg.PRNumber)
	pkg, err = r.updatePackageSummary(ctx, tx, aggregate.Package, derived, params.OccurredAt)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	aggregate.Package = pkg

	if err := r.refreshCurrentProjections(ctx, tx, aggregate); err != nil {
		return domainrepo.Aggregate{}, err
	}
	if err := r.insertFlowEvent(ctx, tx, params.CorrelationID, floweventdomain.ActorIDControlPlane, floweventdomain.EventTypeQualityGovernanceProjectionRefreshed, projectionRefreshedEventPayload{
		PackageID:                 pkg.ID,
		ScopeKind:                 params.ScopeKind,
		ScopeRef:                  params.ScopeRef,
		BlockKind:                 params.BlockKind,
		ProjectionVersion:         pkg.ActiveProjectionVersion,
		EvidenceCompletenessState: pkg.EvidenceCompletenessState,
		VerificationMinimumState:  pkg.VerificationMinimumState,
	}); err != nil {
		return domainrepo.Aggregate{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.Aggregate{}, fmt.Errorf("commit upsert change-governance evidence tx: %w", err)
	}
	return aggregate, nil
}

// GetAggregateByPackageID returns one hydrated aggregate by package id.
func (r *Repository) GetAggregateByPackageID(ctx context.Context, packageID string) (domainrepo.Aggregate, bool, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return domainrepo.Aggregate{}, false, fmt.Errorf("begin get change-governance aggregate tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pkg, found, err := r.getPackageByID(ctx, tx, packageID)
	if err != nil || !found {
		return domainrepo.Aggregate{}, found, err
	}
	aggregate, err := r.buildAggregate(ctx, tx, pkg.ID)
	if err != nil {
		return domainrepo.Aggregate{}, false, err
	}
	return aggregate, true, nil
}

func queryOptionalDomainRow[DBRow any, DomainItem any](
	ctx context.Context,
	tx pgx.Tx,
	query string,
	queryErr string,
	collectErr string,
	convert func(DBRow) DomainItem,
	args ...any,
) (DomainItem, bool, error) {
	var zero DomainItem

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return zero, false, fmt.Errorf("%s: %w", queryErr, err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DBRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("%s: %w", collectErr, err)
	}
	return convert(row), true, nil
}

func queryDomainRowsByPackageID[DBRow any, DomainItem any](
	ctx context.Context,
	tx pgx.Tx,
	query string,
	packageID string,
	queryErr string,
	collectErr string,
	convert func(DBRow) DomainItem,
) ([]DomainItem, error) {
	rows, err := tx.Query(ctx, query, strings.TrimSpace(packageID))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", queryErr, err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBRow])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", collectErr, err)
	}
	result := make([]DomainItem, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}
	return result, nil
}

func (r *Repository) getPackageByID(ctx context.Context, tx pgx.Tx, packageID string) (domainrepo.Package, bool, error) {
	return queryOptionalDomainRow(
		ctx,
		tx,
		queryGetPackageByID,
		"query change-governance package by id",
		"collect change-governance package by id",
		fromPackageRow,
		strings.TrimSpace(packageID),
	)
}

func (r *Repository) getPackageByKeyForUpdate(ctx context.Context, tx pgx.Tx, projectID string, packageKey string) (domainrepo.Package, bool, error) {
	queryArgs := []any{strings.TrimSpace(projectID), strings.TrimSpace(packageKey)}
	return queryOptionalDomainRow(
		ctx,
		tx,
		queryGetPackageByKeyForUpdate,
		"query change-governance package by key for update",
		"collect change-governance package by key for update",
		fromPackageRow,
		queryArgs...,
	)
}

func (r *Repository) getPackageByIDForUpdate(ctx context.Context, tx pgx.Tx, packageID string) (domainrepo.Package, bool, error) {
	return queryOptionalDomainRow(
		ctx,
		tx,
		queryGetPackageByIDForUpdate,
		"query change-governance package by id for update",
		"collect change-governance package by id for update",
		fromPackageRow,
		strings.TrimSpace(packageID),
	)
}

func (r *Repository) insertPackage(ctx context.Context, tx pgx.Tx, params querytypes.ChangeGovernanceDraftSignalParams, packageKey string) (domainrepo.Package, error) {
	rows, err := tx.Query(
		ctx,
		queryInsertPackage,
		packageKey,
		strings.TrimSpace(params.ProjectID),
		strings.TrimSpace(params.RepositoryFullName),
		params.IssueNumber,
		intPtrToPGInt4(params.PRNumber),
	)
	if err != nil {
		return domainrepo.Package{}, fmt.Errorf("insert change-governance package: %w", err)
	}
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.PackageRow])
	if err != nil {
		return domainrepo.Package{}, fmt.Errorf("insert change-governance package: %w", err)
	}
	return fromPackageRow(item), nil
}

func (r *Repository) insertDraftIfAbsent(ctx context.Context, tx pgx.Tx, packageID string, params querytypes.ChangeGovernanceDraftSignalParams) (domainrepo.InternalDraft, bool, error) {
	row := tx.QueryRow(
		ctx,
		queryInsertDraftIfAbsent,
		packageID,
		nullableText(strings.TrimSpace(params.RunID)),
		strings.TrimSpace(params.SignalID),
		strings.TrimSpace(params.DraftRef),
		nullableText(strings.TrimSpace(params.DraftChecksum)),
		string(params.DraftKind),
		jsonOrEmptyObject(marshalJSONPayload(draftMetadata{
			ChangeScopeHints:     params.ChangeScopeHints,
			CandidateRiskDrivers: params.CandidateRiskDrivers,
			BranchName:           strings.TrimSpace(params.BranchName),
		})),
		timestamptzOrNow(params.OccurredAt),
	)
	var item dbmodel.InternalDraftRow
	if err := row.Scan(
		&item.ID,
		&item.PackageID,
		&item.RunID,
		&item.SignalID,
		&item.DraftRef,
		&item.DraftChecksum,
		&item.DraftKind,
		&item.MetadataJSON,
		&item.IsLatest,
		&item.OccurredAt,
		&item.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.InternalDraft{}, false, nil
		}
		return domainrepo.InternalDraft{}, false, fmt.Errorf("insert change-governance draft if absent: %w", err)
	}
	return fromDraftRow(item), true, nil
}

func (r *Repository) getDraftBySignalID(ctx context.Context, tx pgx.Tx, signalID string) (domainrepo.InternalDraft, error) {
	rows, err := tx.Query(ctx, queryGetDraftBySignalID, strings.TrimSpace(signalID))
	if err != nil {
		return domainrepo.InternalDraft{}, fmt.Errorf("query change-governance draft by signal id: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.InternalDraftRow])
	if err != nil {
		return domainrepo.InternalDraft{}, fmt.Errorf("collect change-governance draft by signal id: %w", err)
	}
	return fromDraftRow(row), nil
}

func (r *Repository) stageWavePublishOrders(ctx context.Context, tx pgx.Tx, packageID string) error {
	if _, err := tx.Exec(ctx, queryStageWavePublishOrders, strings.TrimSpace(packageID)); err != nil {
		return fmt.Errorf("stage change-governance wave publish orders: %w", err)
	}
	return nil
}

func (r *Repository) supersedeMissingWaves(ctx context.Context, tx pgx.Tx, packageID string, activeWaveKeys []string) error {
	if _, err := tx.Exec(ctx, querySupersedeMissingWaves, strings.TrimSpace(packageID), activeWaveKeys); err != nil {
		return fmt.Errorf("supersede missing change-governance waves: %w", err)
	}
	return nil
}

func (r *Repository) upsertWave(ctx context.Context, tx pgx.Tx, packageID string, wave querytypes.ChangeGovernanceWaveDraft) (domainrepo.Wave, error) {
	rows, err := tx.Query(
		ctx,
		queryUpsertWave,
		strings.TrimSpace(packageID),
		strings.TrimSpace(wave.WaveKey),
		wave.PublishOrder,
		string(wave.DominantIntent),
		string(wave.BoundedScopeKind),
		string(enumtypes.ChangeGovernanceWavePublicationStatePublished),
		strings.TrimSpace(wave.Summary),
		jsonOrEmptyArray(marshalJSONPayload(wave.VerificationTargets)),
	)
	if err != nil {
		return domainrepo.Wave{}, fmt.Errorf("upsert change-governance wave: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.WaveRow])
	if err != nil {
		return domainrepo.Wave{}, fmt.Errorf("collect change-governance wave upsert: %w", err)
	}
	return fromWaveRow(row), nil
}

func (r *Repository) getWaveByPackageAndKey(ctx context.Context, tx pgx.Tx, packageID string, waveKey string) (domainrepo.Wave, bool, error) {
	rows, err := tx.Query(ctx, queryGetWaveByPackageAndKey, strings.TrimSpace(packageID), strings.TrimSpace(waveKey))
	if err != nil {
		return domainrepo.Wave{}, false, fmt.Errorf("query change-governance wave by key: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.WaveRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Wave{}, false, nil
		}
		return domainrepo.Wave{}, false, fmt.Errorf("collect change-governance wave by key: %w", err)
	}
	return fromWaveRow(row), true, nil
}

func (r *Repository) upsertEvidenceBlock(ctx context.Context, tx pgx.Tx, packageID string, waveID string, params querytypes.ChangeGovernanceEvidenceSignalParams) error {
	existing, found, err := r.getEvidenceBlockByScope(ctx, tx, packageID, waveID, params.BlockKind)
	if err != nil {
		return err
	}

	state := evidenceBlockStateFromSignal(params)
	artifactJSON := jsonOrEmptyArray(marshalJSONPayload(params.ArtifactLinks))
	if !found {
		if _, err := tx.Exec(
			ctx,
			queryInsertEvidenceBlock,
			strings.TrimSpace(packageID),
			nullableText(strings.TrimSpace(waveID)),
			string(params.BlockKind),
			string(state),
			string(verificationStateFromHint(params.VerificationStateHint)),
			params.RequiredByTier,
			string(enumtypes.ChangeGovernanceEvidenceSourceKindAgentSignal),
			artifactJSON,
			nullableText(strings.TrimSpace(params.SignalID)),
			timestamptzOrNow(params.OccurredAt),
		); err != nil {
			return fmt.Errorf("insert change-governance evidence block: %w", err)
		}
		return nil
	}

	if _, err := tx.Exec(
		ctx,
		queryUpdateEvidenceBlock,
		existing.ID,
		string(state),
		string(verificationStateFromHint(params.VerificationStateHint)),
		params.RequiredByTier,
		artifactJSON,
		nullableText(strings.TrimSpace(params.SignalID)),
		timestamptzOrNow(params.OccurredAt),
	); err != nil {
		return fmt.Errorf("update change-governance evidence block: %w", err)
	}
	return nil
}

func (r *Repository) getEvidenceBlockByScope(ctx context.Context, tx pgx.Tx, packageID string, waveID string, blockKind enumtypes.ChangeGovernanceEvidenceBlockKind) (domainrepo.EvidenceBlock, bool, error) {
	rows, err := tx.Query(ctx, queryGetEvidenceBlockByScope, strings.TrimSpace(packageID), nullableText(strings.TrimSpace(waveID)), string(blockKind))
	if err != nil {
		return domainrepo.EvidenceBlock{}, false, fmt.Errorf("query change-governance evidence block by scope: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.EvidenceBlockRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.EvidenceBlock{}, false, nil
		}
		return domainrepo.EvidenceBlock{}, false, fmt.Errorf("collect change-governance evidence block by scope: %w", err)
	}
	return fromEvidenceBlockRow(row), true, nil
}

func (r *Repository) refreshWaveSummaries(ctx context.Context, tx pgx.Tx, packageID string) error {
	waves, err := r.listWaves(ctx, tx, packageID)
	if err != nil {
		return err
	}
	if len(waves) == 0 {
		return nil
	}
	evidenceBlocks, err := r.listEvidenceBlocks(ctx, tx, packageID)
	if err != nil {
		return err
	}
	summaries := deriveWaveSummaryStates(waves, evidenceBlocks)
	for _, wave := range waves {
		summary := summaries[wave.ID]
		if _, err := tx.Exec(
			ctx,
			queryUpdateWaveSummary,
			strings.TrimSpace(wave.ID),
			string(summary.EvidenceCompletenessState),
			string(summary.VerificationMinimumState),
		); err != nil {
			return fmt.Errorf("update change-governance wave summary: %w", err)
		}
	}
	return nil
}

func (r *Repository) upsertPrimaryArtifactLinks(ctx context.Context, tx pgx.Tx, packageID string, repositoryFullName string, issueNumber int, prNumber *int, runID string, draftRef string) error {
	seeds := []querytypes.ChangeGovernanceArtifactLinkSeed{
		{
			ArtifactKind: enumtypes.ChangeGovernanceArtifactKindIssue,
			ArtifactRef:  fmt.Sprintf("%s#%d", strings.TrimSpace(repositoryFullName), issueNumber),
			RelationKind: enumtypes.ChangeGovernanceArtifactRelationKindPrimaryContext,
			DisplayLabel: fmt.Sprintf("Issue #%d", issueNumber),
		},
		{
			ArtifactKind: enumtypes.ChangeGovernanceArtifactKindRun,
			ArtifactRef:  strings.TrimSpace(runID),
			RelationKind: enumtypes.ChangeGovernanceArtifactRelationKindEvidenceSource,
			DisplayLabel: fmt.Sprintf("Run %s", strings.TrimSpace(runID)),
		},
	}
	if prNumber != nil && *prNumber > 0 {
		seeds = append(seeds, querytypes.ChangeGovernanceArtifactLinkSeed{
			ArtifactKind: enumtypes.ChangeGovernanceArtifactKindPullRequest,
			ArtifactRef:  fmt.Sprintf("%s#%d", strings.TrimSpace(repositoryFullName), *prNumber),
			RelationKind: enumtypes.ChangeGovernanceArtifactRelationKindEvidenceSource,
			DisplayLabel: fmt.Sprintf("PR #%d", *prNumber),
		})
	}
	draftRef = strings.TrimSpace(draftRef)
	if draftRef != "" {
		seeds = append(seeds, querytypes.ChangeGovernanceArtifactLinkSeed{
			ArtifactKind: enumtypes.ChangeGovernanceArtifactKindAgentSession,
			ArtifactRef:  draftRef,
			RelationKind: enumtypes.ChangeGovernanceArtifactRelationKindEvidenceSource,
			DisplayLabel: draftRef,
		})
	}
	return r.upsertArtifactLinks(ctx, tx, packageID, seeds)
}

func (r *Repository) upsertArtifactLinks(ctx context.Context, tx pgx.Tx, packageID string, seeds []querytypes.ChangeGovernanceArtifactLinkSeed) error {
	for _, seed := range seeds {
		if strings.TrimSpace(seed.ArtifactRef) == "" {
			continue
		}
		if _, err := tx.Exec(
			ctx,
			queryUpsertArtifactLink,
			strings.TrimSpace(packageID),
			string(seed.ArtifactKind),
			strings.TrimSpace(seed.ArtifactRef),
			string(seed.RelationKind),
			strings.TrimSpace(seed.DisplayLabel),
		); err != nil {
			return fmt.Errorf("upsert change-governance artifact link: %w", err)
		}
	}
	return nil
}

func (r *Repository) updatePackageSummary(ctx context.Context, tx pgx.Tx, current domainrepo.Package, derived derivedPackageState, observedAt time.Time) (domainrepo.Package, error) {
	rows, err := tx.Query(
		ctx,
		queryUpdatePackageSummary,
		strings.TrimSpace(current.ID),
		intPtrToPGInt4(derived.PRNumber),
		nullableText(string(derived.RiskTier)),
		string(derived.BundleAdmissibility),
		string(derived.PublicationState),
		string(derived.EvidenceCompletenessState),
		string(derived.VerificationMinimumState),
		string(derived.WaiverState),
		string(derived.ReleaseReadinessState),
		string(derived.GovernanceFeedbackState),
		nullableText(strings.TrimSpace(derived.LatestCorrelationID)),
		timestamptzOrNow(observedAt),
		current.ActiveProjectionVersion,
	)
	if err != nil {
		return domainrepo.Package{}, fmt.Errorf("update change-governance package summary: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.PackageRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Package{}, errs.Conflict{Msg: "change governance package projection version is stale"}
		}
		return domainrepo.Package{}, fmt.Errorf("collect change-governance package summary update: %w", err)
	}
	return fromPackageRow(row), nil
}

func (r *Repository) buildAggregate(ctx context.Context, tx pgx.Tx, packageID string) (domainrepo.Aggregate, error) {
	pkg, found, err := r.getPackageByID(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	if !found {
		return domainrepo.Aggregate{}, errs.NotFound{Msg: fmt.Sprintf("change governance package %q not found", packageID)}
	}

	drafts, err := r.listDrafts(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	waves, err := r.listWaves(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	evidence, err := r.listEvidenceBlocks(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	decisions, err := r.listDecisionRecords(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	feedback, err := r.listFeedbackRecords(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	artifacts, err := r.listArtifactLinks(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}
	projections, err := r.listCurrentProjectionSnapshots(ctx, tx, packageID)
	if err != nil {
		return domainrepo.Aggregate{}, err
	}

	currentProjections := make(map[enumtypes.ChangeGovernanceProjectionKind]domainrepo.ProjectionSnapshot, len(projections))
	for _, projection := range projections {
		currentProjections[projection.ProjectionKind] = projection
	}

	return valuetypes.ChangeGovernanceAggregate{
		Package:            pkg,
		Drafts:             drafts,
		Waves:              waves,
		EvidenceBlocks:     evidence,
		DecisionRecords:    decisions,
		FeedbackRecords:    feedback,
		ArtifactLinks:      artifacts,
		CurrentProjections: currentProjections,
	}, nil
}

func (r *Repository) listDrafts(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.InternalDraft, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListDrafts,
		packageID,
		"query change-governance drafts",
		"collect change-governance drafts",
		fromDraftRow,
	)
}

func (r *Repository) listWaves(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.Wave, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListWaves,
		packageID,
		"query change-governance waves",
		"collect change-governance waves",
		fromWaveRow,
	)
}

func (r *Repository) listEvidenceBlocks(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.EvidenceBlock, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListEvidenceBlocks,
		packageID,
		"query change-governance evidence blocks",
		"collect change-governance evidence blocks",
		fromEvidenceBlockRow,
	)
}

func (r *Repository) listDecisionRecords(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.DecisionRecord, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListDecisionRecords,
		packageID,
		"query change-governance decision records",
		"collect change-governance decision records",
		fromDecisionRecordRow,
	)
}

func (r *Repository) listFeedbackRecords(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.FeedbackRecord, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListFeedbackRecords,
		packageID,
		"query change-governance feedback records",
		"collect change-governance feedback records",
		fromFeedbackRecordRow,
	)
}

func (r *Repository) listArtifactLinks(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.ArtifactLink, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListArtifactLinks,
		packageID,
		"query change-governance artifact links",
		"collect change-governance artifact links",
		fromArtifactLinkRow,
	)
}

func (r *Repository) listCurrentProjectionSnapshots(ctx context.Context, tx pgx.Tx, packageID string) ([]domainrepo.ProjectionSnapshot, error) {
	return queryDomainRowsByPackageID(
		ctx,
		tx,
		queryListCurrentProjectionSnapshots,
		packageID,
		"query current change-governance projections",
		"collect current change-governance projections",
		fromProjectionSnapshotRow,
	)
}

func (r *Repository) refreshCurrentProjections(ctx context.Context, tx pgx.Tx, aggregate domainrepo.Aggregate) error {
	if _, err := tx.Exec(ctx, queryDeactivateCurrentProjections, strings.TrimSpace(aggregate.Package.ID)); err != nil {
		return fmt.Errorf("deactivate current change-governance projections: %w", err)
	}

	payloads, err := buildProjectionPayloads(aggregate)
	if err != nil {
		return err
	}
	for kind, payload := range payloads {
		if _, err := tx.Exec(
			ctx,
			queryInsertProjectionSnapshot,
			strings.TrimSpace(aggregate.Package.ID),
			string(kind),
			aggregate.Package.ActiveProjectionVersion,
			payload,
			timestamptzOrNow(aggregate.Package.UpdatedAt),
		); err != nil {
			return fmt.Errorf("insert change-governance projection snapshot %s: %w", kind, err)
		}
	}
	return nil
}

func buildProjectionPayloads(aggregate domainrepo.Aggregate) (map[enumtypes.ChangeGovernanceProjectionKind][]byte, error) {
	summary := buildProjectionPackageSummary(aggregate)
	artifacts := buildProjectionArtifactLinks(aggregate.ArtifactLinks)
	waves := buildProjectionWaveItems(aggregate.Waves)
	evidenceBlocks := buildProjectionEvidenceBlocks(aggregate.Waves, aggregate.EvidenceBlocks)
	decisions := buildProjectionDecisionSummaries(aggregate.DecisionRecords)
	feedback := buildProjectionFeedbackRecords(aggregate.FeedbackRecords)

	payloads := map[enumtypes.ChangeGovernanceProjectionKind]any{
		enumtypes.ChangeGovernanceProjectionKindPackageList: summary,
		enumtypes.ChangeGovernanceProjectionKindPackageDetail: packageDetailProjection{
			Package:            summary,
			Waves:              waves,
			EvidenceBlocks:     evidenceBlocks,
			ActiveDecisions:    decisions,
			FeedbackRecords:    feedback,
			ArtifactLinks:      artifacts,
			CommentMirrorState: "not_attempted",
		},
		enumtypes.ChangeGovernanceProjectionKindOperatorGapQueue: operatorGapQueueProjection{
			PackageID: aggregate.Package.ID,
			Items:     feedback,
		},
		enumtypes.ChangeGovernanceProjectionKindReleaseGate: releaseGateProjection{
			Package: summary,
		},
		enumtypes.ChangeGovernanceProjectionKindGitHubStatusComment: githubStatusCommentProjection{
			Package:       summary,
			Waves:         waves,
			ArtifactLinks: artifacts,
		},
	}

	result := make(map[enumtypes.ChangeGovernanceProjectionKind][]byte, len(payloads))
	for kind, payload := range payloads {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal change-governance projection %s: %w", kind, err)
		}
		result[kind] = raw
	}
	return result, nil
}

func buildProjectionPackageSummary(aggregate domainrepo.Aggregate) projectionPackageSummary {
	openGapCount := 0
	for _, item := range aggregate.FeedbackRecords {
		if item.State == enumtypes.ChangeGovernanceFeedbackRecordStateOpen || item.State == enumtypes.ChangeGovernanceFeedbackRecordStateAcknowledged {
			openGapCount++
		}
	}
	summary := projectionPackageSummary{
		PackageID:                 aggregate.Package.ID,
		RepositoryFullName:        aggregate.Package.RepositoryFullName,
		IssueNumber:               aggregate.Package.IssueNumber,
		PRNumber:                  aggregate.Package.PRNumber,
		BundleAdmissibility:       aggregate.Package.BundleAdmissibility,
		PublicationState:          aggregate.Package.PublicationState,
		EvidenceCompletenessState: aggregate.Package.EvidenceCompletenessState,
		VerificationMinimumState:  aggregate.Package.VerificationMinimumState,
		WaiverState:               aggregate.Package.WaiverState,
		ReleaseReadinessState:     aggregate.Package.ReleaseReadinessState,
		GovernanceFeedbackState:   aggregate.Package.GovernanceFeedbackState,
		OpenGapCount:              openGapCount,
		UpdatedAt:                 aggregate.Package.UpdatedAt,
	}
	if aggregate.Package.RiskTier != "" {
		summary.RiskTier = string(aggregate.Package.RiskTier)
	}
	return summary
}

func buildProjectionArtifactLinks(items []domainrepo.ArtifactLink) []projectionArtifactLink {
	result := make([]projectionArtifactLink, 0, len(items))
	for _, item := range items {
		result = append(result, projectionArtifactLink{
			ArtifactKind: string(item.ArtifactKind),
			ArtifactRef:  item.ArtifactRef,
			RelationKind: string(item.RelationKind),
			DisplayLabel: item.DisplayLabel,
		})
	}
	return result
}

func buildProjectionWaveItems(items []domainrepo.Wave) []projectionWaveItem {
	result := make([]projectionWaveItem, 0, len(items))
	for _, item := range items {
		result = append(result, projectionWaveItem{
			WaveKey:                   item.WaveKey,
			PublishOrder:              item.PublishOrder,
			DominantIntent:            item.DominantIntent,
			BoundedScopeKind:          item.BoundedScopeKind,
			PublicationState:          item.PublicationState,
			EvidenceCompletenessState: item.EvidenceCompletenessState,
			VerificationMinimumState:  item.VerificationMinimumState,
			Summary:                   item.Summary,
		})
	}
	return result
}

func buildProjectionEvidenceBlocks(waves []domainrepo.Wave, items []domainrepo.EvidenceBlock) []projectionEvidenceBlock {
	waveByID := make(map[string]string, len(waves))
	for _, wave := range waves {
		waveByID[wave.ID] = wave.WaveKey
	}

	result := make([]projectionEvidenceBlock, 0, len(items))
	for _, item := range items {
		scopeKind := enumtypes.ChangeGovernanceEvidenceScopeKindPackage
		scopeRef := item.PackageID
		if item.WaveID != "" {
			scopeKind = enumtypes.ChangeGovernanceEvidenceScopeKindWave
			if value, found := waveByID[item.WaveID]; found {
				scopeRef = value
			}
		}

		var links []projectionArtifactLink
		_ = json.Unmarshal(item.ArtifactLinksJSON, &links)
		result = append(result, projectionEvidenceBlock{
			BlockID:           item.ID,
			ScopeKind:         scopeKind,
			ScopeRef:          scopeRef,
			BlockKind:         item.BlockKind,
			State:             item.State,
			RequiredByTier:    item.RequiredByTier,
			VerificationState: item.VerificationState,
			ArtifactLinks:     links,
		})
	}
	return result
}

func buildProjectionDecisionSummaries(items []domainrepo.DecisionRecord) []projectionDecisionSummary {
	result := make([]projectionDecisionSummary, 0, len(items))
	for _, item := range items {
		summary := projectionDecisionSummary{
			DecisionKind: item.DecisionKind,
			State:        item.State,
			ActorKind:    item.ActorKind,
			RecordedAt:   item.RecordedAt,
			Summary:      item.SummaryMarkdown,
		}
		if item.ResidualRiskTier != "" {
			summary.ResidualRiskTier = string(item.ResidualRiskTier)
		}
		result = append(result, summary)
	}
	return result
}

func buildProjectionFeedbackRecords(items []domainrepo.FeedbackRecord) []projectionFeedbackRecord {
	result := make([]projectionFeedbackRecord, 0, len(items))
	for _, item := range items {
		result = append(result, projectionFeedbackRecord{
			GapID:           item.ID,
			GapKind:         item.GapKind,
			SourceKind:      item.SourceKind,
			Severity:        item.Severity,
			State:           item.State,
			SummaryMarkdown: item.SummaryMarkdown,
			SuggestedAction: item.SuggestedAction,
		})
	}
	return result
}

type derivedPackageState struct {
	PRNumber                  *int
	RiskTier                  enumtypes.ChangeGovernanceRiskTier
	BundleAdmissibility       enumtypes.ChangeGovernanceBundleAdmissibility
	PublicationState          enumtypes.ChangeGovernancePublicationState
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState
	WaiverState               enumtypes.ChangeGovernanceWaiverState
	ReleaseReadinessState     enumtypes.ChangeGovernanceReleaseReadinessState
	GovernanceFeedbackState   enumtypes.ChangeGovernanceFeedbackState
	LatestCorrelationID       string
}

func derivePackageState(aggregate domainrepo.Aggregate, correlationID string, prNumber *int) derivedPackageState {
	activeWaveItems := activeWaves(aggregate.Waves)
	activeEvidenceBlocks := filterEvidenceBlocksForActiveWaves(aggregate.Waves, aggregate.EvidenceBlocks)
	derived := derivedPackageState{
		PRNumber:                  prNumber,
		RiskTier:                  aggregate.Package.RiskTier,
		BundleAdmissibility:       aggregate.Package.BundleAdmissibility,
		PublicationState:          aggregate.Package.PublicationState,
		EvidenceCompletenessState: aggregate.Package.EvidenceCompletenessState,
		VerificationMinimumState:  aggregate.Package.VerificationMinimumState,
		WaiverState:               aggregate.Package.WaiverState,
		ReleaseReadinessState:     aggregate.Package.ReleaseReadinessState,
		GovernanceFeedbackState:   aggregate.Package.GovernanceFeedbackState,
		LatestCorrelationID:       strings.TrimSpace(correlationID),
	}

	if len(activeWaveItems) == 0 {
		derived.BundleAdmissibility = enumtypes.ChangeGovernanceBundleAdmissibilityRequiresDecomposition
		derived.PublicationState = enumtypes.ChangeGovernancePublicationStateHiddenDraft
	} else {
		derived.BundleAdmissibility = deriveBundleAdmissibility(activeWaveItems)
		if derived.BundleAdmissibility == enumtypes.ChangeGovernanceBundleAdmissibilityRequiresDecomposition {
			derived.PublicationState = enumtypes.ChangeGovernancePublicationStateWaveMapDefined
		} else {
			derived.PublicationState = enumtypes.ChangeGovernancePublicationStateWavesPublished
		}
	}

	derived.EvidenceCompletenessState = deriveEvidenceCompletenessState(activeEvidenceBlocks)
	derived.VerificationMinimumState = deriveVerificationMinimumState(activeEvidenceBlocks)
	derived.GovernanceFeedbackState = deriveFeedbackState(aggregate.FeedbackRecords)
	return derived
}

func deriveBundleAdmissibility(waves []domainrepo.Wave) enumtypes.ChangeGovernanceBundleAdmissibility {
	if len(waves) == 1 {
		wave := waves[0]
		if wave.BoundedScopeKind == enumtypes.ChangeGovernanceBoundedScopeKindMechanicalBoundedScope &&
			(wave.DominantIntent == enumtypes.ChangeGovernanceDominantIntentMechanicalRefactor || wave.DominantIntent == enumtypes.ChangeGovernanceDominantIntentDocsOnly) {
			return enumtypes.ChangeGovernanceBundleAdmissibilityMechanicalBoundedScope
		}
		if wave.BoundedScopeKind == enumtypes.ChangeGovernanceBoundedScopeKindSingleContext {
			return enumtypes.ChangeGovernanceBundleAdmissibilitySingleWave
		}
	}
	return enumtypes.ChangeGovernanceBundleAdmissibilityRequiresDecomposition
}

func deriveEvidenceCompletenessState(items []domainrepo.EvidenceBlock) enumtypes.ChangeGovernanceEvidenceCompletenessState {
	if len(items) == 0 {
		return enumtypes.ChangeGovernanceEvidenceCompletenessStateNotStarted
	}
	allWaived := true
	allSatisfied := true
	anySatisfied := false
	anyMissing := false
	for _, item := range items {
		switch item.State {
		case enumtypes.ChangeGovernanceEvidenceBlockStateVerified, enumtypes.ChangeGovernanceEvidenceBlockStatePresent:
			allWaived = false
			anySatisfied = true
		case enumtypes.ChangeGovernanceEvidenceBlockStateWaived:
			anySatisfied = true
		case enumtypes.ChangeGovernanceEvidenceBlockStateMissing, enumtypes.ChangeGovernanceEvidenceBlockStateStale:
			allSatisfied = false
			allWaived = false
			anyMissing = true
		default:
			allSatisfied = false
			allWaived = false
		}
	}
	if anyMissing {
		return enumtypes.ChangeGovernanceEvidenceCompletenessStateGapped
	}
	if allWaived {
		return enumtypes.ChangeGovernanceEvidenceCompletenessStateWaived
	}
	if allSatisfied && anySatisfied {
		return enumtypes.ChangeGovernanceEvidenceCompletenessStateComplete
	}
	return enumtypes.ChangeGovernanceEvidenceCompletenessStatePartial
}

func deriveVerificationMinimumState(items []domainrepo.EvidenceBlock) enumtypes.ChangeGovernanceVerificationMinimumState {
	if len(items) == 0 {
		return enumtypes.ChangeGovernanceVerificationMinimumStateNotStarted
	}
	anyInProgress := false
	anyMet := false
	allDone := true
	for _, item := range items {
		switch item.VerificationState {
		case enumtypes.ChangeGovernanceVerificationMinimumStateFailed:
			return enumtypes.ChangeGovernanceVerificationMinimumStateFailed
		case enumtypes.ChangeGovernanceVerificationMinimumStateInProgress:
			anyInProgress = true
			allDone = false
		case enumtypes.ChangeGovernanceVerificationMinimumStateMet, enumtypes.ChangeGovernanceVerificationMinimumStateWaived:
			anyMet = true
		default:
			allDone = false
		}
	}
	if anyInProgress {
		return enumtypes.ChangeGovernanceVerificationMinimumStateInProgress
	}
	if anyMet && allDone {
		return enumtypes.ChangeGovernanceVerificationMinimumStateMet
	}
	if anyMet {
		return enumtypes.ChangeGovernanceVerificationMinimumStateInProgress
	}
	return enumtypes.ChangeGovernanceVerificationMinimumStateNotStarted
}

func deriveFeedbackState(items []domainrepo.FeedbackRecord) enumtypes.ChangeGovernanceFeedbackState {
	if len(items) == 0 {
		return enumtypes.ChangeGovernanceFeedbackStateNone
	}
	for _, item := range items {
		switch item.State {
		case enumtypes.ChangeGovernanceFeedbackRecordStateOpen, enumtypes.ChangeGovernanceFeedbackRecordStateAcknowledged:
			return enumtypes.ChangeGovernanceFeedbackStateOpen
		case enumtypes.ChangeGovernanceFeedbackRecordStateReclassified:
			return enumtypes.ChangeGovernanceFeedbackStateReclassified
		}
	}
	return enumtypes.ChangeGovernanceFeedbackStateClosed
}

type draftMetadata struct {
	ChangeScopeHints     []querytypes.ChangeGovernanceScopeHint `json:"change_scope_hints"`
	CandidateRiskDrivers []enumtypes.ChangeGovernanceRiskDriver `json:"candidate_risk_drivers,omitempty"`
	BranchName           string                                 `json:"branch_name,omitempty"`
}

type draftPackageUpsertedEventPayload struct {
	PackageID        string                                     `json:"package_id"`
	PackageKey       string                                     `json:"package_key"`
	DraftSignalID    string                                     `json:"draft_signal_id"`
	PublicationState enumtypes.ChangeGovernancePublicationState `json:"publication_state"`
}

type waveMapPublishedEventPayload struct {
	PackageID         string                                     `json:"package_id"`
	WaveMapID         string                                     `json:"wave_map_id"`
	PublicationState  enumtypes.ChangeGovernancePublicationState `json:"publication_state"`
	ProjectionVersion int64                                      `json:"projection_version"`
}

type projectionRefreshedEventPayload struct {
	PackageID                 string                                              `json:"package_id"`
	ScopeKind                 enumtypes.ChangeGovernanceEvidenceScopeKind         `json:"scope_kind"`
	ScopeRef                  string                                              `json:"scope_ref"`
	BlockKind                 enumtypes.ChangeGovernanceEvidenceBlockKind         `json:"block_kind"`
	ProjectionVersion         int64                                               `json:"projection_version"`
	EvidenceCompletenessState enumtypes.ChangeGovernanceEvidenceCompletenessState `json:"evidence_completeness_state"`
	VerificationMinimumState  enumtypes.ChangeGovernanceVerificationMinimumState  `json:"verification_minimum_state"`
}

func evidenceBlockStateFromSignal(params querytypes.ChangeGovernanceEvidenceSignalParams) enumtypes.ChangeGovernanceEvidenceBlockState {
	switch verificationStateFromHint(params.VerificationStateHint) {
	case enumtypes.ChangeGovernanceVerificationMinimumStateMet:
		return enumtypes.ChangeGovernanceEvidenceBlockStateVerified
	case enumtypes.ChangeGovernanceVerificationMinimumStateWaived:
		return enumtypes.ChangeGovernanceEvidenceBlockStateWaived
	default:
		if len(params.ArtifactLinks) == 0 {
			return enumtypes.ChangeGovernanceEvidenceBlockStateMissing
		}
		return enumtypes.ChangeGovernanceEvidenceBlockStatePresent
	}
}

func verificationStateFromHint(value enumtypes.ChangeGovernanceVerificationMinimumState) enumtypes.ChangeGovernanceVerificationMinimumState {
	switch value {
	case enumtypes.ChangeGovernanceVerificationMinimumStateInProgress,
		enumtypes.ChangeGovernanceVerificationMinimumStateMet,
		enumtypes.ChangeGovernanceVerificationMinimumStateFailed,
		enumtypes.ChangeGovernanceVerificationMinimumStateWaived:
		return value
	default:
		return enumtypes.ChangeGovernanceVerificationMinimumStateNotStarted
	}
}

func (r *Repository) insertFlowEvent(ctx context.Context, tx pgx.Tx, correlationID string, actorID floweventdomain.ActorID, eventType floweventdomain.EventType, payload any) error {
	raw := marshalJSONPayload(payload)
	if _, err := tx.Exec(
		ctx,
		queryInsertFlowEvent,
		strings.TrimSpace(correlationID),
		string(floweventdomain.ActorTypeSystem),
		string(actorID),
		string(eventType),
		raw,
		time.Now().UTC(),
	); err != nil {
		return fmt.Errorf("insert change-governance flow event: %w", err)
	}
	return nil
}

func buildPackageKey(repositoryFullName string, issueNumber int) string {
	return fmt.Sprintf("%s#%d", strings.TrimSpace(repositoryFullName), issueNumber)
}

func marshalJSONPayload(payload any) []byte {
	raw, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func nullableText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func intPtrToPGInt4(value *int) pgtype.Int4 {
	if value == nil || *value <= 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func jsonOrEmptyObject(value []byte) []byte {
	if len(value) == 0 {
		return []byte(`{}`)
	}
	return value
}

func jsonOrEmptyArray(value []byte) []byte {
	if len(value) == 0 {
		return []byte(`[]`)
	}
	return value
}

func timestamptzOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
