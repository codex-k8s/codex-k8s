package query

// StaffResolveByEmailParams defines inputs for staff principal resolution by allowed email.
type StaffResolveByEmailParams struct {
	Email       string
	GitHubLogin string
}

// StaffAuthorizeOAuthUserParams defines inputs for OAuth identity authorization flow.
type StaffAuthorizeOAuthUserParams struct {
	Email        string
	GitHubUserID int64
	GitHubLogin  string
}
