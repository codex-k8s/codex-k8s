package runaccess

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	runaccesskeyrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runaccesskey"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

const (
	defaultTTL = 8 * time.Hour
	defaultMinTTL = 5 * time.Minute
	defaultMaxTTL = 24 * time.Hour
	defaultIssuer = "system"
	runAccessKeyPrefix = "rk_"
)

// RunAccessKeyRepository describes required run_access_keys repository behavior.
type RunAccessKeyRepository interface {
	GetByRunID(ctx context.Context, runID string) (runaccesskeyrepo.Run, bool, error)
	Upsert(ctx context.Context, params runaccesskeyrepo.UpsertParams) (runaccesskeyrepo.Run, error)
	Revoke(ctx context.Context, runID string, revokedAt time.Time, updatedAt time.Time) (runaccesskeyrepo.Run, bool, error)
	TouchLastUsed(ctx context.Context, runID string, usedAt time.Time) error
}

// StaffRunRepository describes required run metadata repository behavior.
type StaffRunRepository interface {
	GetByID(ctx context.Context, runID string) (staffrunrepo.Run, bool, error)
	GetLogsByRunID(ctx context.Context, runID string) (staffrunrepo.RunLogs, bool, error)
	ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]staffrunrepo.FlowEvent, error)
}

// FlowEventsRepository describes required audit events repository behavior.
type FlowEventsRepository interface {
	Insert(ctx context.Context, params floweventrepo.InsertParams) error
}

// Service provides run-scoped OAuth bypass key lifecycle and authorization checks.
type Service struct {
	cfg       Config
	keys      RunAccessKeyRepository
	runs      StaffRunRepository
	flowEvents FlowEventsRepository
	now       func() time.Time
}

// NewService creates run access key service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Keys == nil {
		return nil, fmt.Errorf("run access key repository is required")
	}
	if deps.Runs == nil {
		return nil, fmt.Errorf("staff run repository is required")
	}

	cfg = normalizeConfig(cfg)

	return &Service{
		cfg:       cfg,
		keys:      deps.Keys,
		runs:      deps.Runs,
		flowEvents: deps.FlowEvents,
		now:       time.Now,
	}, nil
}

// Issue creates or rotates run-scoped access key and returns plaintext value.
func (s *Service) Issue(ctx context.Context, params IssueParams) (IssuedKey, error) {
	return s.issueWithEvent(ctx, params, floweventdomain.EventTypeRunAccessKeyIssued)
}

// Regenerate rotates run-scoped access key and returns plaintext value.
func (s *Service) Regenerate(ctx context.Context, params IssueParams) (IssuedKey, error) {
	return s.issueWithEvent(ctx, params, floweventdomain.EventTypeRunAccessKeyRegenerated)
}

// GetStatus returns run access key lifecycle state.
func (s *Service) GetStatus(ctx context.Context, runID string) (KeyStatus, error) {
	run, err := s.requireRun(ctx, runID)
	if err != nil {
		return KeyStatus{}, err
	}

	stored, found, err := s.keys.GetByRunID(ctx, run.ID)
	if err != nil {
		return KeyStatus{}, fmt.Errorf("get run access key: %w", err)
	}
	if !found {
		return KeyStatus{
			RunID:         run.ID,
			ProjectID:     run.ProjectID,
			CorrelationID: run.CorrelationID,
			RuntimeMode:   "",
			Namespace:     run.Namespace,
			TargetEnv:     "",
			Status:        entitytypes.RunAccessKeyStatusMissing,
			HasKey:        false,
		}, nil
	}
	return mapStatus(stored, s.now().UTC()), nil
}

// Revoke marks key as revoked and returns resulting status snapshot.
func (s *Service) Revoke(ctx context.Context, runID string, revokedBy string) (KeyStatus, error) {
	run, err := s.requireRun(ctx, runID)
	if err != nil {
		return KeyStatus{}, err
	}

	now := s.now().UTC()
	stored, found, err := s.keys.Revoke(ctx, run.ID, now, now)
	if err != nil {
		return KeyStatus{}, fmt.Errorf("revoke run access key: %w", err)
	}
	if !found {
		return KeyStatus{
			RunID:         run.ID,
			ProjectID:     run.ProjectID,
			CorrelationID: run.CorrelationID,
			RuntimeMode:   "",
			Namespace:     run.Namespace,
			TargetEnv:     "",
			Status:        entitytypes.RunAccessKeyStatusMissing,
			HasKey:        false,
		}, nil
	}

	status := mapStatus(stored, now)
	s.insertFlowEvent(ctx, run.CorrelationID, floweventdomain.EventTypeRunAccessKeyRevoked, keyLifecyclePayload{
		RunID:       run.ID,
		ProjectID:   run.ProjectID,
		Namespace:   status.Namespace,
		RuntimeMode: status.RuntimeMode,
		TargetEnv:   status.TargetEnv,
		Status:      string(status.Status),
		RevokedBy:   normalizeIssuer(revokedBy),
		Reason:      "manual_revoke",
	})

	return status, nil
}

