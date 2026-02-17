package value

import (
	"net"
	"time"
)

// GitHubPreflightParams describes preflight inputs for a GitHub repository onboarding check.
type GitHubPreflightParams struct {
	PlatformToken string
	BotToken      string
	Owner         string
	Repository    string

	WebhookURL    string
	WebhookSecret string

	ExpectedDomains []string
	DNSExpectedIPs  []net.IP
}

type GitHubPreflightCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

type GitHubPreflightTokenScopes struct {
	Platform string `json:"platform,omitempty"` // repository|project|platform
	Bot      string `json:"bot,omitempty"`      // repository|project|platform
}

type GitHubPreflightArtifact struct {
	Kind           string `json:"kind"`                     // webhook|label|issue|branch|pr
	Name           string `json:"name,omitempty"`           // label/branch title, etc.
	ID             string `json:"id,omitempty"`             // numeric id/number as string.
	CleanupStatus  string `json:"cleanup_status,omitempty"` // ok|failed|skipped
	CleanupDetails string `json:"cleanup_details,omitempty"`
}

type GitHubPreflightReport struct {
	Status      string                     `json:"status"`
	TokenScopes GitHubPreflightTokenScopes `json:"token_scopes,omitempty"`
	Checks      []GitHubPreflightCheck     `json:"checks"`
	Artifacts   []GitHubPreflightArtifact  `json:"artifacts,omitempty"`
	FinishedAt  time.Time                  `json:"finished_at"`
}
