package oauth

import "github.com/slice-soft/ss-keel-core/contracts"

// ProviderName identifies an OAuth2 provider.
type ProviderName string

const (
	ProviderGoogle ProviderName = "google"
	ProviderGitHub ProviderName = "github"
	ProviderGitLab ProviderName = "gitlab"
)

// ProviderConfig holds the OAuth2 credentials for a single provider.
type ProviderConfig struct {
	// ClientID is the OAuth2 application client ID.
	ClientID string
	// ClientSecret is the OAuth2 application client secret.
	ClientSecret string
	// RedirectURL is the callback URL registered in the provider's console.
	// Example: "https://myapp.com/auth/google/callback"
	RedirectURL string
	// Scopes overrides the default scopes for this provider.
	// Leave nil to use provider defaults.
	Scopes []string
}

// Config is the top-level configuration for the OAuth addon.
type Config struct {
	// Google, GitHub, GitLab — configure only the providers you need.
	// A provider is skipped when its config is nil or incomplete.
	Google *ProviderConfig
	GitHub *ProviderConfig
	GitLab *ProviderConfig

	// Signer signs the JWT returned to the client after a successful OAuth flow.
	// Typically satisfied by ss-keel-jwt. Required — New panics if nil.
	Signer contracts.TokenSigner

	// Logger is optional. Warn-level output is suppressed when nil.
	Logger contracts.Logger

	// RedirectOnSuccess, when non-empty, causes the callback handler to redirect
	// the browser to this URL with the signed JWT appended as a query parameter.
	// When empty the handler returns JSON: {"token": "<jwt>"}.
	// Example: "https://myapp.com/auth/done"
	RedirectOnSuccess string

	// RedirectTokenParam is the query parameter name used in the redirect URL
	// when RedirectOnSuccess is set. Defaults to "token".
	// Example: set to "access_token" → https://myapp.com/auth/done?access_token=<jwt>
	RedirectTokenParam string
}
