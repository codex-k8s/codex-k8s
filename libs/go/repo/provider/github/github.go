package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v82/github"

	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
)

// Provider implements RepositoryProvider for GitHub REST API v3.
type Provider struct {
	httpClient *http.Client
}

// NewProvider constructs a GitHub repository provider.
func NewProvider(httpClient *http.Client) *Provider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Provider{httpClient: httpClient}
}

// ValidateRepository verifies repo access using the provided token and returns metadata.
func (p *Provider) ValidateRepository(ctx context.Context, token string, owner string, name string) (provider.RepositoryInfo, error) {
	client := gh.NewClient(p.httpClient).WithAuthToken(token)

	repo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return provider.RepositoryInfo{}, fmt.Errorf("github get repository %s/%s: %w", owner, name, err)
	}

	fullName := strings.TrimSpace(repo.GetFullName())
	if fullName == "" {
		fullName = strings.TrimSpace(owner) + "/" + strings.TrimSpace(name)
	}

	return provider.RepositoryInfo{
		Provider:   provider.ProviderGitHub,
		Owner:      strings.TrimSpace(owner),
		Name:       strings.TrimSpace(name),
		FullName:   fullName,
		Private:    repo.GetPrivate(),
		ExternalID: repo.GetID(),
	}, nil
}

// EnsureWebhook makes sure a webhook exists and is configured for codex-k8s.
func (p *Provider) EnsureWebhook(ctx context.Context, token string, owner string, name string, spec provider.WebhookSpec) error {
	client := gh.NewClient(p.httpClient).WithAuthToken(token)

	hooks, _, err := client.Repositories.ListHooks(ctx, owner, name, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return fmt.Errorf("github list hooks %s/%s: %w", owner, name, err)
	}

	for _, h := range hooks {
		cfg := h.GetConfig()
		if cfg == nil {
			continue
		}
		if strings.EqualFold(cfg.GetURL(), spec.URL) {
			// If URL matches, we assume the hook is ours. Update config and events to desired state.
			_, _, err := client.Repositories.EditHook(ctx, owner, name, h.GetID(), &gh.Hook{
				Active: gh.Ptr(true),
				Events: normalizeEvents(spec.Events),
				Config: &gh.HookConfig{
					URL:         gh.Ptr(spec.URL),
					ContentType: gh.Ptr("json"),
					Secret:      gh.Ptr(spec.Secret),
				},
			})
			if err != nil {
				return fmt.Errorf("github edit hook %s/%s id=%d: %w", owner, name, h.GetID(), err)
			}
			return nil
		}
	}

	_, _, err = client.Repositories.CreateHook(ctx, owner, name, &gh.Hook{
		Active: gh.Ptr(true),
		Events: normalizeEvents(spec.Events),
		Config: &gh.HookConfig{
			URL:         gh.Ptr(spec.URL),
			ContentType: gh.Ptr("json"),
			Secret:      gh.Ptr(spec.Secret),
		},
	})
	if err != nil {
		return fmt.Errorf("github create hook %s/%s: %w", owner, name, err)
	}

	return nil
}

func normalizeEvents(in []string) []string {
	out := make([]string, 0, len(in))
	for _, e := range in {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		out = append(out, e)
	}
	if len(out) == 0 {
		// GitHub default is "push", but codex-k8s expects more. We keep it minimal here.
		out = []string{"push"}
	}
	return out
}
