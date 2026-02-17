package entity

// ProjectGitHubTokens stores project-scoped GitHub credentials overrides (encrypted at rest).
//
// This record is used for credentials fallback chain: repo -> project -> platform.
type ProjectGitHubTokens struct {
	ProjectID string

	HasPlatformToken bool
	HasBotToken      bool

	BotUsername string
	BotEmail    string
}

