package interactionrequest

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/interactionrequest/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/create.sql
	queryCreate string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/find_open_decision_by_run_id.sql
	queryFindOpenDecisionByRunID string
	//go:embed sql/select_for_update.sql
	querySelectForUpdate string
	//go:embed sql/create_delivery_attempt.sql
	queryCreateDeliveryAttempt string
	//go:embed sql/update_last_delivery_attempt_no.sql
	queryUpdateLastDeliveryAttemptNo string
	//go:embed sql/get_callback_event_by_key.sql
	queryGetCallbackEventByKey string
	//go:embed sql/insert_callback_event.sql
	queryInsertCallbackEvent string
	//go:embed sql/insert_response_record.sql
	queryInsertResponseRecord string
	//go:embed sql/update_request_state.sql
	queryUpdateRequestState string
)

// Repository persists interaction aggregate and callback evidence in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL interaction repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts one interaction aggregate row.
func (r *Repository) Create(ctx context.Context, params domainrepo.CreateParams) (domainrepo.Request, error) {
	row := r.db.QueryRow(
		ctx,
		queryCreate,
		strings.TrimSpace(params.ProjectID),
		strings.TrimSpace(params.RunID),
		string(params.InteractionKind),
		string(params.State),
		string(params.ResolutionKind),
		strings.TrimSpace(params.RecipientProvider),
		strings.TrimSpace(params.RecipientRef),
		jsonOrEmptyObject(params.RequestPayloadJSON),
		jsonOrEmptyObject(params.ContextLinksJSON),
		timestamptzPtrOrNil(params.ResponseDeadlineAt),
	)

	item, err := scanRequestRow(row)
	if err != nil {
		return domainrepo.Request{}, fmt.Errorf("create interaction request: %w", err)
	}
	return item, nil
}

// GetByID returns one interaction aggregate by id.
func (r *Repository) GetByID(ctx context.Context, interactionID string) (domainrepo.Request, bool, error) {
	return r.lookupRequest(ctx, queryGetByID, strings.TrimSpace(interactionID), "interaction request by id")
}

// FindOpenDecisionByRunID returns open decision interaction for one run when present.
func (r *Repository) FindOpenDecisionByRunID(ctx context.Context, runID string) (domainrepo.Request, bool, error) {
	return r.lookupRequest(ctx, queryFindOpenDecisionByRunID, strings.TrimSpace(runID), "open decision interaction by run id")
}

func (r *Repository) lookupRequest(ctx context.Context, query string, argument string, operation string) (domainrepo.Request, bool, error) {
	rows, err := r.db.Query(ctx, query, argument)
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("query %s: %w", operation, err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RequestRow])
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("collect %s: %w", operation, err)
	}
	if len(items) == 0 {
		return domainrepo.Request{}, false, nil
	}
	return requestFromDBModel(items[0]), true, nil
}

// CreateDeliveryAttempt appends one dispatch-attempt ledger row and bumps aggregate counter.
func (r *Repository) CreateDeliveryAttempt(ctx context.Context, params domainrepo.CreateDeliveryAttemptParams) (domainrepo.DeliveryAttempt, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("begin create interaction delivery attempt tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, found, err := r.getRequestForUpdate(ctx, tx, params.InteractionID)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, err
	}
	if !found {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("interaction request not found")
	}

	nextAttemptNo := request.LastDeliveryAttemptNo + 1
	row := tx.QueryRow(
		ctx,
		queryCreateDeliveryAttempt,
		request.ID,
		nextAttemptNo,
		strings.TrimSpace(params.AdapterKind),
		string(params.Status),
		jsonOrEmptyObject(params.RequestEnvelopeJSON),
		jsonOrEmptyObject(params.AckPayloadJSON),
		nullableText(params.AdapterDeliveryID),
		params.Retryable,
		timestamptzPtrOrNil(params.NextRetryAt),
		nullableText(params.LastErrorCode),
		timeOrNow(params.StartedAt),
		timestamptzPtrOrNil(params.FinishedAt),
	)

	attempt, err := scanDeliveryAttemptRow(row)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("create interaction delivery attempt: %w", err)
	}
	if _, err := tx.Exec(ctx, queryUpdateLastDeliveryAttemptNo, request.ID, nextAttemptNo); err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("update last interaction delivery attempt no: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("commit interaction delivery attempt tx: %w", err)
	}
	return attempt, nil
}

