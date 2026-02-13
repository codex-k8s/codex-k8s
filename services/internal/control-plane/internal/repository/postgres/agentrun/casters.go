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
