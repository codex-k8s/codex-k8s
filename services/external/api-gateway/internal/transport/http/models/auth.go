package models

type MeUser struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	GitHubLogin     string `json:"github_login"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
	IsPlatformOwner bool   `json:"is_platform_owner"`
}

type MeResponse struct {
	User MeUser `json:"user"`
}