// ApplyCallback persists callback evidence, optional typed response and terminal aggregate mutation.
func (r *Repository) ApplyCallback(ctx context.Context, params domainrepo.ApplyCallbackParams) (domainrepo.ApplyCallbackResult, error) {
	now := params.OccurredAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("begin apply interaction callback tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, found, err := r.getRequestForUpdate(ctx, tx, params.InteractionID)
	if err != nil {
		return domainrepo.ApplyCallbackResult{}, err
	}
	if !found {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("interaction request not found")
	}

	existingEvent, found, err := r.getCallbackEventByKey(ctx, tx, request.ID, params.AdapterEventID)
	if err != nil {
		return domainrepo.ApplyCallbackResult{}, err
	}
	if found {
		if err := tx.Commit(ctx); err != nil {
			return domainrepo.ApplyCallbackResult{}, fmt.Errorf("commit duplicate interaction callback tx: %w", err)
		}
		return domainrepo.ApplyCallbackResult{
			Interaction:    request,
			CallbackEvent:  existingEvent,
			Classification: enumtypes.InteractionCallbackResultClassificationDuplicate,
		}, nil
	}

	decision := classifyCallback(request, params, now)
	callbackEventRows, err := tx.Query(
		ctx,
		queryInsertCallbackEvent,
		request.ID,
		nullableUUID(params.DeliveryID),
		strings.TrimSpace(params.AdapterEventID),
		string(params.CallbackKind),
		string(decision.persistedClassification),
		jsonOrEmptyObject(params.NormalizedPayloadJSON),
		jsonOrEmptyObject(params.RawPayloadJSON),
		now,
		now,
	)
	if err != nil {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("insert interaction callback event: %w", err)
	}
	callbackEventRow, err := pgx.CollectOneRow(callbackEventRows, pgx.RowToStructByName[dbmodel.CallbackEventRow])
	if err != nil {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("collect interaction callback event: %w", err)
	}
	callbackEvent := callbackEventFromDBModel(callbackEventRow)

	var responseRecord *domainrepo.ResponseRecord
	if decision.storeResponseRecord {
		responseRows, err := tx.Query(
			ctx,
			queryInsertResponseRecord,
			request.ID,
			callbackEvent.ID,
			string(decision.responseKind),
			nullableText(decision.selectedOptionID),
			nullableText(decision.freeText),
			nullableText(strings.TrimSpace(params.ResponderRef)),
			string(decision.persistedClassification),
			decision.effectiveResponse,
			now,
		)
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, fmt.Errorf("insert interaction response record: %w", err)
		}
		responseRow, err := pgx.CollectOneRow(responseRows, pgx.RowToStructByName[dbmodel.ResponseRecordRow])
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, fmt.Errorf("collect interaction response record: %w", err)
		}
		record := responseRecordFromDBModel(responseRow)
		responseRecord = &record
	}

	updatedRequest := request
	if decision.stateChanged {
		effectiveResponseID := nullableInt64(responseRecord)
		requestRow := tx.QueryRow(
			ctx,
			queryUpdateRequestState,
			request.ID,
			string(decision.nextState),
			string(decision.nextResolutionKind),
			effectiveResponseID,
		)
		item, err := scanRequestRow(requestRow)
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, fmt.Errorf("update interaction request state: %w", err)
		}
		updatedRequest = item
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("commit interaction callback tx: %w", err)
	}

	result := domainrepo.ApplyCallbackResult{
		Interaction:    updatedRequest,
		CallbackEvent:  callbackEvent,
		ResponseRecord: responseRecord,
		Classification: decision.resultClassification,
		ResumeRequired: decision.resumeRequired,
	}
	if responseRecord != nil {
		result.EffectiveResponseID = responseRecord.ID
	}
	return result, nil
}

