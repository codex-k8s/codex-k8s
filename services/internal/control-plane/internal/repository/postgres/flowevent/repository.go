package flowevent

import (
	"database/sql"
	libflow "github.com/codex-k8s/codex-k8s/libs/go/postgres/floweventrepo"
)

type Repository = libflow.Repository

func NewRepository(db *sql.DB) *Repository { return libflow.NewRepository(db) }
