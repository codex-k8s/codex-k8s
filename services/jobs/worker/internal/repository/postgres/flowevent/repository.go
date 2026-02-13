package flowevent

import (
	libflow "github.com/codex-k8s/codex-k8s/libs/go/postgres/floweventrepo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository = libflow.Repository

func NewRepository(db *pgxpool.Pool) *Repository { return libflow.NewRepository(db) }
