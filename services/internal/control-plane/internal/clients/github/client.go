package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v82/github"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
)

const (
	defaultPageSize = 100
	maxPageSize     = 100
)

// Client wraps go-github for MCP tool operations.
type Client struct {
	httpClient *http.Client
}

// NewClient constructs a GitHub MCP adapter.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) GetIssue(ctx context.Context, params mcpdomain.GitHubGetIssueParams) (mcpdomain.GitHubIssue, error) {
	client := c.clientWithToken(params.Token)
	item, _, err := client.Issues.Get(ctx, params.Owner, params.Repository, params.IssueNumber)
	if err != nil {
		return mcpdomain.GitHubIssue{}, err
	}
	return mcpdomain.GitHubIssue{
		Number: item.GetNumber(),
		Title:  item.GetTitle(),
		State:  item.GetState(),
		URL:    item.GetHTMLURL(),
	}, nil
}

func (c *Client) GetPullRequest(ctx context.Context, params mcpdomain.GitHubGetPullRequestParams) (mcpdomain.GitHubPullRequest, error) {
	client := c.clientWithToken(params.Token)
	item, _, err := client.PullRequests.Get(ctx, params.Owner, params.Repository, params.PullRequestNumber)
	if err != nil {
		return mcpdomain.GitHubPullRequest{}, err
	}
	return mcpdomain.GitHubPullRequest{
		Number: item.GetNumber(),
		Title:  item.GetTitle(),
		State:  item.GetState(),
		URL:    item.GetHTMLURL(),
		Head:   item.GetHead().GetRef(),
		Base:   item.GetBase().GetRef(),
	}, nil
}

func (c *Client) GetAuthenticatedUserLogin(ctx context.Context, token string) (string, error) {
	client := c.clientWithToken(token)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	login := strings.TrimSpace(user.GetLogin())
	if login == "" {
		return "", fmt.Errorf("github token owner login is empty")
	}
	return login, nil
}

func (c *Client) ListIssueComments(ctx context.Context, params mcpdomain.GitHubListIssueCommentsParams) ([]mcpdomain.GitHubIssueComment, error) {
	client := c.clientWithToken(params.Token)
	limit := clampLimit(params.Limit, defaultPageSize, maxPageSize)

	items, _, err := client.Issues.ListComments(ctx, params.Owner, params.Repository, params.IssueNumber, &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: limit},
	})
	if err != nil {
		return nil, err
	}

	out := make([]mcpdomain.GitHubIssueComment, 0, len(items))
	for _, item := range items {
		out = append(out, mcpdomain.GitHubIssueComment{
			ID:   item.GetID(),
			Body: item.GetBody(),
			URL:  item.GetHTMLURL(),
			User: item.GetUser().GetLogin(),
		})
	}
	return out, nil
}

func (c *Client) ListIssueLabels(ctx context.Context, params mcpdomain.GitHubListIssueLabelsParams) ([]mcpdomain.GitHubLabel, error) {
	client := c.clientWithToken(params.Token)

	items, _, err := client.Issues.ListLabelsByIssue(ctx, params.Owner, params.Repository, params.IssueNumber, &gh.ListOptions{
		PerPage: maxPageSize,
	})
	if err != nil {
		return nil, err
	}

	out := make([]mcpdomain.GitHubLabel, 0, len(items))
	for _, item := range items {
		out = append(out, mcpdomain.GitHubLabel{Name: item.GetName()})
	}
	return out, nil
}

func (c *Client) ListBranches(ctx context.Context, params mcpdomain.GitHubListBranchesParams) ([]mcpdomain.GitHubBranch, error) {
	client := c.clientWithToken(params.Token)
	limit := clampLimit(params.Limit, defaultPageSize, maxPageSize)

	items, _, err := client.Repositories.ListBranches(ctx, params.Owner, params.Repository, &gh.BranchListOptions{
		ListOptions: gh.ListOptions{PerPage: limit},
	})
	if err != nil {
		return nil, err
	}

	out := make([]mcpdomain.GitHubBranch, 0, len(items))
	for _, item := range items {
		out = append(out, mcpdomain.GitHubBranch{
			Name: item.GetName(),
			SHA:  item.GetCommit().GetSHA(),
		})
	}
	return out, nil
}

