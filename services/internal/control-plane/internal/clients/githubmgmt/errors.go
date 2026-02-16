package githubmgmt

import (
	"errors"
	"net/http"

	gh "github.com/google/go-github/v82/github"
)

func isGitHubStatus(err error, code int) bool {
	if err == nil {
		return false
	}
	var resp *gh.ErrorResponse
	if errors.As(err, &resp) {
		if resp.Response != nil && resp.Response.StatusCode == code {
			return true
		}
	}
	return false
}

func isGitHubNotFound(err error) bool {
	return isGitHubStatus(err, http.StatusNotFound)
}

func isGitHubConflict(err error) bool {
	return isGitHubStatus(err, http.StatusConflict)
}

func isGitHubUnprocessable(err error) bool {
	return isGitHubStatus(err, http.StatusUnprocessableEntity)
}
