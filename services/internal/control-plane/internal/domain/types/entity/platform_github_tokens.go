package entity

// PlatformGitHubTokens stores encrypted platform and bot GitHub tokens.
type PlatformGitHubTokens struct {
	PlatformTokenEncrypted []byte
	BotTokenEncrypted      []byte
}
