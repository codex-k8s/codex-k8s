package runaccess

import (
	"time"

	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// BypassScope defines one bypass operation audience.
type BypassScope string

const (
	BypassScopeRunDetails BypassScope = "run_details"
	BypassScopeRunEvents  BypassScope = "run_events"
	BypassScopeRunLogs    BypassScope = "run_logs"
)

// Config controls run access key lifecycle.
type Config struct {
	DefaultTTL time.Duration
	MinTTL     time.Duration
	MaxTTL     time.Duration
}

// Dependencies wires domain repositories.
type Dependencies struct {
	Keys      RunAccessKeyRepository
	Runs      StaffRunRepository
	FlowEvents FlowEventsRepository
}

// IssueParams describes one key issue/regenerate operation.
type IssueParams struct {
	RunID     string
	RuntimeMode string
	Namespace string
	TargetEnv string
	CreatedBy string
	TTL       time.Duration
}

// AuthorizeParams describes one bypass authorization request.
type AuthorizeParams struct {
	RunID      string
	AccessKey  string
	Scope      BypassScope
	Namespace  string
	TargetEnv  string
	RuntimeMode string
}

// KeyStatus is one run-scoped access key state snapshot.
type KeyStatus struct {
	RunID         string
	ProjectID     string
	CorrelationID string
	RuntimeMode   string
	Namespace     string
	TargetEnv     string
	Status        entitytypes.RunAccessKeyStatus
	IssuedAt      *time.Time
	ExpiresAt     *time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
	CreatedBy     string
	HasKey        bool
}

// IssuedKey contains plaintext key returned only on issue/regenerate operations.
type IssuedKey struct {
	AccessKey string
	Status    KeyStatus
}

// AuthorizedContext is returned after successful bypass authorization.
type AuthorizedContext struct {
	Run    staffrunrepo.Run
	Status KeyStatus
	Scope  BypassScope
}
