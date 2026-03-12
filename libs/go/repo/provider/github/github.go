package github

import (
	"context"
	"errors"
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

	desired := &gh.Hook{
		Active: gh.Ptr(true),
		Events: normalizeEvents(spec.Events),
		Config: hookConfig(spec),
	}

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
			_, _, err := client.Repositories.EditHook(ctx, owner, name, h.GetID(), desired)
			if err != nil {
				return fmt.Errorf("github edit hook %s/%s id=%d: %w", owner, name, h.GetID(), err)
			}
			return nil
		}
	}

	_, _, err = client.Repositories.CreateHook(ctx, owner, name, desired)
	if err != nil {
		return fmt.Errorf("github create hook %s/%s: %w", owner, name, err)
	}

	return nil
}

// DeleteWebhook deletes GitHub repository webhook(s) that match webhookURL.
func (p *Provider) DeleteWebhook(ctx context.Context, token string, owner string, name string, webhookURL string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return nil
	}

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
		if strings.EqualFold(cfg.GetURL(), webhookURL) {
			_, err := client.Repositories.DeleteHook(ctx, owner, name, h.GetID())
			if err != nil {
				return fmt.Errorf("github delete hook %s/%s id=%d: %w", owner, name, h.GetID(), err)
			}
		}
	}

	return nil
}

// GetPullRequest returns one GitHub pull request by number.
func (p *Provider) GetPullRequest(ctx context.Context, token string, owner string, name string, pullRequestNumber int) (provider.PullRequestInfo, bool, error) {
	if pullRequestNumber <= 0 {
		return provider.PullRequestInfo{}, false, nil
	}

	client := gh.NewClient(p.httpClient).WithAuthToken(token)
	item, _, err := client.PullRequests.Get(ctx, owner, name, pullRequestNumber)
	if err != nil {
		if isNotFound(err) {
			return provider.PullRequestInfo{}, false, nil
		}
		return provider.PullRequestInfo{}, false, fmt.Errorf("github get pull request %s/%s#%d: %w", owner, name, pullRequestNumber, err)
	}

	return toPullRequestInfo(item), true, nil
}

// FindPullRequestByHead returns the most recent pull request for one head branch.
func (p *Provider) FindPullRequestByHead(ctx context.Context, token string, owner string, name string, headRef string) (provider.PullRequestInfo, bool, error) {
	trimmedHeadRef := strings.TrimSpace(headRef)
	if trimmedHeadRef == "" {
		return provider.PullRequestInfo{}, false, nil
	}

	client := gh.NewClient(p.httpClient).WithAuthToken(token)
	headFilter := trimmedHeadRef
	if !strings.Contains(headFilter, ":") {
		headFilter = strings.TrimSpace(owner) + ":" + headFilter
	}

	items, _, err := client.PullRequests.List(ctx, owner, name, &gh.PullRequestListOptions{
		State: "all",
		Head:  headFilter,
		Sort:  "updated",
		ListOptions: gh.ListOptions{
			PerPage: 20,
		},
	})
	if err != nil {
		return provider.PullRequestInfo{}, false, fmt.Errorf("github list pull requests %s/%s head=%s: %w", owner, name, trimmedHeadRef, err)
	}
	if len(items) == 0 {
		return provider.PullRequestInfo{}, false, nil
	}

	return toPullRequestInfo(items[0]), true, nil
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

func hookConfig(spec provider.WebhookSpec) *gh.HookConfig {
	return &gh.HookConfig{
		URL:         gh.Ptr(spec.URL),
		ContentType: gh.Ptr("json"),
		Secret:      gh.Ptr(spec.Secret),
	}
}

func isNotFound(err error) bool {
	var apiErr *gh.ErrorResponse
	return errors.As(err, &apiErr) && apiErr.Response != nil && apiErr.Response.StatusCode == http.StatusNotFound
}

func toPullRequestInfo(item *gh.PullRequest) provider.PullRequestInfo {
	if item == nil {
		return provider.PullRequestInfo{}
	}

	return provider.PullRequestInfo{
		Number: item.GetNumber(),
		URL:    item.GetHTMLURL(),
		State:  item.GetState(),
		Head:   item.GetHead().GetRef(),
		Base:   item.GetBase().GetRef(),
	}
}
