package oauth

// UserInfo is the normalized user profile returned by every supported provider.
// It is embedded as claims when signing the JWT on a successful callback.
type UserInfo struct {
	// Provider identifies which OAuth2 provider authenticated the user.
	Provider ProviderName
	// ID is the provider-specific unique identifier for the user.
	ID string
	// Email is the verified primary email address, or empty if not provided.
	Email string
	// Name is the display name returned by the provider.
	Name string
	// AvatarURL is the profile picture URL, or empty if not provided.
	AvatarURL string
}
