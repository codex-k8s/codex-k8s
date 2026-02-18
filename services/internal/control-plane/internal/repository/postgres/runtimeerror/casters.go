package runtimeerror

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/runtimeerror/dbmodel"
)

func fromDBModel(row dbmodel.RuntimeErrorRow) domainrepo.Item {
	item := domainrepo.Item{
		ID:          row.ID,
		Source:      row.Source,
		Level:       row.Level,
		Message:     row.Message,
		DetailsJSON: row.DetailsJSON,
		CreatedAt:   row.CreatedAt,
	}
	if row.StackTrace.Valid {
		item.StackTrace = row.StackTrace.String
	}
	if row.CorrelationID.Valid {
		item.CorrelationID = row.CorrelationID.String
	}
	if row.RunID.Valid {
		item.RunID = row.RunID.String
	}
	if row.ProjectID.Valid {
		item.ProjectID = row.ProjectID.String
	}
	if row.Namespace.Valid {
		item.Namespace = row.Namespace.String
	}
	if row.JobName.Valid {
		item.JobName = row.JobName.String
	}
	if row.ViewedAt.Valid {
		value := row.ViewedAt.Time
		item.ViewedAt = &value
	}
	if row.ViewedBy.Valid {
		item.ViewedBy = row.ViewedBy.String
	}
	return item
}
