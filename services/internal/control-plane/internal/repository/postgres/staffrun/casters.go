package staffrun

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/staffrun/dbmodel"
)

func runFromDBModel(row dbmodel.RunRow) domainrepo.Run {
	item := domainrepo.Run{
		ID:            row.ID,
		CorrelationID: row.CorrelationID,
		ProjectSlug:   row.ProjectSlug,
		ProjectName:   row.ProjectName,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
	}
	if row.ProjectID.Valid {
		item.ProjectID = row.ProjectID.String
	}
	if row.IssueNumber.Valid && row.IssueNumber.Int32 > 0 {
		item.IssueNumber = int(row.IssueNumber.Int32)
	}
	if row.IssueURL.Valid {
		item.IssueURL = row.IssueURL.String
	}
	if row.TriggerKind.Valid {
		item.TriggerKind = row.TriggerKind.String
	}
	if row.TriggerLabel.Valid {
		item.TriggerLabel = row.TriggerLabel.String
	}
	if row.PRURL.Valid {
		item.PRURL = row.PRURL.String
	}
	if row.PRNumber.Valid && row.PRNumber.Int32 > 0 {
		item.PRNumber = int(row.PRNumber.Int32)
	}
	if row.StartedAt.Valid {
		v := row.StartedAt.Time
		item.StartedAt = &v
	}
	if row.FinishedAt.Valid {
		v := row.FinishedAt.Time
		item.FinishedAt = &v
	}
	return item
}
