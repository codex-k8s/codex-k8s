package projectdatabase

import (
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectdatabase"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projectdatabase/dbmodel"
)

func fromDBModel(row dbmodel.ProjectDatabaseRow) domainrepo.Item {
	return domainrepo.Item{
		ProjectID:    row.ProjectID,
		Environment:  row.Environment,
		DatabaseName: row.DatabaseName,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
