package entity

// User is a persisted staff user record.
type User struct {
	ID              string
	Email           string
	GitHubUserID    int64
	GitHubLogin     string
	IsPlatformAdmin bool
	// IsPlatformOwner marks the single platform owner (highest privileges).
	IsPlatformOwner bool
}
