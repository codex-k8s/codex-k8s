package query

import "time"

// RepositoryBotParamsUpsertParams defines bot params update payload for one repository.
type RepositoryBotParamsUpsertParams struct {
	RepositoryID      string
	BotTokenEncrypted []byte
	BotUsername       string
	BotEmail          string
}

// RepositoryPreflightReportUpsertParams defines persisted preflight report payload.
type RepositoryPreflightReportUpsertParams struct {
	RepositoryID string
	ReportJSON   []byte
}

// RepositoryPreflightLockAcquireParams defines lock acquisition payload for repository preflight.
type RepositoryPreflightLockAcquireParams struct {
	RepositoryID   string
	LockToken      string
	LockedByUserID string
	LockedUntilUTC time.Time
}
