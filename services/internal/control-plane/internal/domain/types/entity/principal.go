package entity

// Principal is an authenticated staff identity.
type Principal struct {
	UserID          string
	Email           string
	GitHubLogin     string
	IsPlatformAdmin bool
	IsPlatformOwner bool
}
