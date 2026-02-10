package provider

import "context"

// Provider defines a repository hosting provider identifier.
type Provider string

const (
	// ProviderGitHub is the GitHub provider id used in DB and API.
	ProviderGitHub Provider = "github"
	// ProviderGitLab is reserved for future GitLab support.
	ProviderGitLab Provider = "gitlab"
)

// WebhookSpec describes how a provider webhook should be configured.
type WebhookSpec struct {
	// URL is a public callback URL, e.g. https://staging.codex-k8s.dev/api/v1/webhooks/github.
	URL string
	// Secret is the shared secret used to sign payloads.
	Secret string
	// Events is a list of provider event names to subscribe to.
	Events []string
}

// RepositoryInfo is basic repo metadata returned by a provider.
type RepositoryInfo struct {
	Provider   Provider
	Owner      string
	Name       string
	FullName   string
	Private    bool
	ExternalID int64
}

// RepositoryProvider is an interface to repository hosting services (GitHub first, GitLab-ready).
//
// Domain code must rely on this interface instead of importing vendor SDK packages.
type RepositoryProvider interface {
	// ValidateRepository checks that the token has access to the repo and returns basic metadata.
	ValidateRepository(ctx context.Context, token string, owner string, name string) (RepositoryInfo, error)
	// EnsureWebhook ensures a webhook with desired spec exists on the repo.
	EnsureWebhook(ctx context.Context, token string, owner string, name string, spec WebhookSpec) error
	// DeleteWebhook attempts to delete codex-k8s webhook(s) matching webhookURL.
	//
	// Callers should treat errors as best-effort failures (tokens may be revoked, permissions missing, etc).
	DeleteWebhook(ctx context.Context, token string, owner string, name string, webhookURL string) error
}