func (c *Client) EnsureBranch(ctx context.Context, params mcpdomain.GitHubEnsureBranchParams) (mcpdomain.GitHubBranch, error) {
	client := c.clientWithToken(params.Token)

	branchName := strings.TrimSpace(params.BranchName)
	if branchName == "" {
		return mcpdomain.GitHubBranch{}, fmt.Errorf("branch name is required")
	}
	refPath := "heads/" + branchName
	baseSHA, err := c.resolveBaseSHA(ctx, client, params)
	if err != nil {
		return mcpdomain.GitHubBranch{}, err
	}

	ref, _, err := client.Git.GetRef(ctx, params.Owner, params.Repository, refPath)
	if err == nil {
		currentSHA := ref.GetObject().GetSHA()
		if currentSHA == baseSHA {
			return mcpdomain.GitHubBranch{Name: branchName, SHA: currentSHA}, nil
		}
		if !params.Force {
			return mcpdomain.GitHubBranch{Name: branchName, SHA: currentSHA}, nil
		}

		updated, _, updateErr := client.Git.UpdateRef(ctx, params.Owner, params.Repository, refPath, gh.UpdateRef{
			SHA:   baseSHA,
			Force: gh.Ptr(true),
		})
		if updateErr != nil {
			return mcpdomain.GitHubBranch{}, updateErr
		}
		return mcpdomain.GitHubBranch{Name: branchName, SHA: updated.GetObject().GetSHA()}, nil
	}
	if !isNotFound(err) {
		return mcpdomain.GitHubBranch{}, err
	}

	created, _, err := client.Git.CreateRef(ctx, params.Owner, params.Repository, gh.CreateRef{
		Ref: "refs/" + refPath,
		SHA: baseSHA,
	})
	if err != nil {
		return mcpdomain.GitHubBranch{}, err
	}
	return mcpdomain.GitHubBranch{
		Name: branchName,
		SHA:  created.GetObject().GetSHA(),
	}, nil
}

func (c *Client) UpsertPullRequest(ctx context.Context, params mcpdomain.GitHubUpsertPullRequestParams) (mcpdomain.GitHubPullRequest, error) {
	client := c.clientWithToken(params.Token)

	title := strings.TrimSpace(params.Title)
	head := strings.TrimSpace(params.HeadBranch)
	base := strings.TrimSpace(params.BaseBranch)
	if title == "" {
		return mcpdomain.GitHubPullRequest{}, fmt.Errorf("pull request title is required")
	}
	if head == "" {
		return mcpdomain.GitHubPullRequest{}, fmt.Errorf("pull request head branch is required")
	}
	if base == "" {
		base = "main"
	}

	if params.PullRequestNumber > 0 {
		req := &gh.PullRequest{
			Title: gh.Ptr(title),
			Body:  gh.Ptr(params.Body),
			Base:  &gh.PullRequestBranch{Ref: gh.Ptr(base)},
		}
		pr, _, err := client.PullRequests.Edit(ctx, params.Owner, params.Repository, params.PullRequestNumber, req)
		if err != nil {
			return mcpdomain.GitHubPullRequest{}, err
		}
		return toPullRequest(pr), nil
	}

	pr, _, err := client.PullRequests.Create(ctx, params.Owner, params.Repository, &gh.NewPullRequest{
		Title: gh.Ptr(title),
		Body:  gh.Ptr(params.Body),
		Head:  gh.Ptr(head),
		Base:  gh.Ptr(base),
		Draft: gh.Ptr(params.Draft),
	})
	if err != nil {
		return mcpdomain.GitHubPullRequest{}, err
	}
	return toPullRequest(pr), nil
}

func (c *Client) CreateIssueComment(ctx context.Context, params mcpdomain.GitHubCreateIssueCommentParams) (mcpdomain.GitHubIssueComment, error) {
	return c.mutateIssueComment(ctx, newCreateIssueCommentMutationParams(params))
}

func (c *Client) EditIssueComment(ctx context.Context, params mcpdomain.GitHubEditIssueCommentParams) (mcpdomain.GitHubIssueComment, error) {
	return c.mutateIssueComment(ctx, newEditIssueCommentMutationParams(params))
}

type issueCommentMutationParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
	CommentID   int64
	Body        string
}

func newCreateIssueCommentMutationParams(params mcpdomain.GitHubCreateIssueCommentParams) issueCommentMutationParams {
	return issueCommentMutationParams{
		Token:       params.Token,
		Owner:       params.Owner,
		Repository:  params.Repository,
		IssueNumber: params.IssueNumber,
		Body:        params.Body,
	}
}

