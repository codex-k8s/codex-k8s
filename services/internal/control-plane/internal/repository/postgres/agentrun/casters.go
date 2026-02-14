package agentrun

import (
	"encoding/json"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agentrun/dbmodel"
)

func runFromDBModel(row dbmodel.RunRow) domainrepo.Run {
	item := domainrepo.Run{
		ID:            row.ID,
		CorrelationID: row.CorrelationID,
		Status:        row.Status,
		RunPayload:    json.RawMessage(row.RunPayload),
	}
	if row.ProjectID.Valid {
		item.ProjectID = row.ProjectID.String
	}
	return item
}

func runLookupItemFromDBModel(row dbmodel.RunLookupRow) domainrepo.RunLookupItem {
	item := domainrepo.RunLookupItem{
		RunID:              row.RunID,
		CorrelationID:      row.CorrelationID,
		RepositoryFullName: row.RepositoryFullName,
		AgentKey:           row.AgentKey,
		IssueURL:           row.IssueURL,
		PullRequestURL:     row.PullRequestURL,
		TriggerKind:        row.TriggerKind,
		TriggerLabel:       row.TriggerLabel,
		Status:             row.Status,
		CreatedAt:          row.CreatedAt.UTC(),
	}
	if row.ProjectID.Valid {
		item.ProjectID = row.ProjectID.String
	}
	if row.IssueNumber.Valid {
		item.IssueNumber = row.IssueNumber.Int64
	}
	if row.PullRequestNumber.Valid {
		item.PullRequestNumber = row.PullRequestNumber.Int64
	}
	if row.StartedAt.Valid {
		startedAt := row.StartedAt.Time.UTC()
		item.StartedAt = &startedAt
	}
	if row.FinishedAt.Valid {
		finishedAt := row.FinishedAt.Time.UTC()
		item.FinishedAt = &finishedAt
	}
	return item
}
