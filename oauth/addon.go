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
			{Key: "OAUTH_GOOGLE_CLIENT_ID", Description: "Google OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GOOGLE_CLIENT_SECRET", Description: "Google OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_GITHUB_CLIENT_ID", Description: "GitHub OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GITHUB_CLIENT_SECRET", Description: "GitHub OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_GITLAB_CLIENT_ID", Description: "GitLab OAuth2 client ID", Required: false, Secret: false, Source: "oauth"},
			{Key: "OAUTH_GITLAB_CLIENT_SECRET", Description: "GitLab OAuth2 client secret", Required: false, Secret: true, Source: "oauth"},
			{Key: "OAUTH_REDIRECT_BASE_URL", Description: "Base URL for OAuth2 callback redirects", Required: true, Secret: false, Source: "oauth"},
			{Key: "OAUTH_REDIRECT_ON_SUCCESS", Description: "Frontend URL to redirect after successful login (if set, returns JWT as query param instead of JSON)", Required: false, Secret: false, Source: "oauth"},
		},
	}
}

// RegisterWithPanel registers this addon with a contracts.PanelRegistry.
func (o *OAuth) RegisterWithPanel(r contracts.PanelRegistry) {
	r.RegisterAddon(o)
}
