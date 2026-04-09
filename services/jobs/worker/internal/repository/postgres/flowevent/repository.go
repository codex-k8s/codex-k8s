package flowevent

import (
	libflow "github.com/codex-k8s/kodex/libs/go/postgres/floweventrepo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository = libflow.Repository

func NewRepository(db *pgxpool.Pool) *Repository { return libflow.NewRepository(db) }
