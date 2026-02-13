package mcpactionrequest

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/mcpactionrequest"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/mcpactionrequest/dbmodel"
)

func fromDBModel(row dbmodel.ActionRequestRow) domainrepo.Item {
	item := domainrepo.Item{
		ID:            row.ID,
		CorrelationID: row.CorrelationID,
		ToolName:      row.ToolName,
		Action:        row.Action,
		TargetRef:     row.TargetRef,
		ApprovalMode:  entitytypes.MCPApprovalMode(row.ApprovalMode),
		ApprovalState: entitytypes.MCPApprovalState(row.ApprovalState),
		RequestedBy:   row.RequestedBy,
		Payload:       row.Payload,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.RunID.Valid {
		item.RunID = row.RunID.String
	}
	if row.ProjectID.Valid {
		item.ProjectID = row.ProjectID.String
	}
	if row.ProjectSlug.Valid {
		item.ProjectSlug = row.ProjectSlug.String
	}
	if row.ProjectName.Valid {
		item.ProjectName = row.ProjectName.String
	}
	if row.IssueNumber.Valid && row.IssueNumber.Int32 > 0 {
		item.IssueNumber = int(row.IssueNumber.Int32)
	}
	if row.PRNumber.Valid && row.PRNumber.Int32 > 0 {
		item.PRNumber = int(row.PRNumber.Int32)
	}
	if row.TriggerLabel.Valid {
		item.TriggerLabel = row.TriggerLabel.String
	}
	if row.AppliedBy.Valid {
		item.AppliedBy = row.AppliedBy.String
	}
	return item
}