// GetRunByAccessKey authorizes bypass and returns one run details payload.
func (s *Service) GetRunByAccessKey(ctx context.Context, params AuthorizeParams) (staffrunrepo.Run, error) {
	authorized, err := s.AuthorizeBypass(ctx, params)
	if err != nil {
		return staffrunrepo.Run{}, err
	}
	return authorized.Run, nil
}

// ListRunEventsByAccessKey authorizes bypass and returns run events list.
func (s *Service) ListRunEventsByAccessKey(ctx context.Context, params AuthorizeParams, limit int) ([]staffrunrepo.FlowEvent, error) {
	authorized, err := s.AuthorizeBypass(ctx, params)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 500
	}
	if limit > 1000 {
		limit = 1000
	}
	items, err := s.runs.ListEventsByCorrelation(ctx, authorized.Run.CorrelationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list run events by access key: %w", err)
	}
	return items, nil
}

// GetRunLogsByAccessKey authorizes bypass and returns run logs snapshot.
func (s *Service) GetRunLogsByAccessKey(ctx context.Context, params AuthorizeParams) (staffrunrepo.RunLogs, error) {
	authorized, err := s.AuthorizeBypass(ctx, params)
	if err != nil {
		return staffrunrepo.RunLogs{}, err
	}
	logs, ok, err := s.runs.GetLogsByRunID(ctx, authorized.Run.ID)
	if err != nil {
		return staffrunrepo.RunLogs{}, fmt.Errorf("get run logs by access key: %w", err)
	}
	if !ok {
		return staffrunrepo.RunLogs{}, errs.Validation{Field: "run_id", Msg: "not found"}
	}
	return logs, nil
}

// AuthorizeBypass verifies run-scoped access key and returns authorized context.
func (s *Service) AuthorizeBypass(ctx context.Context, params AuthorizeParams) (AuthorizedContext, error) {
	run, err := s.requireRun(ctx, params.RunID)
	if err != nil {
		return AuthorizedContext{}, err
	}
	rawKey := strings.TrimSpace(params.AccessKey)
	if rawKey == "" {
		s.auditDenied(ctx, run, params, "missing_key")
		return AuthorizedContext{}, errs.Unauthorized{Msg: "missing run access key"}
	}

	stored, found, err := s.keys.GetByRunID(ctx, run.ID)
	if err != nil {
		return AuthorizedContext{}, fmt.Errorf("get run access key for authorization: %w", err)
	}
	if !found {
		s.auditDenied(ctx, run, params, "key_not_issued")
		return AuthorizedContext{}, errs.Unauthorized{Msg: "run access key is not issued"}
	}

	now := s.now().UTC()
	status := mapStatus(stored, now)
	if status.Status != entitytypes.RunAccessKeyStatusActive {
		s.auditDenied(ctx, run, params, "status_"+string(status.Status))
		return AuthorizedContext{}, errs.Unauthorized{Msg: "run access key is not active"}
	}
	if !equalAccessKeyHash(rawKey, stored.KeyHash) {
		s.auditDenied(ctx, run, params, "invalid_key")
		return AuthorizedContext{}, errs.Unauthorized{Msg: "invalid run access key"}
	}
	if stored.CorrelationID != "" && run.CorrelationID != stored.CorrelationID {
		s.auditDenied(ctx, run, params, "correlation_mismatch")
		return AuthorizedContext{}, errs.Unauthorized{Msg: "run access key context mismatch"}
	}
	if stored.ProjectID != "" && run.ProjectID != "" && stored.ProjectID != run.ProjectID {
		s.auditDenied(ctx, run, params, "project_mismatch")
		return AuthorizedContext{}, errs.Unauthorized{Msg: "run access key context mismatch"}
	}

	requestedNamespace := strings.TrimSpace(params.Namespace)
	if requestedNamespace != "" && status.Namespace != "" && requestedNamespace != status.Namespace {
		s.auditDenied(ctx, run, params, "namespace_mismatch")
		return AuthorizedContext{}, errs.Forbidden{Msg: "namespace mismatch"}
	}

	requestedTargetEnv := strings.TrimSpace(params.TargetEnv)
	if requestedTargetEnv != "" && status.TargetEnv != "" && requestedTargetEnv != status.TargetEnv {
		s.auditDenied(ctx, run, params, "target_env_mismatch")
		return AuthorizedContext{}, errs.Forbidden{Msg: "target environment mismatch"}
	}

	requestedRuntimeMode := strings.TrimSpace(params.RuntimeMode)
	if requestedRuntimeMode != "" && status.RuntimeMode != "" && requestedRuntimeMode != status.RuntimeMode {
		s.auditDenied(ctx, run, params, "runtime_mode_mismatch")
		return AuthorizedContext{}, errs.Forbidden{Msg: "runtime mode mismatch"}
	}

	if err := s.keys.TouchLastUsed(ctx, run.ID, now); err != nil {
		return AuthorizedContext{}, fmt.Errorf("touch run access key last_used_at: %w", err)
	}

	status.LastUsedAt = &now
	s.insertFlowEvent(ctx, run.CorrelationID, floweventdomain.EventTypeRunAccessKeyAuthorized, keyLifecyclePayload{
		RunID:       run.ID,
		ProjectID:   run.ProjectID,
		Namespace:   status.Namespace,
		RuntimeMode: status.RuntimeMode,
		TargetEnv:   status.TargetEnv,
		Status:      string(status.Status),
		ExpiresAt:   formatTime(status.ExpiresAt),
		Scope:       string(normalizeScope(params.Scope)),
		Reason:      "authorized",
	})

	return AuthorizedContext{Run: run, Status: status, Scope: normalizeScope(params.Scope)}, nil
}

