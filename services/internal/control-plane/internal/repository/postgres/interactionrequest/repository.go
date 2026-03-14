package interactionrequest

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/libs/go/mcp/userinteraction"
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
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
	//go:embed sql/ensure_channel_binding.sql
	queryEnsureChannelBinding string
	//go:embed sql/upsert_callback_handle.sql
	queryUpsertCallbackHandle string
	//go:embed sql/list_callback_handles_by_binding.sql
	queryListCallbackHandlesByBinding string
	//go:embed sql/get_callback_handle_by_hash_for_update.sql
	queryGetCallbackHandleByHashForUpdate string
	//go:embed sql/mark_callback_handle_used.sql
	queryMarkCallbackHandleUsed string
	//go:embed sql/get_channel_binding_by_id_for_update.sql
	queryGetChannelBindingByIDForUpdate string
	//go:embed sql/claim_next_dispatch_candidate.sql
	queryClaimNextDispatchCandidate string
	//go:embed sql/select_for_update.sql
	querySelectForUpdate string
	//go:embed sql/select_latest_attempt_for_update.sql
	querySelectLatestAttemptForUpdate string
	//go:embed sql/create_delivery_attempt.sql
	queryCreateDeliveryAttempt string
	//go:embed sql/touch_attempt_started_at.sql
	queryTouchAttemptStartedAt string
	//go:embed sql/get_attempt_by_delivery_id_for_update.sql
	queryGetAttemptByDeliveryIDForUpdate string
	//go:embed sql/update_attempt.sql
	queryUpdateAttempt string
	//go:embed sql/update_last_delivery_attempt_no.sql
	queryUpdateLastDeliveryAttemptNo string
	//go:embed sql/claim_next_expiry_candidate.sql
	queryClaimNextExpiryCandidate string
	//go:embed sql/get_callback_event_by_key.sql
	queryGetCallbackEventByKey string
	//go:embed sql/insert_callback_event.sql
	queryInsertCallbackEvent string
	//go:embed sql/insert_response_record.sql
	queryInsertResponseRecord string
	//go:embed sql/update_request_state.sql
	queryUpdateRequestState string
	//go:embed sql/update_dispatch_binding.sql
	queryUpdateDispatchBinding string
	//go:embed sql/update_request_projection.sql
	queryUpdateRequestProjection string
	//go:embed sql/update_channel_binding_projection.sql
	queryUpdateChannelBindingProjection string
)

const interactionResumeToolName = "user.decision.request"

// Repository persists interaction aggregate and callback evidence in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

type rowQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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
		string(params.ChannelFamily),
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

// EnsureChannelBinding returns or creates one active Telegram binding for the interaction.
func (r *Repository) EnsureChannelBinding(ctx context.Context, params domainrepo.EnsureChannelBindingParams) (domainrepo.ChannelBinding, error) {
	row := r.db.QueryRow(
		ctx,
		queryEnsureChannelBinding,
		strings.TrimSpace(params.InteractionID),
		strings.TrimSpace(params.AdapterKind),
		strings.TrimSpace(params.RecipientRef),
		nullableText(params.CallbackTokenKeyID),
		timestamptzPtrOrNil(params.CallbackTokenExpiresAt),
	)

	var item dbmodel.ChannelBindingRow
	if err := row.Scan(
		&item.ID,
		&item.InteractionID,
		&item.AdapterKind,
		&item.RecipientRef,
		&item.ProviderChatRef,
		&item.ProviderMessageRefJSON,
		&item.CallbackTokenKeyID,
		&item.CallbackTokenExpiresAt,
		&item.EditCapability,
		&item.ContinuationState,
		&item.LastOperatorSignalCode,
		&item.LastOperatorSignalAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return domainrepo.ChannelBinding{}, fmt.Errorf("ensure interaction channel binding: %w", err)
	}
	return channelBindingFromDBModel(item), nil
}

