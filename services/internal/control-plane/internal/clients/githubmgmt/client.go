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
		Status:    "running",
		Checks:    make([]valuetypes.GitHubPreflightCheck, 0, 16),
		Artifacts: make([]valuetypes.GitHubPreflightArtifact, 0, 8),
	}

	stepTimeout := 25 * time.Second
	withTimeout := func(timeout time.Duration, fn func(stepCtx context.Context) error) error {
		if timeout <= 0 {
			timeout = stepTimeout
		}
		stepCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return fn(stepCtx)
	}

	now := time.Now().UTC()
	suffix := now.Format("20060102-150405")
	labelName := "codex-k8s-preflight-" + suffix
	branchName := "codex-k8s-preflight/" + suffix

	platformClient := c.clientWithToken(platformToken)
	botClient := c.clientWithToken(botToken)

	failed := false

	// 1) Platform token: repo access.
	if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
		_, _, err := platformClient.Repositories.Get(stepCtx, owner, repo)
		return err
	}); err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:repo_get", Status: "failed", Details: err.Error()})
	} else {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:repo_get", Status: "ok"})
	}

	// 2) Platform token: create+delete webhook (best-effort cleanup).
	webhookID := int64(0)
	webhookURL := strings.TrimSpace(params.WebhookURL)
	webhookSecret := strings.TrimSpace(params.WebhookSecret)
	hook := (*gh.Hook)(nil)
	err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
		res, _, err := platformClient.Repositories.CreateHook(stepCtx, owner, repo, &gh.Hook{
			Active: gh.Ptr(true),
			Events: []string{"push"},
			Config: &gh.HookConfig{
				URL:         gh.Ptr(webhookURL),
				ContentType: gh.Ptr("json"),
				Secret:      gh.Ptr(webhookSecret),
			},
		})
		hook = res
		return err
	})
	if err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_create", Status: "failed", Details: err.Error()})
	} else {
		webhookID = hook.GetID()
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_create", Status: "ok"})
	}
	if webhookID != 0 {
		artifactIdx := len(report.Artifacts)
		report.Artifacts = append(report.Artifacts, valuetypes.GitHubPreflightArtifact{
			Kind:          "webhook",
			ID:            fmt.Sprintf("%d", webhookID),
			Name:          webhookURL,
			CleanupStatus: "skipped",
		})

		if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
			_, err := platformClient.Repositories.DeleteHook(stepCtx, owner, repo, webhookID)
			return err
		}); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_delete", Status: "failed", Details: err.Error()})
			report.Artifacts[artifactIdx].CleanupStatus = "failed"
			report.Artifacts[artifactIdx].CleanupDetails = err.Error()
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:webhook_delete", Status: "ok"})
			report.Artifacts[artifactIdx].CleanupStatus = "ok"
		}
	}

	// 3) Platform token: create+delete label (best-effort cleanup).
	createdLabel := false
	if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
		_, _, err := platformClient.Issues.CreateLabel(stepCtx, owner, repo, &gh.Label{
			Name:        gh.Ptr(labelName),
			Color:       gh.Ptr("1f6feb"),
			Description: gh.Ptr("codex-k8s preflight label"),
		})
		return err
	}); err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_create", Status: "failed", Details: err.Error()})
	} else {
		createdLabel = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_create", Status: "ok"})
	}
	if createdLabel {
		artifactIdx := len(report.Artifacts)
		report.Artifacts = append(report.Artifacts, valuetypes.GitHubPreflightArtifact{
			Kind:          "label",
			Name:          labelName,
			CleanupStatus: "skipped",
		})

		if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
			_, err := platformClient.Issues.DeleteLabel(stepCtx, owner, repo, labelName)
			return err
		}); err != nil {
			// Non-fatal, but record as failed for cleanup.
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_delete", Status: "failed", Details: err.Error()})
			report.Artifacts[artifactIdx].CleanupStatus = "failed"
			report.Artifacts[artifactIdx].CleanupDetails = err.Error()
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:platform:label_delete", Status: "ok"})
			report.Artifacts[artifactIdx].CleanupStatus = "ok"
		}
	}

	// 4) Bot token: issue + comment + close.
	issueTitle := "codex-k8s-preflight issue " + suffix
	issue := (*gh.Issue)(nil)
	err = withTimeout(stepTimeout, func(stepCtx context.Context) error {
		res, _, err := botClient.Issues.Create(stepCtx, owner, repo, &gh.IssueRequest{
			Title: gh.Ptr(issueTitle),
			Body:  gh.Ptr("codex-k8s preflight issue (auto-cleanup)"),
		})
		issue = res
		return err
	})
	if err != nil {
		failed = true
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_create", Status: "failed", Details: err.Error()})
	} else {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_create", Status: "ok"})
		issueNumber := issue.GetNumber()
		artifactIdx := len(report.Artifacts)
		report.Artifacts = append(report.Artifacts, valuetypes.GitHubPreflightArtifact{
			Kind:          "issue",
			ID:            fmt.Sprintf("%d", issueNumber),
			Name:          issueTitle,
			CleanupStatus: "skipped",
		})

		if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
			_, _, err := botClient.Issues.CreateComment(stepCtx, owner, repo, issueNumber, &gh.IssueComment{Body: gh.Ptr("codex-k8s preflight comment")})
			return err
		}); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_comment", Status: "failed", Details: err.Error()})
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_comment", Status: "ok"})
		}
		if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
			_, _, err := botClient.Issues.Edit(stepCtx, owner, repo, issueNumber, &gh.IssueRequest{State: gh.Ptr("closed")})
			return err
		}); err != nil {
			failed = true
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_close", Status: "failed", Details: err.Error()})
			report.Artifacts[artifactIdx].CleanupStatus = "failed"
			report.Artifacts[artifactIdx].CleanupDetails = err.Error()
		} else {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:issue_close", Status: "ok"})
			report.Artifacts[artifactIdx].CleanupStatus = "ok"
		}
	}

	// 5) Bot token: branch + commit + PR + comment + close + delete branch.
	defaultBranch := "main"
	repoInfo := (*gh.Repository)(nil)
	err = withTimeout(stepTimeout, func(stepCtx context.Context) error {
		res, _, err := platformClient.Repositories.Get(stepCtx, owner, repo)
		repoInfo = res
		return err
	})
	if err == nil {
		if b := strings.TrimSpace(repoInfo.GetDefaultBranch()); b != "" {
			defaultBranch = b
		}
	}
	baseRef := "refs/heads/" + defaultBranch
	ref := (*gh.Reference)(nil)
	err = withTimeout(stepTimeout, func(stepCtx context.Context) error {
		res, _, err := botClient.Git.GetRef(stepCtx, owner, repo, baseRef)
		ref = res
		return err
	})
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
			if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
				_, _, err := botClient.Git.CreateRef(stepCtx, owner, repo, gh.CreateRef{
					Ref: newRef,
					SHA: baseSHA,
				})
				return err
			}); err != nil {
				failed = true
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_create", Status: "failed", Details: err.Error()})
			} else {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_create", Status: "ok"})
				branchArtifactIdx := len(report.Artifacts)
				report.Artifacts = append(report.Artifacts, valuetypes.GitHubPreflightArtifact{
					Kind:          "branch",
					Name:          branchName,
					CleanupStatus: "skipped",
				})

				// Create one commit with a trivial file.
				content := "codex-k8s preflight " + suffix + "\n"
				blob := (*gh.Blob)(nil)
				err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
					res, _, err := botClient.Git.CreateBlob(stepCtx, owner, repo, gh.Blob{
						Content:  gh.Ptr(content),
						Encoding: gh.Ptr("utf-8"),
					})
					blob = res
					return err
				})
				if err != nil {
					failed = true
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:blob_create", Status: "failed", Details: err.Error()})
				} else {
					commitObj := (*gh.Commit)(nil)
					err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
						res, _, err := botClient.Git.GetCommit(stepCtx, owner, repo, baseSHA)
						commitObj = res
						return err
					})
					if err != nil {
						failed = true
						report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_get", Status: "failed", Details: err.Error()})
					} else {
						baseTree := commitObj.GetTree().GetSHA()
						tree := (*gh.Tree)(nil)
						err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
							res, _, err := botClient.Git.CreateTree(stepCtx, owner, repo, baseTree, []*gh.TreeEntry{{
								Path: gh.Ptr(".codex-k8s-preflight.txt"),
								Mode: gh.Ptr("100644"),
								Type: gh.Ptr("blob"),
								SHA:  blob.SHA,
							}})
							tree = res
							return err
						})
						if err != nil {
							failed = true
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:tree_create", Status: "failed", Details: err.Error()})
						} else {
							commitMsg := "chore: codex-k8s preflight " + suffix
							newCommit := (*gh.Commit)(nil)
							err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
								res, _, err := botClient.Git.CreateCommit(stepCtx, owner, repo, gh.Commit{
									Message: gh.Ptr(commitMsg),
									Tree:    tree,
									Parents: []*gh.Commit{{SHA: gh.Ptr(baseSHA)}},
								}, nil)
								newCommit = res
								return err
							})
							if err != nil {
								failed = true
								report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_create", Status: "failed", Details: err.Error()})
							} else {
								if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
									_, _, err := botClient.Git.UpdateRef(stepCtx, owner, repo, "heads/"+branchName, gh.UpdateRef{
										SHA:   strings.TrimSpace(newCommit.GetSHA()),
										Force: gh.Ptr(true),
									})
									return err
								}); err != nil {
									failed = true
									report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:ref_update", Status: "failed", Details: err.Error()})
								} else {
									report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:commit_push", Status: "ok"})

									prTitle := "codex-k8s preflight " + suffix
									pr := (*gh.PullRequest)(nil)
									err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
										res, _, err := botClient.PullRequests.Create(stepCtx, owner, repo, &gh.NewPullRequest{
											Title: gh.Ptr(prTitle),
											Head:  gh.Ptr(branchName),
											Base:  gh.Ptr(defaultBranch),
											Body:  gh.Ptr("codex-k8s preflight PR (auto-cleanup)"),
										})
										pr = res
										return err
									})
									if err != nil {
										failed = true
										report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_create", Status: "failed", Details: err.Error()})
									} else {
										report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_create", Status: "ok"})
										prArtifactIdx := len(report.Artifacts)
										report.Artifacts = append(report.Artifacts, valuetypes.GitHubPreflightArtifact{
											Kind:          "pr",
											ID:            fmt.Sprintf("%d", pr.GetNumber()),
											Name:          prTitle,
											CleanupStatus: "skipped",
										})

										if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
											_, _, err := botClient.Issues.CreateComment(stepCtx, owner, repo, pr.GetNumber(), &gh.IssueComment{Body: gh.Ptr("codex-k8s preflight PR comment")})
											return err
										}); err != nil {
											failed = true
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_comment", Status: "failed", Details: err.Error()})
										} else {
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_comment", Status: "ok"})
										}

										if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
											_, _, err := botClient.PullRequests.Edit(stepCtx, owner, repo, pr.GetNumber(), &gh.PullRequest{State: gh.Ptr("closed")})
											return err
										}); err != nil {
											failed = true
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_close", Status: "failed", Details: err.Error()})
											report.Artifacts[prArtifactIdx].CleanupStatus = "failed"
											report.Artifacts[prArtifactIdx].CleanupDetails = err.Error()
										} else {
											report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:pr_close", Status: "ok"})
											report.Artifacts[prArtifactIdx].CleanupStatus = "ok"
										}
									}
								}
							}
						}
					}
				}

				// Cleanup branch.
				if err := withTimeout(stepTimeout, func(stepCtx context.Context) error {
					_, err := botClient.Git.DeleteRef(stepCtx, owner, repo, "heads/"+branchName)
					return err
				}); err != nil {
					failed = true
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_delete", Status: "failed", Details: err.Error()})
					report.Artifacts[branchArtifactIdx].CleanupStatus = "failed"
					report.Artifacts[branchArtifactIdx].CleanupDetails = err.Error()
				} else {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:bot:branch_delete", Status: "ok"})
					report.Artifacts[branchArtifactIdx].CleanupStatus = "ok"
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