func (s *Service) issueWithEvent(ctx context.Context, params IssueParams, eventType floweventdomain.EventType) (IssuedKey, error) {
	run, err := s.requireRun(ctx, params.RunID)
	if err != nil {
		return IssuedKey{}, err
	}

	now := s.now().UTC()
	ttl := normalizeTTL(params.TTL, s.cfg)
	expiresAt := now.Add(ttl)
	plainKey, err := generateAccessKey()
	if err != nil {
		return IssuedKey{}, err
	}

	namespace := strings.TrimSpace(params.Namespace)
	if namespace == "" {
		namespace = strings.TrimSpace(run.Namespace)
	}
	targetEnv := strings.TrimSpace(params.TargetEnv)
	runtimeMode := strings.TrimSpace(params.RuntimeMode)
	createdBy := normalizeIssuer(params.CreatedBy)

	stored, err := s.keys.Upsert(ctx, runaccesskeyrepo.UpsertParams{
		RunID:         run.ID,
		ProjectID:     run.ProjectID,
		CorrelationID: run.CorrelationID,
		RuntimeMode:   runtimeMode,
		Namespace:     namespace,
		TargetEnv:     targetEnv,
		KeyHash:       hashAccessKey(plainKey),
		Status:        string(entitytypes.RunAccessKeyStatusActive),
		IssuedAt:      now,
		ExpiresAt:     expiresAt,
		RevokedAt:     nil,
		LastUsedAt:    nil,
		CreatedBy:     createdBy,
		UpdatedAt:     now,
	})
	if err != nil {
		return IssuedKey{}, fmt.Errorf("upsert run access key: %w", err)
	}

	status := mapStatus(stored, now)
	s.insertFlowEvent(ctx, run.CorrelationID, eventType, keyLifecyclePayload{
		RunID:       run.ID,
		ProjectID:   run.ProjectID,
		Namespace:   status.Namespace,
		RuntimeMode: status.RuntimeMode,
		TargetEnv:   status.TargetEnv,
		Status:      string(status.Status),
		ExpiresAt:   formatTime(status.ExpiresAt),
		IssuedBy:    createdBy,
		Reason:      "issued",
	})

	return IssuedKey{AccessKey: plainKey, Status: status}, nil
}