func (r *Repository) getRequestForUpdate(ctx context.Context, tx pgx.Tx, interactionID string) (domainrepo.Request, bool, error) {
	rows, err := tx.Query(ctx, querySelectForUpdate, strings.TrimSpace(interactionID))
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("query interaction request for update: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RequestRow])
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("collect interaction request for update: %w", err)
	}
	if len(items) == 0 {
		return domainrepo.Request{}, false, nil
	}
	return requestFromDBModel(items[0]), true, nil
}

func (r *Repository) getCallbackEventByKey(ctx context.Context, tx pgx.Tx, interactionID string, adapterEventID string) (domainrepo.CallbackEvent, bool, error) {
	rows, err := tx.Query(ctx, queryGetCallbackEventByKey, strings.TrimSpace(interactionID), strings.TrimSpace(adapterEventID))
	if err != nil {
		return domainrepo.CallbackEvent{}, false, fmt.Errorf("query interaction callback event by key: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.CallbackEventRow])
	if err != nil {
		return domainrepo.CallbackEvent{}, false, fmt.Errorf("collect interaction callback event by key: %w", err)
	}
	if len(items) == 0 {
		return domainrepo.CallbackEvent{}, false, nil
	}
	return callbackEventFromDBModel(items[0]), true, nil
}

type callbackDecision struct {
	persistedClassification enumtypes.InteractionCallbackRecordClassification
	resultClassification    enumtypes.InteractionCallbackResultClassification
	nextState               enumtypes.InteractionState
	nextResolutionKind      enumtypes.InteractionResolutionKind
	responseKind            enumtypes.InteractionResponseKind
	selectedOptionID        string
	freeText                string
	storeResponseRecord     bool
	effectiveResponse       bool
	stateChanged            bool
	resumeRequired          bool
}

func classifyCallback(request domainrepo.Request, params domainrepo.ApplyCallbackParams, now time.Time) callbackDecision {
	decision := callbackDecision{
		persistedClassification: enumtypes.InteractionCallbackRecordClassificationApplied,
		resultClassification:    enumtypes.InteractionCallbackResultClassificationAccepted,
		nextState:               request.State,
		nextResolutionKind:      request.ResolutionKind,
	}

	switch params.CallbackKind {
	case enumtypes.InteractionCallbackKindDeliveryReceipt:
		switch strings.TrimSpace(params.DeliveryStatus) {
		case "accepted", "delivered", "failed":
			return decision
		default:
			decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationInvalid
			decision.resultClassification = enumtypes.InteractionCallbackResultClassificationInvalid
			return decision
		}
	case enumtypes.InteractionCallbackKindDecisionResponse:
		if request.InteractionKind != enumtypes.InteractionKindDecisionRequest {
			decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationInvalid
			decision.resultClassification = enumtypes.InteractionCallbackResultClassificationInvalid
			return decision
		}
	default:
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationInvalid
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationInvalid
		return decision
	}

	responseDecision, valid := classifyDecisionResponsePayload(request.RequestPayloadJSON, params)
	if valid {
		decision.responseKind = responseDecision.responseKind
		decision.selectedOptionID = responseDecision.selectedOptionID
		decision.freeText = responseDecision.freeText
		decision.storeResponseRecord = true
	}

	switch request.State {
	case enumtypes.InteractionStateResolved, enumtypes.InteractionStateCancelled, enumtypes.InteractionStateDeliveryExhausted:
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationStale
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationStale
		return decision
	case enumtypes.InteractionStateExpired:
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationExpired
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationExpired
		return decision
	}

	if request.ResponseDeadlineAt != nil && now.After(request.ResponseDeadlineAt.UTC()) {
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationExpired
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationExpired
		decision.nextState = enumtypes.InteractionStateExpired
		decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
		decision.stateChanged = request.State != enumtypes.InteractionStateExpired
		decision.resumeRequired = true
		return decision
	}

	if !valid {
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationInvalid
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationInvalid
		return decision
	}

	decision.nextState = enumtypes.InteractionStateResolved
	decision.stateChanged = request.State != enumtypes.InteractionStateResolved || request.ResolutionKind == enumtypes.InteractionResolutionKindNone
	decision.resumeRequired = true
	decision.effectiveResponse = true
	switch decision.responseKind {
	case enumtypes.InteractionResponseKindOption:
		decision.nextResolutionKind = enumtypes.InteractionResolutionKindOptionSelected
	case enumtypes.InteractionResponseKindFreeText:
		decision.nextResolutionKind = enumtypes.InteractionResolutionKindFreeTextSubmitted
	}
	return decision
}

type decisionResponseValidation struct {
	responseKind     enumtypes.InteractionResponseKind
	selectedOptionID string
	freeText         string
}

type decisionRequestPayload struct {
	AllowFreeText bool `json:"allow_free_text,omitempty"`
	Options       []struct {
		OptionID string `json:"option_id"`
	} `json:"options"`
}

func classifyDecisionResponsePayload(requestPayloadJSON []byte, params domainrepo.ApplyCallbackParams) (decisionResponseValidation, bool) {
	responseKind := enumtypes.InteractionResponseKind(strings.ToLower(strings.TrimSpace(string(params.ResponseKind))))
	switch responseKind {
	case enumtypes.InteractionResponseKindOption:
		optionID := strings.TrimSpace(params.SelectedOptionID)
		if optionID == "" {
			return decisionResponseValidation{}, false
		}
		var payload decisionRequestPayload
		if err := json.Unmarshal(requestPayloadJSON, &payload); err != nil {
			return decisionResponseValidation{}, false
		}
		for _, option := range payload.Options {
			if strings.TrimSpace(option.OptionID) == optionID {
				return decisionResponseValidation{responseKind: responseKind, selectedOptionID: optionID}, true
			}
		}
		return decisionResponseValidation{}, false
	case enumtypes.InteractionResponseKindFreeText:
		freeText := strings.TrimSpace(params.FreeText)
		if freeText == "" {
			return decisionResponseValidation{}, false
		}
		var payload decisionRequestPayload
		if err := json.Unmarshal(requestPayloadJSON, &payload); err != nil {
			return decisionResponseValidation{}, false
		}
		if !payload.AllowFreeText {
			return decisionResponseValidation{}, false
		}
		return decisionResponseValidation{responseKind: responseKind, freeText: freeText}, true
	default:
		return decisionResponseValidation{}, false
	}
}

func scanRequestRow(row pgx.Row) (domainrepo.Request, error) {
	var item dbmodel.RequestRow
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.RunID,
		&item.InteractionKind,
		&item.State,
		&item.ResolutionKind,
		&item.RecipientProvider,
		&item.RecipientRef,
		&item.RequestPayloadJSON,
		&item.ContextLinksJSON,
		&item.ResponseDeadlineAt,
		&item.EffectiveResponseID,
		&item.LastDeliveryAttemptNo,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domainrepo.Request{}, err
	}
	return requestFromDBModel(item), nil
}

func scanDeliveryAttemptRow(row pgx.Row) (domainrepo.DeliveryAttempt, error) {
	var item dbmodel.DeliveryAttemptRow
	err := row.Scan(
		&item.ID,
		&item.InteractionID,
		&item.AttemptNo,
		&item.DeliveryID,
		&item.AdapterKind,
		&item.Status,
		&item.RequestEnvelopeJSON,
		&item.AckPayloadJSON,
		&item.AdapterDeliveryID,
		&item.Retryable,
		&item.NextRetryAt,
		&item.LastErrorCode,
		&item.StartedAt,
		&item.FinishedAt,
	)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, err
	}
	return deliveryAttemptFromDBModel(item), nil
}

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableInt64(record *domainrepo.ResponseRecord) any {
	if record == nil || record.ID == 0 {
		return nil
	}
	return record.ID
}

func jsonOrEmptyObject(raw []byte) []byte {
	if len(raw) == 0 || !json.Valid(raw) {
		return []byte(`{}`)
	}
	return raw
}

func timestamptzPtrOrNil(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC()
}

func timeOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
