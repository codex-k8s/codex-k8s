package githubmgmt

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	gh "github.com/google/go-github/v82/github"
)

func (c *Client) GetDefaultBranch(ctx context.Context, token string, owner string, repo string) (string, error) {
	client := c.clientWithToken(token)
	info, _, err := client.Repositories.Get(ctx, strings.TrimSpace(owner), strings.TrimSpace(repo))
	if err != nil {
		return "", fmt.Errorf("github get repository %s/%s: %w", owner, repo, err)
	}
	branch := strings.TrimSpace(info.GetDefaultBranch())
	if branch == "" {
		branch = "main"
	}
	return branch, nil
}

func (c *Client) GetFile(ctx context.Context, token string, owner string, repo string, filePath string, ref string) ([]byte, bool, error) {
	client := c.clientWithToken(token)
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, false, fmt.Errorf("path is required")
	}
	opt := &gh.RepositoryContentGetOptions{Ref: strings.TrimSpace(ref)}
	content, _, resp, err := client.Repositories.GetContents(ctx, strings.TrimSpace(owner), strings.TrimSpace(repo), filePath, opt)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("github get contents %s/%s %s: %w", owner, repo, filePath, err)
	}
	if content == nil {
		return nil, false, fmt.Errorf("github path %s is not a file", filePath)
	}
	rawContent, err := decodeRepositoryContent(content)
	if err != nil {
		// Large files may come with encoding=none; fall back to download API.
		if errors.Is(err, errGitHubContentNeedsDownload) {
			rc, resp, dlErr := client.Repositories.DownloadContents(ctx, strings.TrimSpace(owner), strings.TrimSpace(repo), filePath, opt)
			if dlErr != nil {
				if resp != nil && resp.StatusCode == 404 {
					return nil, false, nil
				}
				return nil, false, fmt.Errorf("github download contents %s/%s %s: %w", owner, repo, filePath, dlErr)
			}
			if rc == nil {
				return nil, false, fmt.Errorf("github download contents %s: empty body", filePath)
			}
			body, readErr := io.ReadAll(rc)
			closeErr := rc.Close()
			if readErr != nil {
				return nil, false, fmt.Errorf("read github download %s: %w", filePath, readErr)
			}
			if closeErr != nil {
				return nil, false, fmt.Errorf("close github download %s: %w", filePath, closeErr)
			}
			if strings.TrimSpace(string(body)) == "" {
				return []byte{}, true, nil
			}
			return body, true, nil
		}
		return nil, false, fmt.Errorf("read github content %s: %w", filePath, err)
	}
	if strings.TrimSpace(rawContent) == "" {
		return []byte{}, true, nil
	}
	return []byte(rawContent), true, nil
}

var errGitHubContentNeedsDownload = errors.New("github content needs download")

func decodeRepositoryContent(content *gh.RepositoryContent) (string, error) {
	if content == nil {
		return "", fmt.Errorf("content is required")
	}

	encoding := strings.TrimSpace(content.GetEncoding())
	switch encoding {
	case "base64":
		if content.Content == nil {
			return "", fmt.Errorf("malformed response: base64 encoding of null content")
		}
		// GitHub returns base64 with embedded newlines; go-github uses DecodeString which fails on them.
		cleaned := strings.NewReplacer("\n", "", "\r", "").Replace(*content.Content)
		decoded, err := base64.StdEncoding.DecodeString(cleaned)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	case "":
		if content.Content == nil {
			return "", nil
		}
		return *content.Content, nil
	case "none":
		return "", errGitHubContentNeedsDownload
	default:
		return "", fmt.Errorf("unsupported content encoding: %s", encoding)
	}
}

