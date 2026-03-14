package dbmodel

import "github.com/jackc/pgx/v5/pgtype"

// WaitRow mirrors one github_rate_limit_waits row.
type WaitRow struct {
	ID                     string             `db:"id"`
	ProjectID              string             `db:"project_id"`
	RunID                  string             `db:"run_id"`
	ContourKind            string             `db:"contour_kind"`
	SignalOrigin           string             `db:"signal_origin"`
	OperationClass         string             `db:"operation_class"`
	State                  string             `db:"state"`
	LimitKind              string             `db:"limit_kind"`
	Confidence             string             `db:"confidence"`
	RecoveryHintKind       string             `db:"recovery_hint_kind"`
	DominantForRun         bool               `db:"dominant_for_run"`
	SignalID               string             `db:"signal_id"`
	RequestFingerprint     pgtype.Text        `db:"request_fingerprint"`
	CorrelationID          string             `db:"correlation_id"`
	ResumeActionKind       string             `db:"resume_action_kind"`
	ResumePayloadJSON      []byte             `db:"resume_payload_json"`
	ManualActionKind       pgtype.Text        `db:"manual_action_kind"`
	AutoResumeAttemptsUsed int32              `db:"auto_resume_attempts_used"`
	MaxAutoResumeAttempts  int32              `db:"max_auto_resume_attempts"`
	ResumeNotBefore        pgtype.Timestamptz `db:"resume_not_before"`
	LastResumeAttemptAt    pgtype.Timestamptz `db:"last_resume_attempt_at"`
	FirstDetectedAt        pgtype.Timestamptz `db:"first_detected_at"`
	LastSignalAt           pgtype.Timestamptz `db:"last_signal_at"`
	ResolvedAt             pgtype.Timestamptz `db:"resolved_at"`
	CreatedAt              pgtype.Timestamptz `db:"created_at"`
	UpdatedAt              pgtype.Timestamptz `db:"updated_at"`
}
