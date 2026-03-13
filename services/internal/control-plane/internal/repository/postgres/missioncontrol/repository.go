package missioncontrol

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/missioncontrol/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/upsert_entity.sql
	queryUpsertEntity string
	//go:embed sql/update_entity_projection.sql
	queryUpdateEntityProjection string
	//go:embed sql/get_entity_by_public_id.sql
	queryGetEntityByPublicID string
	//go:embed sql/list_entities.sql
	queryListEntities string
	//go:embed sql/delete_relations_for_source.sql
	queryDeleteRelationsForSource string
	//go:embed sql/insert_relation.sql
	queryInsertRelation string
	//go:embed sql/list_relations_for_entity.sql
	queryListRelationsForEntity string
	//go:embed sql/upsert_timeline_entry.sql
	queryUpsertTimelineEntry string
	//go:embed sql/list_timeline_entries.sql
	queryListTimelineEntries string
	//go:embed sql/insert_command.sql
	queryInsertCommand string
	//go:embed sql/get_command_by_id.sql
	queryGetCommandByID string
	//go:embed sql/list_commands.sql
	queryListCommands string
	//go:embed sql/update_command_status.sql
	queryUpdateCommandStatus string
	//go:embed sql/get_warmup_summary.sql
	queryGetWarmupSummary string
)