// UpsertCallbackHandles inserts missing callback handle hashes for the active binding.
func (r *Repository) UpsertCallbackHandles(ctx context.Context, params domainrepo.UpsertCallbackHandlesParams) ([]domainrepo.CallbackHandle, error) {
	for _, item := range params.Items {
		if _, err := r.db.Exec(
			ctx,
			queryUpsertCallbackHandle,
			strings.TrimSpace(params.InteractionID),
			params.ChannelBindingID,
			item.HandleHash,
			string(item.HandleKind),
			nullableText(item.OptionID),
			item.ResponseDeadlineAt.UTC(),
			item.GraceExpiresAt.UTC(),
		); err != nil {
			return nil, fmt.Errorf("upsert interaction callback handle: %w", err)
		}
	}

	rows, err := r.db.Query(ctx, queryListCallbackHandlesByBinding, params.ChannelBindingID)
	if err != nil {
		return nil, fmt.Errorf("query interaction callback handles by binding: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.CallbackHandleRow])
	if err != nil {
		return nil, fmt.Errorf("collect interaction callback handles by binding: %w", err)
	}

	out := make([]domainrepo.CallbackHandle, 0, len(items))
	for _, item := range items {
		out = append(out, callbackHandleFromDBModel(item))
	}
	return out, nil
}

// ClaimNextDispatch reserves or reclaims one due dispatch attempt for worker execution.
func (r *Repository) ClaimNextDispatch(ctx context.Context, params domainrepo.ClaimDispatchParams) (domainrepo.DispatchClaim, bool, error) {
	now := timeOrNow(params.Now)
	pendingAttemptTimeout := params.PendingAttemptTimeout
	if pendingAttemptTimeout <= 0 {
		pendingAttemptTimeout = time.Minute
	}
	staleBefore := now.Add(-pendingAttemptTimeout)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.DispatchClaim{}, false, fmt.Errorf("begin claim interaction dispatch tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, found, err := r.lookupRequestTx(ctx, tx, queryClaimNextDispatchCandidate, now, staleBefore)
	if err != nil {
		return domainrepo.DispatchClaim{}, false, err
	}
	if !found {
		return domainrepo.DispatchClaim{}, false, nil
	}

	var attempt domainrepo.DeliveryAttempt
	latestAttempt, latestFound, err := r.getLatestAttemptForUpdate(ctx, tx, request.ID)
	if err != nil {
		return domainrepo.DispatchClaim{}, false, err
	}
	if latestFound && latestAttempt.Status == enumtypes.InteractionDeliveryAttemptStatusPending {
		attempt, err = r.touchAttemptStartedAt(ctx, tx, latestAttempt.DeliveryID, now)
		if err != nil {
			return domainrepo.DispatchClaim{}, false, err
		}
	} else {
		attempt, err = r.createDeliveryAttemptTx(ctx, tx, request, domainrepo.CreateDeliveryAttemptParams{
			ChannelBindingID:       request.ActiveChannelBindingID,
			AdapterKind:            request.RecipientProvider,
			DeliveryRole:           enumtypes.InteractionDeliveryRolePrimaryDispatch,
			RequestEnvelopeJSON:    json.RawMessage(`{}`),
			AckPayloadJSON:         json.RawMessage(`{}`),
			ProviderMessageRefJSON: json.RawMessage(`{}`),
			Status:                 enumtypes.InteractionDeliveryAttemptStatusPending,
			StartedAt:              now,
		})
		if err != nil {
			return domainrepo.DispatchClaim{}, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.DispatchClaim{}, false, fmt.Errorf("commit claim interaction dispatch tx: %w", err)
	}

	return domainrepo.DispatchClaim{
		Interaction: request,
		Attempt:     attempt,
	}, true, nil
}

// CompleteDispatch persists one dispatch attempt outcome and applies terminal mutation when needed.
func (r *Repository) CompleteDispatch(ctx context.Context, params domainrepo.CompleteDispatchParams) (domainrepo.CompleteDispatchResult, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.CompleteDispatchResult{}, fmt.Errorf("begin complete interaction dispatch tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, found, err := r.getRequestForUpdate(ctx, tx, params.InteractionID)
	if err != nil {
		return domainrepo.CompleteDispatchResult{}, err
	}
	if !found {
		return domainrepo.CompleteDispatchResult{}, fmt.Errorf("interaction request not found")
	}

	attempt, found, err := r.getAttemptByDeliveryIDForUpdate(ctx, tx, params.DeliveryID)
	if err != nil {
		return domainrepo.CompleteDispatchResult{}, err
	}
	if !found {
		return domainrepo.CompleteDispatchResult{}, fmt.Errorf("interaction delivery attempt not found")
	}
	if attempt.InteractionID != request.ID {
		return domainrepo.CompleteDispatchResult{}, fmt.Errorf("interaction delivery attempt does not belong to request")
	}

	if attempt.Status != enumtypes.InteractionDeliveryAttemptStatusPending {
		if err := tx.Commit(ctx); err != nil {
			return domainrepo.CompleteDispatchResult{}, fmt.Errorf("commit duplicate interaction dispatch completion tx: %w", err)
		}
		return domainrepo.CompleteDispatchResult{
			Interaction: request,
			Attempt:     attempt,
		}, nil
	}

	updatedAttempt, err := r.updateAttempt(ctx, tx, params)
	if err != nil {
		return domainrepo.CompleteDispatchResult{}, err
	}
	bindingIDForDispatch := updatedAttempt.ChannelBindingID
	if bindingIDForDispatch == 0 {
		bindingIDForDispatch = request.ActiveChannelBindingID
	}
	if updatedAttempt.Status == enumtypes.InteractionDeliveryAttemptStatusAccepted && bindingIDForDispatch > 0 {
		if _, err := r.updateDispatchBindingTx(ctx, tx, bindingIDForDispatch, params.ProviderMessageRefJSON, params.EditCapability, params.CallbackTokenExpiresAt); err != nil {
			return domainrepo.CompleteDispatchResult{}, err
		}
	}

	decision := classifyDispatchCompletion(request, updatedAttempt.Status)
	updatedRequest := request
	if decision.stateChanged {
		updatedRequest, err = r.updateRequestState(ctx, tx, request.ID, decision.nextState, decision.nextResolutionKind, nil)
		if err != nil {
			return domainrepo.CompleteDispatchResult{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.CompleteDispatchResult{}, fmt.Errorf("commit complete interaction dispatch tx: %w", err)
	}

	return domainrepo.CompleteDispatchResult{
		Interaction:    updatedRequest,
		Attempt:        updatedAttempt,
		StateChanged:   decision.stateChanged,
		ResumeRequired: decision.resumeRequired,
	}, nil
}

// UpdateDispatchBinding stores adapter ack data after accepted primary delivery.
func (r *Repository) UpdateDispatchBinding(ctx context.Context, params domainrepo.UpdateDispatchBindingParams) (domainrepo.ChannelBinding, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.ChannelBinding{}, fmt.Errorf("begin update interaction dispatch binding tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	attempt, found, err := r.getAttemptByDeliveryIDForUpdate(ctx, tx, params.DeliveryID)
	if err != nil {
		return domainrepo.ChannelBinding{}, err
	}
	if !found {
		return domainrepo.ChannelBinding{}, errs.NotFound{Msg: "delivery_id: not found"}
	}
	if strings.TrimSpace(attempt.InteractionID) != strings.TrimSpace(params.InteractionID) {
		return domainrepo.ChannelBinding{}, fmt.Errorf("interaction delivery attempt does not belong to request")
	}
	bindingID := attempt.ChannelBindingID
	if bindingID == 0 {
		request, found, err := r.getRequestForUpdate(ctx, tx, params.InteractionID)
		if err != nil {
			return domainrepo.ChannelBinding{}, err
		}
		if !found {
			return domainrepo.ChannelBinding{}, errs.NotFound{Msg: "interaction_id: not found"}
		}
		bindingID = request.ActiveChannelBindingID
	}
	if bindingID == 0 {
		return domainrepo.ChannelBinding{}, fmt.Errorf("interaction delivery attempt has no channel binding")
	}

	binding, err := r.updateDispatchBindingTx(ctx, tx, bindingID, params.ProviderMessageRefJSON, params.EditCapability, params.CallbackTokenExpiresAt)
	if err != nil {
		return domainrepo.ChannelBinding{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domainrepo.ChannelBinding{}, fmt.Errorf("commit update interaction dispatch binding tx: %w", err)
	}
	return binding, nil
}

// ExpireNextDue marks one deadline-expired decision interaction terminal.
func (r *Repository) ExpireNextDue(ctx context.Context, params domainrepo.ExpireDueParams) (domainrepo.ExpireDueResult, bool, error) {
	now := timeOrNow(params.Now)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domainrepo.ExpireDueResult{}, false, fmt.Errorf("begin expire interaction tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, found, err := r.lookupRequestTx(ctx, tx, queryClaimNextExpiryCandidate, now)
	if err != nil {
		return domainrepo.ExpireDueResult{}, false, err
	}
	if !found {
		return domainrepo.ExpireDueResult{}, false, nil
	}

	decision := classifyExpiry(request)
	if !decision.stateChanged {
		if err := tx.Commit(ctx); err != nil {
			return domainrepo.ExpireDueResult{}, false, fmt.Errorf("commit no-op expire interaction tx: %w", err)
		}
		return domainrepo.ExpireDueResult{
			Interaction:    request,
			StateChanged:   false,
			ResumeRequired: decision.resumeRequired,
		}, true, nil
	}

	var exhaustedAttempt *domainrepo.DeliveryAttempt
	if request.State == enumtypes.InteractionStatePendingDispatch {
		latestAttempt, latestFound, err := r.getLatestAttemptForUpdate(ctx, tx, request.ID)
		if err != nil {
			return domainrepo.ExpireDueResult{}, false, err
		}
		if latestFound && (latestAttempt.Status == enumtypes.InteractionDeliveryAttemptStatusPending || latestAttempt.Status == enumtypes.InteractionDeliveryAttemptStatusFailed) {
			updatedAttempt, err := r.updateAttempt(ctx, tx, domainrepo.CompleteDispatchParams{
				InteractionID:       request.ID,
				DeliveryID:          latestAttempt.DeliveryID,
				AdapterKind:         latestAttempt.AdapterKind,
				Status:              enumtypes.InteractionDeliveryAttemptStatusExhausted,
				RequestEnvelopeJSON: latestAttempt.RequestEnvelopeJSON,
				AckPayloadJSON:      latestAttempt.AckPayloadJSON,
				AdapterDeliveryID:   latestAttempt.AdapterDeliveryID,
				LastErrorCode:       "deadline_exceeded",
				FinishedAt:          now,
			})
			if err != nil {
				return domainrepo.ExpireDueResult{}, false, err
			}
			exhaustedAttempt = &updatedAttempt
		}
	}

	updatedRequest, err := r.updateRequestState(ctx, tx, request.ID, decision.nextState, decision.nextResolutionKind, nil)
	if err != nil {
		return domainrepo.ExpireDueResult{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.ExpireDueResult{}, false, fmt.Errorf("commit expire interaction tx: %w", err)
	}

	return domainrepo.ExpireDueResult{
		Interaction:    updatedRequest,
		Attempt:        exhaustedAttempt,
		StateChanged:   true,
		ResumeRequired: decision.resumeRequired,
	}, true, nil
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

func (r *Repository) lookupRequestTx(ctx context.Context, tx pgx.Tx, query string, args ...any) (domainrepo.Request, bool, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("query interaction request tx: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RequestRow])
	if err != nil {
		return domainrepo.Request{}, false, fmt.Errorf("collect interaction request tx: %w", err)
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

	attempt, err := r.createDeliveryAttemptTx(ctx, tx, request, params)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, err
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
		return domainrepo.ApplyCallbackResult{}, errs.NotFound{Msg: "interaction_id: not found"}
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
			Accepted:       false,
			Classification: enumtypes.InteractionCallbackResultClassificationDuplicate,
		}, nil
	}

	var binding *domainrepo.ChannelBinding
	if request.ActiveChannelBindingID > 0 {
		item, found, err := r.getChannelBindingByIDForUpdate(ctx, tx, request.ActiveChannelBindingID)
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, err
		}
		if found {
			binding = &item
		}
	}

	var handle *domainrepo.CallbackHandle
	var handleHash []byte
	if callbackHandle := strings.TrimSpace(params.CallbackHandle); callbackHandle != "" {
		sum := sha256.Sum256([]byte(callbackHandle))
		handleHash = sum[:]
		item, found, err := r.getCallbackHandleByHashForUpdate(ctx, tx, handleHash)
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, err
		}
		if found {
			handle = &item
		}
	}

	decision := classifyCallback(request, binding, handle, handleHash, params, now)
	callbackEventRows, err := tx.Query(
		ctx,
		queryInsertCallbackEvent,
		request.ID,
		nullableInt64Value(decision.bindingID),
		nullableUUID(params.DeliveryID),
		strings.TrimSpace(params.AdapterEventID),
		string(params.CallbackKind),
		string(decision.persistedClassification),
		nullableBytes(decision.handleHash),
		jsonOrEmptyObject(params.NormalizedPayloadJSON),
		jsonOrEmptyObject(params.RawPayloadJSON),
		jsonOrEmptyObject(params.ProviderMessageRefJSON),
		nullableText(params.ProviderUpdateID),
		nullableText(params.ProviderCallbackQueryID),
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
	if decision.markHandleUsed && handle != nil {
		if _, err := r.markCallbackHandleUsed(ctx, tx, handle.ID, callbackEvent.ID, now); err != nil {
			return domainrepo.ApplyCallbackResult{}, err
		}
	}

	var responseRecord *domainrepo.ResponseRecord
	if decision.storeResponseRecord {
		responseRows, err := tx.Query(
			ctx,
			queryInsertResponseRecord,
			request.ID,
			nullableInt64Value(decision.bindingID),
			callbackEvent.ID,
			string(decision.handleKind),
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
	if decision.updateRequestProjection {
		updatedRequest, err = r.updateRequestProjectionTx(ctx, tx, request.ID, decision.operatorState, decision.operatorSignalCode, timePtr(now))
		if err != nil {
			return domainrepo.ApplyCallbackResult{}, err
		}
	}

	var updatedBinding *domainrepo.ChannelBinding
	if binding != nil {
		updatedBinding = binding
		if decision.updateBindingProjection {
			item, err := r.updateChannelBindingProjectionTx(ctx, tx, binding.ID, decision.continuationState, decision.operatorSignalCode, timePtr(now))
			if err != nil {
				return domainrepo.ApplyCallbackResult{}, err
			}
			updatedBinding = &item
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domainrepo.ApplyCallbackResult{}, fmt.Errorf("commit interaction callback tx: %w", err)
	}

	result := domainrepo.ApplyCallbackResult{
		Interaction:        updatedRequest,
		Binding:            updatedBinding,
		CallbackEvent:      callbackEvent,
		ResponseRecord:     responseRecord,
		Accepted:           decision.accepted,
		Classification:     decision.resultClassification,
		ContinuationAction: decision.continuationAction,
		OperatorSignalCode: decision.operatorSignalCode,
		ResumeRequired:     decision.resumeRequired,
	}
	if responseRecord != nil {
		result.EffectiveResponseID = responseRecord.ID
	}
	return result, nil
}

func (r *Repository) getRequestForUpdate(ctx context.Context, tx pgx.Tx, interactionID string) (domainrepo.Request, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		querySelectForUpdate,
		"query interaction request for update",
		"collect interaction request for update",
		requestFromDBModel,
		strings.TrimSpace(interactionID),
	)
}

func (r *Repository) getLatestAttemptForUpdate(ctx context.Context, tx pgx.Tx, interactionID string) (domainrepo.DeliveryAttempt, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		querySelectLatestAttemptForUpdate,
		"query latest interaction delivery attempt for update",
		"collect latest interaction delivery attempt for update",
		deliveryAttemptFromDBModel,
		strings.TrimSpace(interactionID),
	)
}

func (r *Repository) getAttemptByDeliveryIDForUpdate(ctx context.Context, tx pgx.Tx, deliveryID string) (domainrepo.DeliveryAttempt, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		queryGetAttemptByDeliveryIDForUpdate,
		"query interaction delivery attempt by delivery id",
		"collect interaction delivery attempt by delivery id",
		deliveryAttemptFromDBModel,
		strings.TrimSpace(deliveryID),
	)
}

func (r *Repository) getCallbackEventByKey(ctx context.Context, tx pgx.Tx, interactionID string, adapterEventID string) (domainrepo.CallbackEvent, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		queryGetCallbackEventByKey,
		"query interaction callback event by key",
		"collect interaction callback event by key",
		callbackEventFromDBModel,
		strings.TrimSpace(interactionID),
		strings.TrimSpace(adapterEventID),
	)
}

func (r *Repository) getChannelBindingByIDForUpdate(ctx context.Context, tx pgx.Tx, bindingID int64) (domainrepo.ChannelBinding, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		queryGetChannelBindingByIDForUpdate,
		"query interaction channel binding by id",
		"collect interaction channel binding by id",
		channelBindingFromDBModel,
		bindingID,
	)
}

func (r *Repository) getCallbackHandleByHashForUpdate(ctx context.Context, tx pgx.Tx, handleHash []byte) (domainrepo.CallbackHandle, bool, error) {
	return queryOptionalModelByName(
		ctx,
		tx,
		queryGetCallbackHandleByHashForUpdate,
		"query interaction callback handle by hash",
		"collect interaction callback handle by hash",
		callbackHandleFromDBModel,
		handleHash,
	)
}

func (r *Repository) createDeliveryAttemptTx(ctx context.Context, tx pgx.Tx, request domainrepo.Request, params domainrepo.CreateDeliveryAttemptParams) (domainrepo.DeliveryAttempt, error) {
	nextAttemptNo := request.LastDeliveryAttemptNo + 1
	row := tx.QueryRow(
		ctx,
		queryCreateDeliveryAttempt,
		request.ID,
		nullableInt64Value(params.ChannelBindingID),
		nextAttemptNo,
		strings.TrimSpace(params.AdapterKind),
		string(params.DeliveryRole),
		string(params.Status),
		jsonOrEmptyObject(params.RequestEnvelopeJSON),
		jsonOrEmptyObject(params.AckPayloadJSON),
		nullableText(params.AdapterDeliveryID),
		jsonOrEmptyObject(params.ProviderMessageRefJSON),
		params.Retryable,
		timestamptzPtrOrNil(params.NextRetryAt),
		nullableText(params.LastErrorCode),
		nullableText(params.ContinuationReason),
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
	return attempt, nil
}

func (r *Repository) touchAttemptStartedAt(ctx context.Context, tx pgx.Tx, deliveryID string, startedAt time.Time) (domainrepo.DeliveryAttempt, error) {
	row := tx.QueryRow(ctx, queryTouchAttemptStartedAt, strings.TrimSpace(deliveryID), timeOrNow(startedAt))
	attempt, err := scanDeliveryAttemptRow(row)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("touch interaction delivery attempt started_at: %w", err)
	}
	return attempt, nil
}

func (r *Repository) updateAttempt(ctx context.Context, tx pgx.Tx, params domainrepo.CompleteDispatchParams) (domainrepo.DeliveryAttempt, error) {
	row := tx.QueryRow(
		ctx,
		queryUpdateAttempt,
		strings.TrimSpace(params.DeliveryID),
		strings.TrimSpace(params.AdapterKind),
		string(params.Status),
		jsonOrEmptyObject(params.RequestEnvelopeJSON),
		jsonOrEmptyObject(params.AckPayloadJSON),
		nullableText(params.AdapterDeliveryID),
		jsonOrEmptyObject(params.ProviderMessageRefJSON),
		params.Retryable,
		timestamptzPtrOrNil(params.NextRetryAt),
		nullableText(params.LastErrorCode),
		timestamptzPtrOrNil(timePtr(timeOrNow(params.FinishedAt))),
	)
	attempt, err := scanDeliveryAttemptRow(row)
	if err != nil {
		return domainrepo.DeliveryAttempt{}, fmt.Errorf("update interaction delivery attempt: %w", err)
	}
	return attempt, nil
}

func (r *Repository) updateRequestState(ctx context.Context, tx pgx.Tx, interactionID string, state enumtypes.InteractionState, resolutionKind enumtypes.InteractionResolutionKind, effectiveResponseID any) (domainrepo.Request, error) {
	requestRow := tx.QueryRow(
		ctx,
		queryUpdateRequestState,
		interactionID,
		string(state),
		string(resolutionKind),
		effectiveResponseID,
	)
	item, err := scanRequestRow(requestRow)
	if err != nil {
		return domainrepo.Request{}, fmt.Errorf("update interaction request state: %w", err)
	}
	return item, nil
}

func (r *Repository) updateRequestProjectionTx(ctx context.Context, tx pgx.Tx, interactionID string, operatorState enumtypes.InteractionOperatorState, signalCode enumtypes.InteractionOperatorSignalCode, signalAt *time.Time) (domainrepo.Request, error) {
	row := tx.QueryRow(
		ctx,
		queryUpdateRequestProjection,
		strings.TrimSpace(interactionID),
		string(operatorState),
		nullableText(string(signalCode)),
		timestamptzPtrOrNil(signalAt),
	)
	item, err := scanRequestRow(row)
	if err != nil {
		return domainrepo.Request{}, fmt.Errorf("update interaction request projection: %w", err)
	}
	return item, nil
}

func (r *Repository) updateChannelBindingProjectionTx(ctx context.Context, tx pgx.Tx, bindingID int64, continuationState enumtypes.InteractionContinuationState, signalCode enumtypes.InteractionOperatorSignalCode, signalAt *time.Time) (domainrepo.ChannelBinding, error) {
	row := tx.QueryRow(
		ctx,
		queryUpdateChannelBindingProjection,
		bindingID,
		string(continuationState),
		nullableText(string(signalCode)),
		timestamptzPtrOrNil(signalAt),
	)
	var item dbmodel.ChannelBindingRow
	if err := row.Scan(
		&item.ID,
		&item.InteractionID,
		&item.AdapterKind,
		&item.RecipientRef,
		&item.ProviderChatRef,
		&item.ProviderMessageRefJSON,
		&item.CallbackTokenKeyID,
		&item.CallbackTokenExpiresAt,
		&item.EditCapability,
		&item.ContinuationState,
		&item.LastOperatorSignalCode,
		&item.LastOperatorSignalAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return domainrepo.ChannelBinding{}, fmt.Errorf("update interaction channel binding projection: %w", err)
	}
	return channelBindingFromDBModel(item), nil
}

func (r *Repository) updateDispatchBindingTx(ctx context.Context, tx pgx.Tx, bindingID int64, providerMessageRefJSON json.RawMessage, editCapability enumtypes.InteractionEditCapability, tokenExpiresAt *time.Time) (domainrepo.ChannelBinding, error) {
	row := tx.QueryRow(
		ctx,
		queryUpdateDispatchBinding,
		bindingID,
		jsonOrEmptyObject(providerMessageRefJSON),
		string(editCapability),
		timestamptzPtrOrNil(tokenExpiresAt),
	)
	var item dbmodel.ChannelBindingRow
	if err := row.Scan(
		&item.ID,
		&item.InteractionID,
		&item.AdapterKind,
		&item.RecipientRef,
		&item.ProviderChatRef,
		&item.ProviderMessageRefJSON,
		&item.CallbackTokenKeyID,
		&item.CallbackTokenExpiresAt,
		&item.EditCapability,
		&item.ContinuationState,
		&item.LastOperatorSignalCode,
		&item.LastOperatorSignalAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return domainrepo.ChannelBinding{}, fmt.Errorf("update interaction dispatch binding: %w", err)
	}
	return channelBindingFromDBModel(item), nil
}

func (r *Repository) markCallbackHandleUsed(ctx context.Context, tx pgx.Tx, handleID int64, callbackEventID int64, usedAt time.Time) (domainrepo.CallbackHandle, error) {
	row := tx.QueryRow(ctx, queryMarkCallbackHandleUsed, handleID, callbackEventID, usedAt.UTC())
	var item dbmodel.CallbackHandleRow
	if err := row.Scan(
		&item.ID,
		&item.InteractionID,
		&item.ChannelBindingID,
		&item.HandleHash,
		&item.HandleKind,
		&item.OptionID,
		&item.State,
		&item.ResponseDeadlineAt,
		&item.GraceExpiresAt,
		&item.UsedCallbackEventID,
		&item.UsedAt,
		&item.CreatedAt,
	); err != nil {
		return domainrepo.CallbackHandle{}, fmt.Errorf("mark interaction callback handle used: %w", err)
	}
	return callbackHandleFromDBModel(item), nil
}

type callbackDecision struct {
	persistedClassification enumtypes.InteractionCallbackRecordClassification
	resultClassification    enumtypes.InteractionCallbackResultClassification
	nextState               enumtypes.InteractionState
	nextResolutionKind      enumtypes.InteractionResolutionKind
	accepted                bool
	bindingID               int64
	handleHash              []byte
	handleKind              enumtypes.InteractionCallbackHandleKind
	responseKind            enumtypes.InteractionResponseKind
	selectedOptionID        string
	freeText                string
	storeResponseRecord     bool
	effectiveResponse       bool
	markHandleUsed          bool
	stateChanged            bool
	updateRequestProjection bool
	updateBindingProjection bool
	operatorState           enumtypes.InteractionOperatorState
	operatorSignalCode      enumtypes.InteractionOperatorSignalCode
	continuationState       enumtypes.InteractionContinuationState
	continuationAction      enumtypes.InteractionContinuationAction
	resumeRequired          bool
}

type dispatchCompletionDecision struct {
	nextState          enumtypes.InteractionState
	nextResolutionKind enumtypes.InteractionResolutionKind
	stateChanged       bool
	resumeRequired     bool
}

func classifyDispatchCompletion(request domainrepo.Request, attemptStatus enumtypes.InteractionDeliveryAttemptStatus) dispatchCompletionDecision {
	decision := dispatchCompletionDecision{
		nextState:          request.State,
		nextResolutionKind: request.ResolutionKind,
	}

	if request.State != enumtypes.InteractionStatePendingDispatch {
		return decision
	}

	switch attemptStatus {
	case enumtypes.InteractionDeliveryAttemptStatusAccepted:
		if request.InteractionKind == enumtypes.InteractionKindNotify {
			decision.nextState = enumtypes.InteractionStateResolved
			decision.nextResolutionKind = enumtypes.InteractionResolutionKindDeliveryOnly
		} else {
			decision.nextState = enumtypes.InteractionStateOpen
			decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
		}
		decision.stateChanged = decision.nextState != request.State || decision.nextResolutionKind != request.ResolutionKind
	case enumtypes.InteractionDeliveryAttemptStatusExhausted:
		decision.nextState = enumtypes.InteractionStateDeliveryExhausted
		decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
		decision.stateChanged = decision.nextState != request.State || decision.nextResolutionKind != request.ResolutionKind
		decision.resumeRequired = request.InteractionKind == enumtypes.InteractionKindDecisionRequest
	}

	return decision
}

type expiryDecision struct {
	nextState          enumtypes.InteractionState
	nextResolutionKind enumtypes.InteractionResolutionKind
	stateChanged       bool
	resumeRequired     bool
}

func classifyExpiry(request domainrepo.Request) expiryDecision {
	decision := expiryDecision{
		nextState:          request.State,
		nextResolutionKind: request.ResolutionKind,
		resumeRequired:     request.InteractionKind == enumtypes.InteractionKindDecisionRequest,
	}

	switch request.State {
	case enumtypes.InteractionStatePendingDispatch:
		decision.nextState = enumtypes.InteractionStateDeliveryExhausted
	case enumtypes.InteractionStateOpen:
		decision.nextState = enumtypes.InteractionStateExpired
	default:
		return decision
	}

	decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
	decision.stateChanged = decision.nextState != request.State || decision.nextResolutionKind != request.ResolutionKind
	return decision
}

func classifyCallback(request domainrepo.Request, binding *domainrepo.ChannelBinding, handle *domainrepo.CallbackHandle, handleHash []byte, params domainrepo.ApplyCallbackParams, now time.Time) callbackDecision {
	decision := callbackDecision{
		persistedClassification: enumtypes.InteractionCallbackRecordClassificationApplied,
		resultClassification:    enumtypes.InteractionCallbackResultClassificationAccepted,
		accepted:                true,
		bindingID:               bindingID(binding),
		handleHash:              handleHash,
		nextState:               request.State,
		nextResolutionKind:      request.ResolutionKind,
		operatorState:           request.OperatorState,
		continuationAction:      enumtypes.InteractionContinuationActionNone,
	}

	switch params.CallbackKind {
	case enumtypes.InteractionCallbackKindDeliveryReceipt:
		switch strings.TrimSpace(params.DeliveryStatus) {
		case "accepted", "delivered", "failed":
			return decision
		default:
			return invalidCallbackDecision(request, decision, false)
		}
	case enumtypes.InteractionCallbackKindTransportFailure:
		if binding == nil {
			return invalidCallbackDecision(request, decision, false)
		}
		if params.TransportRetryable {
			return decision
		}
		decision.updateRequestProjection = true
		decision.updateBindingProjection = true
		decision.operatorState = enumtypes.InteractionOperatorStateManualFallbackRequired
		decision.operatorSignalCode = enumtypes.InteractionOperatorSignalCodeDeliveryRetryExhausted
		decision.continuationState = enumtypes.InteractionContinuationStateManualFallbackRequired
		decision.continuationAction = enumtypes.InteractionContinuationActionManualFallback
		return decision
	case enumtypes.InteractionCallbackKindOptionSelected, enumtypes.InteractionCallbackKindFreeTextReceived:
		if request.InteractionKind != enumtypes.InteractionKindDecisionRequest {
			return invalidCallbackDecision(request, decision, false)
		}
	default:
		return invalidCallbackDecision(request, decision, false)
	}

	switch request.State {
	case enumtypes.InteractionStateResolved, enumtypes.InteractionStateCancelled, enumtypes.InteractionStateDeliveryExhausted:
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationStale
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationStale
		decision.accepted = false
		return decision
	case enumtypes.InteractionStateExpired:
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationExpired
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationExpired
		decision.accepted = false
		return decision
	}

	if binding == nil || handle == nil || handle.ChannelBindingID == 0 || handle.ChannelBindingID != binding.ID || strings.TrimSpace(handle.InteractionID) != strings.TrimSpace(request.ID) {
		return invalidCallbackDecision(request, decision, true)
	}
	decision.bindingID = handle.ChannelBindingID
	decision.handleKind = handle.HandleKind

	if handle.GraceExpiresAt.Before(now) {
		return invalidCallbackDecision(request, decision, true)
	}
	if request.ResponseDeadlineAt != nil && now.After(request.ResponseDeadlineAt.UTC()) {
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationExpired
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationExpired
		decision.accepted = false
		decision.nextState = enumtypes.InteractionStateExpired
		decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
		decision.stateChanged = request.State != enumtypes.InteractionStateExpired
		decision.updateRequestProjection = true
		decision.updateBindingProjection = true
		decision.operatorState = enumtypes.InteractionOperatorStateWatch
		decision.operatorSignalCode = enumtypes.InteractionOperatorSignalCodeExpiredWait
		decision.continuationState = enumtypes.InteractionContinuationStateClosed
		decision.resumeRequired = true
		return decision
	}
	if now.After(handle.ResponseDeadlineAt.UTC()) || handle.State == enumtypes.InteractionCallbackHandleStateExpired {
		return expiredCallbackDecision(request, decision)
	}
	if handle.State == enumtypes.InteractionCallbackHandleStateUsed || handle.UsedCallbackEventID != 0 {
		decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationDuplicate
		decision.resultClassification = enumtypes.InteractionCallbackResultClassificationDuplicate
		decision.accepted = false
		return decision
	}

	responseDecision, valid := classifyDecisionResponsePayload(request, handle, params)
	if !valid {
		return invalidCallbackDecision(request, decision, true)
	}
	decision.responseKind = responseDecision.responseKind
	decision.selectedOptionID = responseDecision.selectedOptionID
	decision.freeText = responseDecision.freeText
	decision.storeResponseRecord = true
	decision.effectiveResponse = true
	decision.markHandleUsed = true
	decision.updateRequestProjection = true
	decision.updateBindingProjection = true
	decision.operatorState = enumtypes.InteractionOperatorStateResolved
	decision.continuationAction, decision.continuationState = continuationForBinding(binding)

	decision.nextState = enumtypes.InteractionStateResolved
	decision.stateChanged = request.State != enumtypes.InteractionStateResolved || request.ResolutionKind == enumtypes.InteractionResolutionKindNone
	decision.resumeRequired = true
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

func fitsInteractionResumePayloadLimit(
	requestID string,
	responseKind enumtypes.InteractionResponseKind,
	selectedOptionID string,
	freeText string,
	occurredAt time.Time,
) bool {
	candidate := valuetypes.InteractionResumePayload{
		InteractionID:    requestID,
		ToolName:         interactionResumeToolName,
		RequestStatus:    enumtypes.InteractionRequestStatusAnswered,
		ResponseKind:     responseKind,
		ResolvedAt:       occurredAt.UTC().Format(time.RFC3339Nano),
		ResolutionReason: string(enumtypes.InteractionCallbackResultClassificationAccepted),
	}
	switch responseKind {
	case enumtypes.InteractionResponseKindOption:
		candidate.SelectedOptionID = selectedOptionID
	case enumtypes.InteractionResponseKindFreeText:
		candidate.FreeText = freeText
	}

	encodedCandidate, err := json.Marshal(candidate)
	if err != nil {
		return false
	}
	return len(encodedCandidate) <= userinteraction.ResumePayloadMaxBytes
}

func classifyDecisionResponsePayload(request domainrepo.Request, handle *domainrepo.CallbackHandle, params domainrepo.ApplyCallbackParams) (decisionResponseValidation, bool) {
	switch params.CallbackKind {
	case enumtypes.InteractionCallbackKindOptionSelected:
		if handle == nil || handle.HandleKind != enumtypes.InteractionCallbackHandleKindOption {
			return decisionResponseValidation{}, false
		}
		optionID := strings.TrimSpace(handle.OptionID)
		if optionID == "" {
			return decisionResponseValidation{}, false
		}
		if !fitsInteractionResumePayloadLimit(request.ID, enumtypes.InteractionResponseKindOption, optionID, "", params.OccurredAt) {
			return decisionResponseValidation{}, false
		}
		return decisionResponseValidation{responseKind: enumtypes.InteractionResponseKindOption, selectedOptionID: optionID}, true
	case enumtypes.InteractionCallbackKindFreeTextReceived:
		if handle == nil || handle.HandleKind != enumtypes.InteractionCallbackHandleKindFreeTextSession {
			return decisionResponseValidation{}, false
		}
		freeText := strings.TrimSpace(params.FreeText)
		if freeText == "" {
			return decisionResponseValidation{}, false
		}
		if len([]byte(freeText)) > userinteraction.DecisionResponseFreeTextMaxBytes {
			return decisionResponseValidation{}, false
		}
		var payload decisionRequestPayload
		if err := json.Unmarshal(request.RequestPayloadJSON, &payload); err != nil {
			return decisionResponseValidation{}, false
		}
		if !payload.AllowFreeText {
			return decisionResponseValidation{}, false
		}
		if !fitsInteractionResumePayloadLimit(request.ID, enumtypes.InteractionResponseKindFreeText, "", freeText, params.OccurredAt) {
			return decisionResponseValidation{}, false
		}
		return decisionResponseValidation{responseKind: enumtypes.InteractionResponseKindFreeText, freeText: freeText}, true
	default:
		return decisionResponseValidation{}, false
	}
}

func invalidCallbackDecision(request domainrepo.Request, decision callbackDecision, signalInvalidPayload bool) callbackDecision {
	decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationInvalid
	decision.resultClassification = enumtypes.InteractionCallbackResultClassificationInvalid
	decision.accepted = false
	if signalInvalidPayload {
		decision.updateRequestProjection = true
		decision.operatorState = enumtypes.InteractionOperatorStateWatch
		decision.operatorSignalCode = enumtypes.InteractionOperatorSignalCodeInvalidCallbackPayload
		if decision.bindingID != 0 {
			decision.updateBindingProjection = true
			decision.continuationState = enumtypes.InteractionContinuationStateManualFallbackRequired
		}
	}
	_ = request
	return decision
}

func expiredCallbackDecision(request domainrepo.Request, decision callbackDecision) callbackDecision {
	decision.persistedClassification = enumtypes.InteractionCallbackRecordClassificationExpired
	decision.resultClassification = enumtypes.InteractionCallbackResultClassificationExpired
	decision.accepted = false
	decision.nextState = enumtypes.InteractionStateExpired
	decision.nextResolutionKind = enumtypes.InteractionResolutionKindNone
	decision.stateChanged = request.State != enumtypes.InteractionStateExpired
	decision.resumeRequired = true
	decision.updateRequestProjection = true
	decision.updateBindingProjection = decision.bindingID != 0
	decision.operatorState = enumtypes.InteractionOperatorStateWatch
	decision.operatorSignalCode = enumtypes.InteractionOperatorSignalCodeExpiredWait
	decision.continuationState = enumtypes.InteractionContinuationStateClosed
	return decision
}

func continuationForBinding(binding *domainrepo.ChannelBinding) (enumtypes.InteractionContinuationAction, enumtypes.InteractionContinuationState) {
	if binding == nil {
		return enumtypes.InteractionContinuationActionNone, enumtypes.InteractionContinuationStateClosed
	}
	switch binding.EditCapability {
	case enumtypes.InteractionEditCapabilityEditable, enumtypes.InteractionEditCapabilityKeyboardOnly:
		return enumtypes.InteractionContinuationActionEditMessage, enumtypes.InteractionContinuationStateReadyForEdit
	default:
		return enumtypes.InteractionContinuationActionSendFollowUp, enumtypes.InteractionContinuationStateFollowUpRequired
	}
}

func bindingID(binding *domainrepo.ChannelBinding) int64 {
	if binding == nil {
		return 0
	}
	return binding.ID
}

func queryOptionalStructByName[T any](ctx context.Context, querier rowQuerier, query string, queryMessage string, collectMessage string, args ...any) (T, bool, error) {
	var zero T

	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return zero, false, fmt.Errorf("%s: %w", queryMessage, err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return zero, false, fmt.Errorf("%s: %w", collectMessage, err)
	}
	if len(items) == 0 {
		return zero, false, nil
	}
	return items[0], true, nil
}

func queryOptionalModelByName[Row any, Model any](ctx context.Context, querier rowQuerier, query string, queryMessage string, collectMessage string, caster func(Row) Model, args ...any) (Model, bool, error) {
	var zero Model

	item, found, err := queryOptionalStructByName[Row](ctx, querier, query, queryMessage, collectMessage, args...)
	if err != nil {
		return zero, false, err
	}
	if !found {
		return zero, false, nil
	}
	return caster(item), true, nil
}

func scanRequestRow(row pgx.Row) (domainrepo.Request, error) {
	var item dbmodel.RequestRow
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.RunID,
		&item.InteractionKind,
		&item.ChannelFamily,
		&item.State,
		&item.ResolutionKind,
		&item.RecipientProvider,
		&item.RecipientRef,
		&item.RequestPayloadJSON,
		&item.ContextLinksJSON,
		&item.ResponseDeadlineAt,
		&item.EffectiveResponseID,
		&item.ActiveChannelBindingID,
		&item.OperatorState,
		&item.OperatorSignalCode,
		&item.OperatorSignalAt,
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
		&item.ChannelBindingID,
		&item.AttemptNo,
		&item.DeliveryID,
		&item.AdapterKind,
		&item.DeliveryRole,
		&item.Status,
		&item.RequestEnvelopeJSON,
		&item.AckPayloadJSON,
		&item.AdapterDeliveryID,
		&item.ProviderMessageRefJSON,
		&item.Retryable,
		&item.NextRetryAt,
		&item.LastErrorCode,
		&item.ContinuationReason,
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

func nullableInt64Value(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
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

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	result := value.UTC()
	return &result
}
