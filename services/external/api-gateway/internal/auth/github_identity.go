package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type gitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func fetchGitHubIdentity(ctx context.Context, token *oauth2.Token) (gitHubUser, string, error) {
	if token == nil || !token.Valid() {
		return gitHubUser{}, "", errors.New("oauth token is missing or invalid")
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	user, err := fetchGitHubUser(ctx, client)
	if err != nil {
		return gitHubUser{}, "", err
	}

	email, err := fetchGitHubPrimaryVerifiedEmail(ctx, client)
	if err != nil {
		return gitHubUser{}, "", err
	}

	return user, email, nil
}

func fetchGitHubUser(ctx context.Context, client *http.Client) (gitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return gitHubUser{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return gitHubUser{}, fmt.Errorf("call github user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return gitHubUser{}, fmt.Errorf("github user returned %s", resp.Status)
	}

	var u gitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return gitHubUser{}, fmt.Errorf("decode github user: %w", err)
	}
	if u.ID == 0 || u.Login == "" {
		return gitHubUser{}, errors.New("github user response missing id/login")
	}
	return u, nil
}

func fetchGitHubPrimaryVerifiedEmail(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call github emails: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github emails returned %s", resp.Status)
	}

	var emails []gitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode github emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified && e.Email != "" {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified && e.Email != "" {
			return e.Email, nil
		}
	}

	return "", nil
}
