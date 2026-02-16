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

type GitHubPreflightReport struct {
	Status     string             `json:"status"`
	Checks     []GitHubPreflightCheck `json:"checks"`
	FinishedAt time.Time          `json:"finished_at"`
}

