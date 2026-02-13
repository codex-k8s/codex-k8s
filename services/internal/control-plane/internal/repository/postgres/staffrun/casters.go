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
	if row.AgentKey.Valid {
		item.AgentKey = row.AgentKey.String
	}
	if row.JobName.Valid {
		item.JobName = row.JobName.String
	}
	if row.JobNamespace.Valid {
		item.JobNamespace = row.JobNamespace.String
	}
	if row.Namespace.Valid {
		item.Namespace = row.Namespace.String
	}
	if row.WaitState.Valid {
		item.WaitState = row.WaitState.String
	}
	if row.WaitReason.Valid {
		item.WaitReason = row.WaitReason.String
	}
	if row.WaitSince.Valid {
		v := row.WaitSince.Time
		item.WaitSince = &v
	}
	if row.LastHeartbeat.Valid {
		v := row.LastHeartbeat.Time
		item.LastHeartbeatAt = &v
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

func runLogsFromDBModel(row dbmodel.RunLogsRow) domainrepo.RunLogs {
	item := domainrepo.RunLogs{
		RunID:        row.RunID,
		Status:       row.Status,
		SnapshotJSON: row.SnapshotJSON,
	}
	if row.UpdatedAt.Valid {
		v := row.UpdatedAt.Time
		item.UpdatedAt = &v
	}
	return item
}

func flowEventFromDBModel(row dbmodel.FlowEventRow) domainrepo.FlowEvent {
	return domainrepo.FlowEvent{
		CorrelationID: row.CorrelationID,
		EventType:     row.EventType,
		CreatedAt:     row.CreatedAt,
		PayloadJSON:   []byte(row.PayloadText),
	}
}