func newEditIssueCommentMutationParams(params mcpdomain.GitHubEditIssueCommentParams) issueCommentMutationParams {
	return issueCommentMutationParams{
		Token:      params.Token,
		Owner:      params.Owner,
		Repository: params.Repository,
		CommentID:  params.CommentID,
		Body:       params.Body,
	}
}

func (c *Client) mutateIssueComment(ctx context.Context, params issueCommentMutationParams) (mcpdomain.GitHubIssueComment, error) {
	client := c.clientWithToken(params.Token)
	request := &gh.IssueComment{Body: gh.Ptr(params.Body)}
	var (
		comment *gh.IssueComment
		err     error
	)
	if params.CommentID > 0 {
		comment, _, err = client.Issues.EditComment(ctx, params.Owner, params.Repository, params.CommentID, request)
	} else {
		comment, _, err = client.Issues.CreateComment(ctx, params.Owner, params.Repository, params.IssueNumber, request)
	}
	if err != nil {
		return mcpdomain.GitHubIssueComment{}, err
	}
	return toIssueComment(comment), nil
}

func (c *Client) AddLabels(ctx context.Context, params mcpdomain.GitHubMutateLabelsParams) ([]mcpdomain.GitHubLabel, error) {
	client := c.clientWithToken(params.Token)
	items, _, err := client.Issues.AddLabelsToIssue(ctx, params.Owner, params.Repository, params.IssueNumber, params.Labels)
	if err != nil {
		return nil, err
	}

	out := make([]mcpdomain.GitHubLabel, 0, len(items))
	for _, item := range items {
		out = append(out, mcpdomain.GitHubLabel{Name: item.GetName()})
	}
	return out, nil
}

func (c *Client) RemoveLabels(ctx context.Context, params mcpdomain.GitHubMutateLabelsParams) ([]mcpdomain.GitHubLabel, error) {
	client := c.clientWithToken(params.Token)
	for _, label := range params.Labels {
		_, err := client.Issues.RemoveLabelForIssue(ctx, params.Owner, params.Repository, params.IssueNumber, label)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
	}
	return c.ListIssueLabels(ctx, mcpdomain.GitHubListIssueLabelsParams{
		Token:       params.Token,
		Owner:       params.Owner,
		Repository:  params.Repository,
		IssueNumber: params.IssueNumber,
	})
}

func (c *Client) resolveBaseSHA(ctx context.Context, client *gh.Client, params mcpdomain.GitHubEnsureBranchParams) (string, error) {
	if sha := strings.TrimSpace(params.BaseSHA); sha != "" {
		return sha, nil
	}

	baseBranch := strings.TrimSpace(params.BaseBranch)
	if baseBranch == "" {
		baseBranch = "main"
	}

	base, _, err := client.Repositories.GetBranch(ctx, params.Owner, params.Repository, baseBranch, 0)
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(base.GetCommit().GetSHA())
	if sha == "" {
		return "", fmt.Errorf("base branch %q has empty SHA", baseBranch)
	}
	return sha, nil
}

func (c *Client) clientWithToken(token string) *gh.Client {
	return gh.NewClient(c.httpClient).WithAuthToken(strings.TrimSpace(token))
}

func isNotFound(err error) bool {
	var apiErr *gh.ErrorResponse
	return errors.As(err, &apiErr) && apiErr.Response != nil && apiErr.Response.StatusCode == http.StatusNotFound
}

func clampLimit(value int, def int, max int) int {
	if value <= 0 {
		return def
	}
	if value > max {
		return max
	}
	return value
}

func toPullRequest(item *gh.PullRequest) mcpdomain.GitHubPullRequest {
	return mcpdomain.GitHubPullRequest{
		Number: item.GetNumber(),
		Title:  item.GetTitle(),
		State:  item.GetState(),
		URL:    item.GetHTMLURL(),
		Head:   item.GetHead().GetRef(),
		Base:   item.GetBase().GetRef(),
	}
}

func toIssueComment(item *gh.IssueComment) mcpdomain.GitHubIssueComment {
	if item == nil {
		return mcpdomain.GitHubIssueComment{}
	}
	return mcpdomain.GitHubIssueComment{
		ID:   item.GetID(),
		Body: item.GetBody(),
		URL:  item.GetHTMLURL(),
		User: item.GetUser().GetLogin(),
	}
}
