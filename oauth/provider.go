package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// provider is the internal contract every OAuth2 backend must satisfy.
// Each supported provider (Google, GitHub, GitLab) implements this interface.
type provider interface {
	// providerName returns the constant that identifies this provider.
	providerName() ProviderName
	// config returns the underlying golang.org/x/oauth2 configuration.
	config() *oauth2.Config
	// userInfo exchanges an OAuth2 token for the normalized user profile.
	userInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}
