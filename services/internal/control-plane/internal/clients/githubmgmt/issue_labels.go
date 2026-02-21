package githubmgmt

import (
	"context"
	"fmt"
	"slices"
	"strings"

	gh "github.com/google/go-github/v82/github"
)

// ListIssueLabels returns normalized issue labels.
func (c *Client) ListIssueLabels(ctx context.Context, token string, owner string, repo string, issueNumber int) ([]string, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repository are required")
	}
	if issueNumber <= 0 {
		return nil, fmt.Errorf("issue_number must be positive")
	}

	client := c.clientWithToken(token)
	items, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, issueNumber, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("github list issue labels %s/%s#%d: %w", owner, repo, issueNumber, err)
	}
	return normalizeIssueLabels(itemsToLabelNames(items)), nil
}

// AddIssueLabels appends labels to one issue and returns resulting normalized label set.
func (c *Client) AddIssueLabels(ctx context.Context, token string, owner string, repo string, issueNumber int, labels []string) ([]string, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repository are required")
	}
	if issueNumber <= 0 {
		return nil, fmt.Errorf("issue_number must be positive")
	}
	normalizedLabels := normalizeIssueLabels(labels)
	if len(normalizedLabels) == 0 {
		return nil, fmt.Errorf("labels are required")
	}

	client := c.clientWithToken(token)
	items, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, issueNumber, normalizedLabels)
	if err != nil {
		return nil, fmt.Errorf("github add issue labels %s/%s#%d: %w", owner, repo, issueNumber, err)
	}
	return normalizeIssueLabels(itemsToLabelNames(items)), nil
}

// RemoveIssueLabel removes one label from issue; missing labels are treated as already removed.
func (c *Client) RemoveIssueLabel(ctx context.Context, token string, owner string, repo string, issueNumber int, label string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repository are required")
	}
	if issueNumber <= 0 {
		return fmt.Errorf("issue_number must be positive")
	}
	normalizedLabel := strings.ToLower(strings.TrimSpace(label))
	if normalizedLabel == "" {
		return fmt.Errorf("label is required")
	}

	client := c.clientWithToken(token)
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, issueNumber, normalizedLabel)
	if err != nil {
		if isGitHubNotFound(err) {
			return nil
		}
		return fmt.Errorf("github remove issue label %q %s/%s#%d: %w", normalizedLabel, owner, repo, issueNumber, err)
	}
	return nil
}

func itemsToLabelNames(items []*gh.Label) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		label := strings.TrimSpace(item.GetName())
		if label == "" {
			continue
		}
		out = append(out, label)
	}
	return out
}

func normalizeIssueLabels(labels []string) []string {
	out := make([]string, 0, len(labels))
	seen := make(map[string]struct{}, len(labels))
	for _, item := range labels {
		label := strings.ToLower(strings.TrimSpace(item))
		if label == "" {
			continue
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
	}
	slices.Sort(out)
	return out
}
