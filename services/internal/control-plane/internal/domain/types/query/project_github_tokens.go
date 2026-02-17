package query

// ProjectGitHubTokensUpsertParams defines inputs for project github tokens upsert.
type ProjectGitHubTokensUpsertParams struct {
	ProjectID string

	PlatformTokenEncrypted []byte
	BotTokenEncrypted      []byte

	BotUsername string
	BotEmail    string
}

