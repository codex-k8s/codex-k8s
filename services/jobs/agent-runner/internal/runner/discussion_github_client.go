package runner

import (
	"fmt"
	"strings"

	gh "github.com/google/go-github/v82/github"
)

func (s *Service) newDiscussionGitHubClient() (*gh.Client, string, string, error) {
	repositoryOwner, repositoryName, err := splitRepositoryFullName(s.cfg.RepositoryFullName)
	if err != nil {
		return nil, "", "", err
	}

	token := strings.TrimSpace(s.cfg.GitBotToken)
	if token == "" {
		return nil, "", "", fmt.Errorf("github token is required")
	}

	return gh.NewClient(nil).WithAuthToken(token), repositoryOwner, repositoryName, nil
}

func splitRepositoryFullName(value string) (string, string, error) {
	trimmedValue := strings.TrimSpace(value)
	parts := strings.Split(trimmedValue, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("repository full name must have owner/repo format")
	}

	owner := strings.TrimSpace(parts[0])
	repository := strings.TrimSpace(parts[1])
	if owner == "" || repository == "" {
		return "", "", fmt.Errorf("repository full name must have owner/repo format")
	}

	return owner, repository, nil
}
