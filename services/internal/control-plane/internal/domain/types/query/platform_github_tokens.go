package query

// PlatformGitHubTokensUpsertParams defines encrypted token values for singleton upsert.
type PlatformGitHubTokensUpsertParams struct {
	PlatformTokenEncrypted []byte
	BotTokenEncrypted      []byte
}
