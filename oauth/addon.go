package oauth

import "github.com/slice-soft/ss-keel-core/contracts"

// Compile-time assertions: OAuth must satisfy all three contracts.
var (
	_ contracts.Addon        = (*OAuth)(nil)
	_ contracts.Debuggable   = (*OAuth)(nil)
	_ contracts.Manifestable = (*OAuth)(nil)
)

// ID implements contracts.Addon.
func (o *OAuth) ID() string { return "oauth" }

// PanelID implements contracts.Debuggable.
func (o *OAuth) PanelID() string { return "oauth" }

// PanelLabel implements contracts.Debuggable.
func (o *OAuth) PanelLabel() string { return "Auth (OAuth2)" }

// PanelEvents implements contracts.Debuggable.
func (o *OAuth) PanelEvents() <-chan contracts.PanelEvent { return o.events }

// Manifest implements contracts.Manifestable.
func (o *OAuth) Manifest() contracts.AddonManifest {
	return contracts.AddonManifest{
		ID:           "oauth",
		Version:      "1.0.0",
		Capabilities: []string{"auth"},
		Resources:    []string{},
		EnvVars: []contracts.EnvVar{
			{Key: "OAUTH_GOOGLE_CLIENT_ID", ConfigKey: "oauth.google.client-id", Description: "Google OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GOOGLE_CLIENT_SECRET", ConfigKey: "oauth.google.client-secret", Description: "Google OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_GITHUB_CLIENT_ID", ConfigKey: "oauth.github.client-id", Description: "GitHub OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GITHUB_CLIENT_SECRET", ConfigKey: "oauth.github.client-secret", Description: "GitHub OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_GITLAB_CLIENT_ID", ConfigKey: "oauth.gitlab.client-id", Description: "GitLab OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GITLAB_CLIENT_SECRET", ConfigKey: "oauth.gitlab.client-secret", Description: "GitLab OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_REDIRECT_BASE_URL", ConfigKey: "oauth.redirect-base-url", Description: "Base URL for OAuth2 callback redirects", Required: false, Secret: false, Default: "http://127.0.0.1:7331", Source: "oauth"},
			{Key: "OAUTH_ROUTE_PREFIX", ConfigKey: "oauth.route-prefix", Description: "Route prefix for OAuth login and callback routes", Required: false, Secret: false, Default: "/auth", Source: "oauth"},
			{Key: "OAUTH_ENABLED_PROVIDERS", ConfigKey: "oauth.enabled-providers", Description: "Optional comma-separated list of providers to enable", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_REDIRECT_ON_SUCCESS", ConfigKey: "oauth.redirect-on-success", Description: "Frontend URL to redirect after successful login (if set, returns JWT as query param instead of JSON)", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_REDIRECT_TOKEN_PARAM", ConfigKey: "oauth.redirect-token-param", Description: "Query parameter name used for redirect mode", Required: false, Secret: false, Default: "token", Source: "oauth"},
		},
	}
}

// RegisterWithPanel registers this addon with a contracts.PanelRegistry.
func (o *OAuth) RegisterWithPanel(r contracts.PanelRegistry) {
	r.RegisterAddon(o)
}
