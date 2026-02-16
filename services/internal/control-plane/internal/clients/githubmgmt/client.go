package githubmgmt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	gh "github.com/google/go-github/v82/github"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// Client provides GitHub management operations used by staff/control-plane (preflight, repo mutations).
type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) Preflight(ctx context.Context, params valuetypes.GitHubPreflightParams) (valuetypes.GitHubPreflightReport, error) {
	owner := strings.TrimSpace(params.Owner)
	repo := strings.TrimSpace(params.Repository)
	if owner == "" || repo == "" {
		return valuetypes.GitHubPreflightReport{}, fmt.Errorf("owner and repository are required")
	}
	platformToken := strings.TrimSpace(params.PlatformToken)
	botToken := strings.TrimSpace(params.BotToken)
	if platformToken == "" {
		return valuetypes.GitHubPreflightReport{}, fmt.Errorf("platform token is required")
	}
	if botToken == "" {
		return valuetypes.GitHubPreflightReport{}, fmt.Errorf("bot token is required")
	}

	report := valuetypes.GitHubPreflightReport{
		Status: "running",
		Checks: make([]valuetypes.GitHubPreflightCheck, 0, 12),
	}

	now := time.Now().UTC()
	suffix := now.Format("20060102-150405")
	labelName := "codex-k8s-preflight-" + suffix
	branchName := "codex-k8s-preflight/" + suffix

	platformClient := c.clientWithToken(platformToken)
	botClient := c.clientWithToken(botToken)

	failed := false

	// 1) Platform token: repo access.
	if _, _, err := platformClient.Repositories.Get(ctx, owner, repo); err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:repo_get", Status: "failed", Details: err.Error()})
	} else {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:repo_get", Status: "ok"})
	}

	// 2) Platform token: create+delete webhook (best-effort cleanup).
	webhookID := int64(0)
	hook, _, err := platformClient.Repositories.CreateHook(ctx, owner, repo, &gh.Hook{
		Active: gh.Ptr(true),
		Events: []string{"push"},
		Config: &gh.HookConfig{
			URL:         gh.Ptr(strings.TrimSpace(params.WebhookURL)),
			ContentType: gh.Ptr("json"),
			Secret:      gh.Ptr(strings.TrimSpace(params.WebhookSecret)),
		},
	})
	if err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_create", Status: "failed", Details: err.Error()})
	} else {
		webhookID = hook.GetID()
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_create", Status: "ok"})
	}
	if webhookID != 0 {
		if _, err := platformClient.Repositories.DeleteHook(ctx, owner, repo, webhookID); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_delete", Status: "failed", Details: err.Error()})
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_delete", Status: "ok"})
		}
	}

	// 3) Platform token: create+delete label (best-effort cleanup).
	createdLabel := false
	if _, _, err := platformClient.Issues.CreateLabel(ctx, owner, repo, &gh.Label{
		Name:        gh.Ptr(labelName),
		Color:       gh.Ptr("1f6feb"),
		Description: gh.Ptr("codex-k8s preflight label"),
	}); err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_create", Status: "failed", Details: err.Error()})
	} else {
		createdLabel = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_create", Status: "ok"})
	}
	if createdLabel {
		if _, err := platformClient.Issues.DeleteLabel(ctx, owner, repo, labelName); err != nil {
			// Non-fatal, but record as failed for cleanup.
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_delete", Status: "failed", Details: err.Error()})
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_delete", Status: "ok"})
		}
	}

	// 4) Bot token: issue + comment + close.
	issueTitle := "codex-k8s-preflight issue " + suffix
	issue, _, err := botClient.Issues.Create(ctx, owner, repo, &gh.IssueRequest{
		Title: gh.Ptr(issueTitle),
		Body:  gh.Ptr("codex-k8s preflight issue (auto-cleanup)"),
	})
	if err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_create", Status: "failed", Details: err.Error()})
	} else {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_create", Status: "ok"})
		issueNumber := issue.GetNumber()
		if _, _, err := botClient.Issues.CreateComment(ctx, owner, repo, issueNumber, &gh.IssueComment{Body: gh.Ptr("codex-k8s preflight comment")}); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_comment", Status: "failed", Details: err.Error()})
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_comment", Status: "ok"})
		}
		if _, _, err := botClient.Issues.Edit(ctx, owner, repo, issueNumber, &gh.IssueRequest{State: gh.Ptr("closed")}); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_close", Status: "failed", Details: err.Error()})
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_close", Status: "ok"})
		}
	}

	// 5) Bot token: branch + commit + PR + comment + close + delete branch.
	defaultBranch := "main"
	repoInfo, _, err := platformClient.Repositories.Get(ctx, owner, repo)
	if err == nil {
		if b := strings.TrimSpace(repoInfo.GetDefaultBranch()); b != "" {
			defaultBranch = b
		}
	}
	baseRef := "refs/heads/" + defaultBranch
	ref, _, err := botClient.Git.GetRef(ctx, owner, repo, baseRef)
	if err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:git_get_ref", Status: "failed", Details: err.Error()})
	} else {
		baseSHA := strings.TrimSpace(ref.GetObject().GetSHA())
		if baseSHA == "" {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:git_get_ref", Status: "failed", Details: "empty base sha"})
		} else {
			newRef := "refs/heads/" + branchName
			if _, _, err := botClient.Git.CreateRef(ctx, owner, repo, gh.CreateRef{
				Ref: newRef,
				SHA: baseSHA,
			}); err != nil {
				failed = true
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_create", Status: "failed", Details: err.Error()})
			} else {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_create", Status: "ok"})

				// Create one commit with a trivial file.
				content := "codex-k8s preflight " + suffix + "\n"
				blob, _, err := botClient.Git.CreateBlob(ctx, owner, repo, gh.Blob{
					Content:  gh.Ptr(content),
					Encoding: gh.Ptr("utf-8"),
				})
				if err != nil {
					failed = true
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:blob_create", Status: "failed", Details: err.Error()})
				} else {
					commitObj, _, err := botClient.Git.GetCommit(ctx, owner, repo, baseSHA)
					if err != nil {
						failed = true
						report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_get", Status: "failed", Details: err.Error()})
					} else {
						baseTree := commitObj.GetTree().GetSHA()
						tree, _, err := botClient.Git.CreateTree(ctx, owner, repo, baseTree, []*gh.TreeEntry{{
							Path: gh.Ptr(".codex-k8s-preflight.txt"),
							Mode: gh.Ptr("100644"),
							Type: gh.Ptr("blob"),
							SHA:  blob.SHA,
						}})
						if err != nil {
							failed = true
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:tree_create", Status: "failed", Details: err.Error()})
						} else {
							commitMsg := "chore: codex-k8s preflight " + suffix
							newCommit, _, err := botClient.Git.CreateCommit(ctx, owner, repo, gh.Commit{
								Message: gh.Ptr(commitMsg),
								Tree:    tree,
								Parents: []*gh.Commit{{SHA: gh.Ptr(baseSHA)}},
							}, nil)
							if err != nil {
								failed = true
								report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_create", Status: "failed", Details: err.Error()})
							} else {
								if _, _, err := botClient.Git.UpdateRef(ctx, owner, repo, "heads/"+branchName, gh.UpdateRef{
									SHA:   strings.TrimSpace(newCommit.GetSHA()),
									Force: gh.Ptr(true),
								}); err != nil {
									failed = true
									report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:ref_update", Status: "failed", Details: err.Error()})
								} else {
									report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_push", Status: "ok"})

									prTitle := "codex-k8s preflight " + suffix
									pr, _, err := botClient.PullRequests.Create(ctx, owner, repo, &gh.NewPullRequest{
										Title: gh.Ptr(prTitle),
										Head:  gh.Ptr(branchName),
										Base:  gh.Ptr(defaultBranch),
										Body:  gh.Ptr("codex-k8s preflight PR (auto-cleanup)"),
									})
									if err != nil {
										failed = true
										report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_create", Status: "failed", Details: err.Error()})
									} else {
										report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_create", Status: "ok"})

										if _, _, err := botClient.Issues.CreateComment(ctx, owner, repo, pr.GetNumber(), &gh.IssueComment{Body: gh.Ptr("codex-k8s preflight PR comment")}); err != nil {
											failed = true
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_comment", Status: "failed", Details: err.Error()})
										} else {
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_comment", Status: "ok"})
										}

										if _, _, err := botClient.PullRequests.Edit(ctx, owner, repo, pr.GetNumber(), &gh.PullRequest{State: gh.Ptr("closed")}); err != nil {
											failed = true
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_close", Status: "failed", Details: err.Error()})
										} else {
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_close", Status: "ok"})
										}
									}
								}
							}
						}
					}
				}

				// Cleanup branch.
				if _, err := botClient.Git.DeleteRef(ctx, owner, repo, "heads/"+branchName); err != nil {
					failed = true
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_delete", Status: "failed", Details: err.Error()})
				} else {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_delete", Status: "ok"})
				}
			}
		}
	}

	report.FinishedAt = time.Now().UTC()
	if failed {
		report.Status = "failed"
	} else {
		report.Status = "ok"
	}
	return report, nil
}

func (c *Client) clientWithToken(token string) *gh.Client {
	return gh.NewClient(c.httpClient).WithAuthToken(strings.TrimSpace(token))
}
