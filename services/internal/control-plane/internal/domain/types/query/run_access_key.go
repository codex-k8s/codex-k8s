package query

import "time"

// RunAccessKeyUpsertParams defines run access key upsert payload.
type RunAccessKeyUpsertParams struct {
	RunID         string
	ProjectID     string
	CorrelationID string
	RuntimeMode   string
	Namespace     string
	TargetEnv     string
	KeyHash       []byte
	Status        string
	IssuedAt      time.Time
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
	CreatedBy     string
	UpdatedAt     time.Time
}
