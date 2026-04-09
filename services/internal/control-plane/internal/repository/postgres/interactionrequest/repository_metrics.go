package interactionrequest

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/repository/postgres/interactionrequest/dbmodel"
	"github.com/jackc/pgx/v5"
)

var (
	//go:embed sql/count_request_states.sql
	queryCountRequestStates string
	//go:embed sql/count_pending_dispatch_backlog.sql
	queryCountPendingDispatchBacklog string
	//go:embed sql/count_overdue_waits.sql
	queryCountOverdueWaits string
	//go:embed sql/count_callback_events.sql
	queryCountCallbackEvents string
	//go:embed sql/count_dispatch_attempts.sql
	queryCountDispatchAttempts string
)

// LoadMetricsSnapshot returns current persisted interaction observability totals.
func (r *Repository) LoadMetricsSnapshot(ctx context.Context, now time.Time) (MetricsSnapshot, error) {
	requestStateRows, err := r.queryRequestStateTotals(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	pendingDispatchRows, err := r.queryPendingDispatchBacklog(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	overdueWaitRows, err := r.queryOverdueWaitTotals(ctx, timeOrNow(now))
	if err != nil {
		return MetricsSnapshot{}, err
	}
	callbackRows, err := r.queryCallbackEventTotals(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	dispatchRows, err := r.queryDispatchAttemptTotals(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}

	return MetricsSnapshot{
		CollectedAt:            timeOrNow(now),
		RequestStateTotals:     requestStateRows,
		PendingDispatchBacklog: pendingDispatchRows,
		OverdueWaitTotals:      overdueWaitRows,
		CallbackEventTotals:    callbackRows,
		DispatchAttemptTotals:  dispatchRows,
	}, nil
}

func (r *Repository) queryRequestStateTotals(ctx context.Context) ([]RequestStateTotal, error) {
	return queryMetricRows(ctx, r, queryCountRequestStates, "interaction request state totals", requestStateTotalFromRow)
}

func (r *Repository) queryPendingDispatchBacklog(ctx context.Context) ([]PendingDispatchBacklog, error) {
	return queryMetricRows(ctx, r, queryCountPendingDispatchBacklog, "interaction pending dispatch backlog", pendingDispatchBacklogFromRow)
}

func (r *Repository) queryOverdueWaitTotals(ctx context.Context, now time.Time) ([]OverdueWaitTotal, error) {
	return queryMetricRows(ctx, r, queryCountOverdueWaits, "interaction overdue waits", overdueWaitTotalFromRow, now.UTC())
}

func (r *Repository) queryCallbackEventTotals(ctx context.Context) ([]CallbackEventTotal, error) {
	return queryMetricRows(ctx, r, queryCountCallbackEvents, "interaction callback event totals", callbackEventTotalFromRow)
}

func (r *Repository) queryDispatchAttemptTotals(ctx context.Context) ([]DispatchAttemptTotal, error) {
	return queryMetricRows(ctx, r, queryCountDispatchAttempts, "interaction dispatch attempt totals", dispatchAttemptTotalFromRow)
}

func queryMetricRows[Row any, Out any](ctx context.Context, r *Repository, query string, operation string, mapper func(Row) Out, args ...any) ([]Out, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", operation, err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[Row])
	if err != nil {
		return nil, fmt.Errorf("collect %s: %w", operation, err)
	}

	result := make([]Out, 0, len(items))
	for _, item := range items {
		result = append(result, mapper(item))
	}
	return result, nil
}

func requestStateTotalFromRow(item dbmodel.RequestStateMetricRow) RequestStateTotal {
	return RequestStateTotal{
		InteractionKind: item.InteractionKind,
		State:           item.State,
		Total:           item.Total,
	}
}

func pendingDispatchBacklogFromRow(item dbmodel.PendingDispatchBacklogMetricRow) PendingDispatchBacklog {
	return PendingDispatchBacklog{
		InteractionKind: item.InteractionKind,
		QueueKind:       item.QueueKind,
		Total:           item.Total,
	}
}

func overdueWaitTotalFromRow(item dbmodel.OverdueWaitMetricRow) OverdueWaitTotal {
	return OverdueWaitTotal{
		InteractionKind: item.InteractionKind,
		Total:           item.Total,
	}
}

func callbackEventTotalFromRow(item dbmodel.CallbackEventMetricRow) CallbackEventTotal {
	return CallbackEventTotal{
		CallbackKind:   item.CallbackKind,
		Classification: item.Classification,
		Total:          item.Total,
	}
}

func dispatchAttemptTotalFromRow(item dbmodel.DispatchAttemptMetricRow) DispatchAttemptTotal {
	return DispatchAttemptTotal{
		InteractionKind: item.InteractionKind,
		AdapterKind:     item.AdapterKind,
		Status:          item.Status,
		Total:           item.Total,
	}
}