func (c *Client) CreatePullRequestWithFiles(ctx context.Context, token string, owner string, repo string, baseBranch string, headBranch string, title string, body string, files map[string][]byte) (prNumber int, prURL string, err error) {
	client := c.clientWithToken(token)
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return 0, "", fmt.Errorf("owner and repository are required")
	}
	baseBranch = strings.TrimSpace(baseBranch)
	if baseBranch == "" {
		baseBranch = "main"
	}
	headBranch = strings.TrimSpace(headBranch)
	if headBranch == "" {
		return 0, "", fmt.Errorf("head branch is required")
	}
	if len(files) == 0 {
		return 0, "", fmt.Errorf("files are required")
	}

	refPath := "refs/heads/" + baseBranch
	ref, _, err := client.Git.GetRef(ctx, owner, repo, refPath)
	if err != nil {
		return 0, "", fmt.Errorf("github get base ref %s: %w", refPath, err)
	}
	baseSHA := strings.TrimSpace(ref.GetObject().GetSHA())
	if baseSHA == "" {
		return 0, "", fmt.Errorf("github base ref %s has empty sha", refPath)
	}

	headRef := "refs/heads/" + headBranch
	if _, _, err := client.Git.GetRef(ctx, owner, repo, headRef); err != nil {
		_, _, createErr := client.Git.CreateRef(ctx, owner, repo, gh.CreateRef{Ref: headRef, SHA: baseSHA})
		if createErr != nil {
			return 0, "", fmt.Errorf("github create head ref %s: %w", headRef, createErr)
		}
	}

	baseCommit, _, err := client.Git.GetCommit(ctx, owner, repo, baseSHA)
	if err != nil {
		return 0, "", fmt.Errorf("github get base commit: %w", err)
	}
	baseTreeSHA := strings.TrimSpace(baseCommit.GetTree().GetSHA())
	if baseTreeSHA == "" {
		return 0, "", fmt.Errorf("github base tree sha is empty")
	}

	entries := make([]*gh.TreeEntry, 0, len(files))
	for filePath, data := range files {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return 0, "", fmt.Errorf("file path is empty")
		}
		blob, _, err := client.Git.CreateBlob(ctx, owner, repo, gh.Blob{
			Content:  gh.Ptr(string(data)),
			Encoding: gh.Ptr("utf-8"),
		})
		if err != nil {
			return 0, "", fmt.Errorf("github create blob %s: %w", filePath, err)
		}
		entries = append(entries, &gh.TreeEntry{
			Path: gh.Ptr(filePath),
			Mode: gh.Ptr("100644"),
			Type: gh.Ptr("blob"),
			SHA:  blob.SHA,
		})
	}

	tree, _, err := client.Git.CreateTree(ctx, owner, repo, baseTreeSHA, entries)
	if err != nil {
		return 0, "", fmt.Errorf("github create tree: %w", err)
	}
	commitMsg := strings.TrimSpace(title)
	if commitMsg == "" {
		commitMsg = "chore: docset sync"
	}
	newCommit, _, err := client.Git.CreateCommit(ctx, owner, repo, gh.Commit{
		Message: gh.Ptr(commitMsg),
		Tree:    tree,
		Parents: []*gh.Commit{{SHA: gh.Ptr(baseSHA)}},
	}, nil)
	if err != nil {
		return 0, "", fmt.Errorf("github create commit: %w", err)
	}
	newSHA := strings.TrimSpace(newCommit.GetSHA())
	if newSHA == "" {
		return 0, "", fmt.Errorf("github created commit sha is empty")
	}

	if _, _, err := client.Git.UpdateRef(ctx, owner, repo, "heads/"+headBranch, gh.UpdateRef{
		SHA:   newSHA,
		Force: gh.Ptr(true),
	}); err != nil {
		return 0, "", fmt.Errorf("github update ref heads/%s: %w", headBranch, err)
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &gh.NewPullRequest{
		Title: gh.Ptr(strings.TrimSpace(title)),
		Head:  gh.Ptr(headBranch),
		Base:  gh.Ptr(baseBranch),
		Body:  gh.Ptr(body),
	})
	if err != nil {
		return 0, "", fmt.Errorf("github create pr: %w", err)
	}
	return pr.GetNumber(), strings.TrimSpace(pr.GetHTMLURL()), nil
}
