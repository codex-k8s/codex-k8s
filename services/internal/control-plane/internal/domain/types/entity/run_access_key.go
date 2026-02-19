package entity

import "time"

// RunAccessKeyStatus keeps normalized run-access key lifecycle state.
type RunAccessKeyStatus string

const (
	RunAccessKeyStatusActive  RunAccessKeyStatus = "active"
	RunAccessKeyStatusRevoked RunAccessKeyStatus = "revoked"
	RunAccessKeyStatusExpired RunAccessKeyStatus = "expired"
	RunAccessKeyStatusMissing RunAccessKeyStatus = "missing"
)

// RunAccessKey stores one run-scoped OAuth bypass key metadata.
type RunAccessKey struct {
	RunID         string
	ProjectID     string
	CorrelationID string
	RuntimeMode   string
	Namespace     string
	TargetEnv     string
	KeyHash       []byte
	Status        RunAccessKeyStatus
	IssuedAt      time.Time
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