// Repository persists Mission Control foundation state in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs Mission Control PostgreSQL repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// UpsertEntity stores one projection row.
func (r *Repository) UpsertEntity(ctx context.Context, params domainrepo.UpsertEntityParams) (domainrepo.Entity, error) {
	normalized := normalizeEntityUpsertParams(params)
	rows, err := r.db.Query(
		ctx,
		queryUpsertEntity,
		normalized.ProjectID,
		string(normalized.EntityKind),
		normalized.EntityExternalKey,
		string(normalized.ProviderKind),
		nullableText(normalized.ProviderURL),
		normalized.Title,
		string(normalized.ActiveState),
		normalized.ProjectionVersion,
		jsonOrEmptyObject(normalized.CardPayloadJSON),
		jsonOrEmptyObject(normalized.DetailPayloadJSON),
		timestamptzPtrOrNil(normalized.LastTimelineAt),
		timestamptzPtrOrNil(normalized.ProviderUpdatedAt),
		timestamptzOrNil(normalized.ProjectedAt),
		timestamptzPtrOrNil(normalized.StaleAfter),
		string(normalized.SyncStatus),
	)
	if err != nil {
		return domainrepo.Entity{}, fmt.Errorf("upsert mission control entity: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.EntityRow])
	if err != nil {
		return domainrepo.Entity{}, fmt.Errorf("collect mission control entity upsert: %w", err)
	}
	return fromEntityRow(row), nil
}

// UpdateEntityProjection stores one projection update guarded by projection_version.
func (r *Repository) UpdateEntityProjection(ctx context.Context, params domainrepo.UpdateEntityParams) (domainrepo.Entity, error) {
	normalized := normalizeEntityUpdateParams(params)
	rows, err := r.db.Query(
		ctx,
		queryUpdateEntityProjection,
		normalized.ProjectID,
		string(normalized.EntityKind),
		normalized.EntityExternalKey,
		nullableText(normalized.ProviderURL),
		normalized.Title,
		string(normalized.ActiveState),
		string(normalized.SyncStatus),
		jsonOrEmptyObject(normalized.CardPayloadJSON),
		jsonOrEmptyObject(normalized.DetailPayloadJSON),
		timestamptzPtrOrNil(normalized.LastTimelineAt),
		timestamptzPtrOrNil(normalized.ProviderUpdatedAt),
		timestamptzOrNil(normalized.ProjectedAt),
		timestamptzPtrOrNil(normalized.StaleAfter),
		normalized.ExpectedProjectionVersion,
	)
	if err != nil {
		return domainrepo.Entity{}, fmt.Errorf("update mission control entity projection: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.EntityRow])
	if err == nil {
		return fromEntityRow(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.Entity{}, fmt.Errorf("collect mission control entity projection update: %w", err)
	}

	current, found, lookupErr := r.GetEntityByPublicID(ctx, normalized.ProjectID, normalized.EntityKind, normalized.EntityExternalKey)
	if lookupErr != nil {
		return domainrepo.Entity{}, lookupErr
	}
	if !found {
		return domainrepo.Entity{}, errs.NotFound{Msg: "mission control entity not found"}
	}
	return domainrepo.Entity{}, domainrepo.ProjectionVersionConflict{
		ProjectID:                 normalized.ProjectID,
		EntityKind:                normalized.EntityKind,
		EntityExternalKey:         normalized.EntityExternalKey,
		ExpectedProjectionVersion: normalized.ExpectedProjectionVersion,
		ActualProjectionVersion:   current.ProjectionVersion,
	}
}

// GetEntityByPublicID loads one entity by public identity tuple.
func (r *Repository) GetEntityByPublicID(ctx context.Context, projectID string, entityKind enumtypes.MissionControlEntityKind, entityExternalKey string) (domainrepo.Entity, bool, error) {
	rows, err := r.db.Query(ctx, queryGetEntityByPublicID, strings.TrimSpace(projectID), string(entityKind), strings.TrimSpace(entityExternalKey))
	if err != nil {
		return domainrepo.Entity{}, false, fmt.Errorf("query mission control entity by public id: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.EntityRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Entity{}, false, nil
		}
		return domainrepo.Entity{}, false, fmt.Errorf("collect mission control entity by public id: %w", err)
	}
	return fromEntityRow(row), true, nil
}

// ListEntities returns projection rows for one project with optional filters.
func (r *Repository) ListEntities(ctx context.Context, filter domainrepo.EntityListFilter) ([]domainrepo.Entity, error) {
	normalized := normalizeEntityListFilter(filter)
	rows, err := r.db.Query(
		ctx,
		queryListEntities,
		normalized.ProjectID,
		activeStatesToStrings(normalized.ActiveStates),
		syncStatusesToStrings(normalized.SyncStatuses),
		normalized.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list mission control entities: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.EntityRow])
	if err != nil {
		return nil, fmt.Errorf("collect mission control entities: %w", err)
	}
	out := make([]domainrepo.Entity, 0, len(items))
	for _, item := range items {
		out = append(out, fromEntityRow(item))
	}
	return out, nil
}

// ReplaceRelationsForSource rewrites edges for one source entity.
func (r *Repository) ReplaceRelationsForSource(ctx context.Context, params domainrepo.ReplaceRelationsParams) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin mission control relation replace: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, queryDeleteRelationsForSource, strings.TrimSpace(params.ProjectID), params.SourceEntityID); err != nil {
		return fmt.Errorf("delete mission control relations: %w", err)
	}

	if len(params.Relations) > 0 {
		batch := &pgx.Batch{}
		for _, relation := range params.Relations {
			batch.Queue(
				queryInsertRelation,
				strings.TrimSpace(params.ProjectID),
				params.SourceEntityID,
				string(relation.RelationKind),
				relation.TargetEntityID,
				string(relation.SourceKind),
			)
		}
		results := tx.SendBatch(ctx, batch)
		for range params.Relations {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return fmt.Errorf("insert mission control relation: %w", err)
			}
		}
		if err := results.Close(); err != nil {
			return fmt.Errorf("close mission control relation batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit mission control relation replace: %w", err)
	}
	return nil
}

// ListRelationsForEntity returns all edges where one entity participates.
func (r *Repository) ListRelationsForEntity(ctx context.Context, projectID string, entityID int64) ([]domainrepo.Relation, error) {
	rows, err := r.db.Query(ctx, queryListRelationsForEntity, strings.TrimSpace(projectID), entityID)
	if err != nil {
		return nil, fmt.Errorf("list mission control relations: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RelationRow])
	if err != nil {
		return nil, fmt.Errorf("collect mission control relations: %w", err)
	}
	out := make([]domainrepo.Relation, 0, len(items))
	for _, item := range items {
		out = append(out, fromRelationRow(item))
	}
	return out, nil
}

// UpsertTimelineEntry stores one timeline projection row.
func (r *Repository) UpsertTimelineEntry(ctx context.Context, params domainrepo.UpsertTimelineEntryParams) (domainrepo.TimelineEntry, error) {
	normalized := normalizeTimelineEntryParams(params)
	rows, err := r.db.Query(
		ctx,
		queryUpsertTimelineEntry,
		strings.TrimSpace(normalized.ProjectID),
		normalized.EntityID,
		string(normalized.SourceKind),
		normalized.EntryExternalKey,
		nullableUUID(normalized.CommandID),
		normalized.Summary,
		nullableText(normalized.BodyMarkdown),
		jsonOrEmptyObject(normalized.PayloadJSON),
		timestamptzOrNil(normalized.OccurredAt),
		nullableText(normalized.ProviderURL),
		normalized.IsReadOnly,
	)
	if err != nil {
		return domainrepo.TimelineEntry{}, fmt.Errorf("upsert mission control timeline entry: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.TimelineEntryRow])
	if err != nil {
		return domainrepo.TimelineEntry{}, fmt.Errorf("collect mission control timeline entry upsert: %w", err)
	}
	return fromTimelineEntryRow(row), nil
}

// ListTimelineEntries returns one entity timeline ordered newest first.
func (r *Repository) ListTimelineEntries(ctx context.Context, filter domainrepo.TimelineListFilter) ([]domainrepo.TimelineEntry, error) {
	normalized := normalizeTimelineListFilter(filter)
	rows, err := r.db.Query(ctx, queryListTimelineEntries, normalized.ProjectID, normalized.EntityID, normalized.Limit)
	if err != nil {
		return nil, fmt.Errorf("list mission control timeline entries: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.TimelineEntryRow])
	if err != nil {
		return nil, fmt.Errorf("collect mission control timeline entries: %w", err)
	}
	out := make([]domainrepo.TimelineEntry, 0, len(items))
	for _, item := range items {
		out = append(out, fromTimelineEntryRow(item))
	}
	return out, nil
}

// CreateCommand inserts one command-ledger row.
func (r *Repository) CreateCommand(ctx context.Context, params domainrepo.CreateCommandParams) (domainrepo.Command, error) {
	normalized := normalizeCommandCreateParams(params)
	rows, err := r.db.Query(
		ctx,
		queryInsertCommand,
		strings.TrimSpace(normalized.ProjectID),
		string(normalized.CommandKind),
		nullableInt64Ptr(normalized.TargetEntityID),
		normalized.ActorID,
		normalized.BusinessIntentKey,
		normalized.CorrelationID,
		string(normalized.Status),
		nullableFailureReason(normalized.FailureReason),
		nullableUUID(normalized.ApprovalRequestID),
		string(normalized.ApprovalState),
		timestamptzPtrOrNil(normalized.ApprovalRequestedAt),
		timestamptzPtrOrNil(normalized.ApprovalDecidedAt),
		jsonOrEmptyObject(normalized.PayloadJSON),
		jsonOrEmptyObject(normalized.ResultPayloadJSON),
		jsonOrEmptyArray(normalized.ProviderDeliveries),
		timestamptzOrNil(normalized.RequestedAt),
		timestamptzOrNil(normalized.UpdatedAt),
		timestamptzPtrOrNil(normalized.ReconciledAt),
	)
	if err != nil {
		return domainrepo.Command{}, fmt.Errorf("insert mission control command: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.CommandRow])
	if err != nil {
		return domainrepo.Command{}, fmt.Errorf("collect mission control command insert: %w", err)
	}
	return fromCommandRow(row), nil
}

// GetCommandByID loads one command row scoped to one project.
func (r *Repository) GetCommandByID(ctx context.Context, projectID string, commandID string) (domainrepo.Command, bool, error) {
	rows, err := r.db.Query(ctx, queryGetCommandByID, strings.TrimSpace(projectID), strings.TrimSpace(commandID))
	if err != nil {
		return domainrepo.Command{}, false, fmt.Errorf("query mission control command by id: %w", err)
	}
	return collectOptionalCommand(rows, "collect mission control command by id")
}

// ListCommands returns command rows for one project.
func (r *Repository) ListCommands(ctx context.Context, filter domainrepo.CommandListFilter) ([]domainrepo.Command, error) {
	normalized := normalizeCommandListFilter(filter)
	rows, err := r.db.Query(ctx, queryListCommands, normalized.ProjectID, commandStatusesToStrings(normalized.Statuses), normalized.Limit)
	if err != nil {
		return nil, fmt.Errorf("list mission control commands: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.CommandRow])
	if err != nil {
		return nil, fmt.Errorf("collect mission control commands: %w", err)
	}
	out := make([]domainrepo.Command, 0, len(items))
	for _, item := range items {
		out = append(out, fromCommandRow(item))
	}
	return out, nil
}

// UpdateCommandStatus persists one command transition.
func (r *Repository) UpdateCommandStatus(ctx context.Context, params domainrepo.UpdateCommandStatusParams) (domainrepo.Command, bool, error) {
	normalized := normalizeCommandStatusUpdateParams(params)
	failureReasonSet, failureReasonValue := failureReasonPatchArg(normalized.FailureReasonPatch)
	approvalRequestIDSet, approvalRequestIDValue := uuidStringPatchArg(normalized.ApprovalRequestIDPatch)
	approvalStateSet, approvalStateValue := approvalStatePatchArg(normalized.ApprovalStatePatch)
	approvalRequestedAtSet, approvalRequestedAtValue := timestamptzPatchArg(normalized.ApprovalRequestedAtPatch)
	approvalDecidedAtSet, approvalDecidedAtValue := timestamptzPatchArg(normalized.ApprovalDecidedAtPatch)
	resultPayloadSet, resultPayloadValue := jsonPatchArg(normalized.ResultPayloadPatch, jsonOrEmptyObject)
	providerDeliveriesSet, providerDeliveriesValue := jsonPatchArg(normalized.ProviderDeliveriesPatch, jsonOrEmptyArray)
	reconciledAtSet, reconciledAtValue := timestamptzPatchArg(normalized.ReconciledAtPatch)
	rows, err := r.db.Query(
		ctx,
		queryUpdateCommandStatus,
		normalized.ProjectID,
		normalized.CommandID,
		string(normalized.Status),
		failureReasonSet,
		failureReasonValue,
		approvalRequestIDSet,
		approvalRequestIDValue,
		approvalStateSet,
		approvalStateValue,
		approvalRequestedAtSet,
		approvalRequestedAtValue,
		approvalDecidedAtSet,
		approvalDecidedAtValue,
		resultPayloadSet,
		resultPayloadValue,
		providerDeliveriesSet,
		providerDeliveriesValue,
		timestamptzOrNil(normalized.UpdatedAt),
		reconciledAtSet,
		reconciledAtValue,
	)
	if err != nil {
		return domainrepo.Command{}, false, fmt.Errorf("update mission control command status: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.CommandRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Command{}, false, nil
		}
		return domainrepo.Command{}, false, fmt.Errorf("collect mission control command status update: %w", err)
	}
	return fromCommandRow(row), true, nil
}

// GetWarmupSummary returns aggregate counts for worker warmup verification.
func (r *Repository) GetWarmupSummary(ctx context.Context, projectID string) (domainrepo.WarmupSummary, error) {
	rows, err := r.db.Query(ctx, queryGetWarmupSummary, strings.TrimSpace(projectID))
	if err != nil {
		return domainrepo.WarmupSummary{}, fmt.Errorf("query mission control warmup summary: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.WarmupSummaryRow])
	if err != nil {
		return domainrepo.WarmupSummary{}, fmt.Errorf("collect mission control warmup summary: %w", err)
	}
	return fromWarmupSummaryRow(row), nil
}

func normalizeEntityUpsertParams(params domainrepo.UpsertEntityParams) domainrepo.UpsertEntityParams {
	normalized := params
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.EntityExternalKey = strings.TrimSpace(normalized.EntityExternalKey)
	normalized.ProviderURL = strings.TrimSpace(normalized.ProviderURL)
	normalized.Title = strings.TrimSpace(normalized.Title)
	if normalized.ProviderKind == "" {
		normalized.ProviderKind = enumtypes.MissionControlProviderKindGitHub
	}
	if normalized.ActiveState == "" {
		normalized.ActiveState = enumtypes.MissionControlActiveStateWorking
	}
	if normalized.SyncStatus == "" {
		normalized.SyncStatus = enumtypes.MissionControlSyncStatusSynced
	}
	if normalized.ProjectionVersion <= 0 {
		normalized.ProjectionVersion = 1
	}
	if normalized.ProjectedAt.IsZero() {
		normalized.ProjectedAt = time.Now().UTC()
	}
	return normalized
}

func normalizeEntityUpdateParams(params domainrepo.UpdateEntityParams) domainrepo.UpdateEntityParams {
	normalized := params
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.EntityExternalKey = strings.TrimSpace(normalized.EntityExternalKey)
	normalized.ProviderURL = strings.TrimSpace(normalized.ProviderURL)
	normalized.Title = strings.TrimSpace(normalized.Title)
	if normalized.SyncStatus == "" {
		normalized.SyncStatus = enumtypes.MissionControlSyncStatusSynced
	}
	if normalized.ProjectedAt.IsZero() {
		normalized.ProjectedAt = time.Now().UTC()
	}
	return normalized
}

func normalizeEntityListFilter(filter domainrepo.EntityListFilter) domainrepo.EntityListFilter {
	normalized := filter
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.Limit = normalizeLimit(normalized.Limit)
	return normalized
}

func normalizeTimelineEntryParams(params domainrepo.UpsertTimelineEntryParams) domainrepo.UpsertTimelineEntryParams {
	normalized := params
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.EntryExternalKey = strings.TrimSpace(normalized.EntryExternalKey)
	normalized.CommandID = strings.TrimSpace(normalized.CommandID)
	normalized.Summary = strings.TrimSpace(normalized.Summary)
	normalized.BodyMarkdown = strings.TrimSpace(normalized.BodyMarkdown)
	normalized.ProviderURL = strings.TrimSpace(normalized.ProviderURL)
	if normalized.OccurredAt.IsZero() {
		normalized.OccurredAt = time.Now().UTC()
	}
	return normalized
}

func normalizeTimelineListFilter(filter domainrepo.TimelineListFilter) domainrepo.TimelineListFilter {
	normalized := filter
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.Limit = normalizeLimit(normalized.Limit)
	return normalized
}

func normalizeCommandCreateParams(params domainrepo.CreateCommandParams) domainrepo.CreateCommandParams {
	normalized := params
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.ActorID = strings.TrimSpace(normalized.ActorID)
	normalized.BusinessIntentKey = strings.TrimSpace(normalized.BusinessIntentKey)
	normalized.CorrelationID = strings.TrimSpace(normalized.CorrelationID)
	normalized.ApprovalRequestID = strings.TrimSpace(normalized.ApprovalRequestID)
	if normalized.Status == "" {
		normalized.Status = enumtypes.MissionControlCommandStatusAccepted
	}
	if normalized.ApprovalState == "" {
		normalized.ApprovalState = enumtypes.MissionControlApprovalStateNotRequired
	}
	if normalized.RequestedAt.IsZero() {
		normalized.RequestedAt = time.Now().UTC()
	}
	if normalized.UpdatedAt.IsZero() {
		normalized.UpdatedAt = normalized.RequestedAt
	}
	return normalized
}

func normalizeCommandListFilter(filter domainrepo.CommandListFilter) domainrepo.CommandListFilter {
	normalized := filter
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.Limit = normalizeLimit(normalized.Limit)
	return normalized
}

func normalizeCommandStatusUpdateParams(params domainrepo.UpdateCommandStatusParams) domainrepo.UpdateCommandStatusParams {
	normalized := params
	normalized.ProjectID = strings.TrimSpace(normalized.ProjectID)
	normalized.CommandID = strings.TrimSpace(normalized.CommandID)
	normalized.ApprovalRequestIDPatch = normalizeStringPatch(normalized.ApprovalRequestIDPatch)
	normalized.ApprovalRequestedAtPatch = normalizeTimePatch(normalized.ApprovalRequestedAtPatch)
	normalized.ApprovalDecidedAtPatch = normalizeTimePatch(normalized.ApprovalDecidedAtPatch)
	normalized.ReconciledAtPatch = normalizeTimePatch(normalized.ReconciledAtPatch)
	if normalized.ApprovalStatePatch.Set && normalized.ApprovalStatePatch.Value == "" {
		normalized.ApprovalStatePatch.Value = enumtypes.MissionControlApprovalStateNotRequired
	}
	if normalized.UpdatedAt.IsZero() {
		normalized.UpdatedAt = time.Now().UTC()
	}
	return normalized
}

func normalizeStringPatch(patch domainrepo.OptionalStringPatch) domainrepo.OptionalStringPatch {
	normalized := patch
	if normalized.Set {
		normalized.Value = strings.TrimSpace(normalized.Value)
	}
	return normalized
}

func normalizeTimePatch(patch domainrepo.OptionalTimePatch) domainrepo.OptionalTimePatch {
	normalized := patch
	if normalized.Set && normalized.Value != nil {
		value := normalized.Value.UTC()
		normalized.Value = &value
	}
	return normalized
}

func failureReasonPatchArg(patch domainrepo.CommandFailureReasonPatch) (bool, any) {
	if !patch.Set {
		return false, nil
	}
	return true, nullableFailureReason(patch.Value)
}

func uuidStringPatchArg(patch domainrepo.OptionalStringPatch) (bool, any) {
	if !patch.Set {
		return false, nil
	}
	return true, nullableUUID(patch.Value)
}

func approvalStatePatchArg(patch domainrepo.CommandApprovalStatePatch) (bool, any) {
	if !patch.Set {
		return false, nil
	}
	return true, string(patch.Value)
}

func timestamptzPatchArg(patch domainrepo.OptionalTimePatch) (bool, pgtype.Timestamptz) {
	if !patch.Set {
		return false, pgtype.Timestamptz{}
	}
	return true, timestamptzPtrOrNil(patch.Value)
}

func jsonPatchArg(patch domainrepo.OptionalJSONPatch, normalize func([]byte) []byte) (bool, []byte) {
	if !patch.Set {
		return false, nil
	}
	return true, normalize(patch.Value)
}

func nullableText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableInt64Ptr(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func nullableFailureReason(value enumtypes.MissionControlCommandFailureReason) any {
	if strings.TrimSpace(string(value)) == "" {
		return nil
	}
	return string(value)
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

func activeStatesToStrings(values []enumtypes.MissionControlActiveState) []string {
	return enumStrings(values)
}

func syncStatusesToStrings(values []enumtypes.MissionControlSyncStatus) []string {
	return enumStrings(values)
}

func commandStatusesToStrings(values []enumtypes.MissionControlCommandStatus) []string {
	return enumStrings(values)
}

func collectOptionalCommand(rows pgx.Rows, op string) (domainrepo.Command, bool, error) {
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.CommandRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Command{}, false, nil
		}
		return domainrepo.Command{}, false, fmt.Errorf("%s: %w", op, err)
	}
	return fromCommandRow(row), true, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func enumStrings[T ~string](values []T) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(string(value)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