func (s *Service) requireRun(ctx context.Context, runID string) (staffrunrepo.Run, error) {
	normalizedRunID := strings.TrimSpace(runID)
	if normalizedRunID == "" {
		return staffrunrepo.Run{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}
	run, ok, err := s.runs.GetByID(ctx, normalizedRunID)
	if err != nil {
		return staffrunrepo.Run{}, fmt.Errorf("get run by id: %w", err)
	}
	if !ok {
		return staffrunrepo.Run{}, errs.Validation{Field: "run_id", Msg: "not found"}
	}
	if strings.TrimSpace(run.CorrelationID) == "" {
		return staffrunrepo.Run{}, errs.Validation{Field: "run_id", Msg: "missing correlation"}
	}
	return run, nil
}

func (s *Service) auditDenied(ctx context.Context, run staffrunrepo.Run, params AuthorizeParams, reason string) {
	s.insertFlowEvent(ctx, run.CorrelationID, floweventdomain.EventTypeRunAccessKeyDenied, keyLifecyclePayload{
		RunID:       run.ID,
		ProjectID:   run.ProjectID,
		Namespace:   strings.TrimSpace(params.Namespace),
		RuntimeMode: strings.TrimSpace(params.RuntimeMode),
		TargetEnv:   strings.TrimSpace(params.TargetEnv),
		Scope:       string(normalizeScope(params.Scope)),
		Reason:      strings.TrimSpace(reason),
	})
}

func (s *Service) insertFlowEvent(ctx context.Context, correlationID string, eventType floweventdomain.EventType, payload keyLifecyclePayload) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlane,
		EventType:     eventType,
		Payload:       encodeKeyLifecyclePayload(payload),
		CreatedAt:     s.now().UTC(),
	})
}

func normalizeConfig(cfg Config) Config {
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = defaultTTL
	}
	if cfg.MinTTL <= 0 {
		cfg.MinTTL = defaultMinTTL
	}
	if cfg.MaxTTL <= 0 {
		cfg.MaxTTL = defaultMaxTTL
	}
	if cfg.MinTTL > cfg.MaxTTL {
		cfg.MinTTL = cfg.MaxTTL
	}
	if cfg.DefaultTTL < cfg.MinTTL {
		cfg.DefaultTTL = cfg.MinTTL
	}
	if cfg.DefaultTTL > cfg.MaxTTL {
		cfg.DefaultTTL = cfg.MaxTTL
	}
	return cfg
}

func normalizeTTL(value time.Duration, cfg Config) time.Duration {
	if value <= 0 {
		return cfg.DefaultTTL
	}
	if value < cfg.MinTTL {
		return cfg.MinTTL
	}
	if value > cfg.MaxTTL {
		return cfg.MaxTTL
	}
	return value
}

func normalizeScope(scope BypassScope) BypassScope {
	switch scope {
	case BypassScopeRunEvents:
		return BypassScopeRunEvents
	case BypassScopeRunLogs:
		return BypassScopeRunLogs
	default:
		return BypassScopeRunDetails
	}
}

func normalizeIssuer(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultIssuer
	}
	return value
}

func mapStatus(stored runaccesskeyrepo.Run, now time.Time) KeyStatus {
	status := stored.Status
	if status == entitytypes.RunAccessKeyStatusActive && !stored.ExpiresAt.IsZero() && now.After(stored.ExpiresAt) {
		status = entitytypes.RunAccessKeyStatusExpired
	}
	issuedAt := stored.IssuedAt.UTC()
	expiresAt := stored.ExpiresAt.UTC()
	return KeyStatus{
		RunID:         strings.TrimSpace(stored.RunID),
		ProjectID:     strings.TrimSpace(stored.ProjectID),
		CorrelationID: strings.TrimSpace(stored.CorrelationID),
		RuntimeMode:   strings.TrimSpace(stored.RuntimeMode),
		Namespace:     strings.TrimSpace(stored.Namespace),
		TargetEnv:     strings.TrimSpace(stored.TargetEnv),
		Status:        status,
		IssuedAt:      &issuedAt,
		ExpiresAt:     &expiresAt,
		RevokedAt:     normalizeTimePtr(stored.RevokedAt),
		LastUsedAt:    normalizeTimePtr(stored.LastUsedAt),
		CreatedBy:     strings.TrimSpace(stored.CreatedBy),
		HasKey:        true,
	}
}

func normalizeTimePtr(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	result := value.UTC()
	return &result
}

func formatTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func generateAccessKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate run access key: %w", err)
	}
	return runAccessKeyPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashAccessKey(value string) []byte {
	hash := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hash[:]
}

func equalAccessKeyHash(raw string, stored []byte) bool {
	if len(stored) == 0 {
		return false
	}
	calculated := hashAccessKey(raw)
	if len(calculated) != len(stored) {
		return false
	}
	return subtle.ConstantTimeCompare(calculated, stored) == 1
}
