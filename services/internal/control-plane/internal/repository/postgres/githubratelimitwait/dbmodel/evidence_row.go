package dbmodel

import "github.com/jackc/pgx/v5/pgtype"

// EvidenceRow mirrors one github_rate_limit_wait_evidence row.
type EvidenceRow struct {
	ID                 int64              `db:"id"`
	WaitID             string             `db:"wait_id"`
	EventKind          string             `db:"event_kind"`
	SignalID           pgtype.Text        `db:"signal_id"`
	SignalOrigin       pgtype.Text        `db:"signal_origin"`
	ProviderStatusCode pgtype.Int4        `db:"provider_status_code"`
	RetryAfterSeconds  pgtype.Int4        `db:"retry_after_seconds"`
	RateLimitLimit     pgtype.Int4        `db:"rate_limit_limit"`
	RateLimitRemaining pgtype.Int4        `db:"rate_limit_remaining"`
	RateLimitUsed      pgtype.Int4        `db:"rate_limit_used"`
	RateLimitResetAt   pgtype.Timestamptz `db:"rate_limit_reset_at"`
	RateLimitResource  pgtype.Text        `db:"rate_limit_resource"`
	GitHubRequestID    pgtype.Text        `db:"github_request_id"`
	DocumentationURL   pgtype.Text        `db:"documentation_url"`
	MessageExcerpt     pgtype.Text        `db:"message_excerpt"`
	StderrExcerpt      pgtype.Text        `db:"stderr_excerpt"`
	PayloadJSON        []byte             `db:"payload_json"`
	ObservedAt         pgtype.Timestamptz `db:"observed_at"`
	CreatedAt          pgtype.Timestamptz `db:"created_at"`
}
