package oauth

import (
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core/httpx"
)

// Controller implements contracts.Controller[httpx.Route] and registers
// OAuth login and callback routes for every configured provider.
//
// Default routes (with prefix "/auth"):
//
//	GET /auth/google           → redirect to Google authorization page
//	GET /auth/google/callback  → exchange code, sign JWT, return token
//	GET /auth/github           → redirect to GitHub authorization page
//	GET /auth/github/callback  → exchange code, sign JWT, return token
//	GET /auth/gitlab           → redirect to GitLab authorization page
//	GET /auth/gitlab/callback  → exchange code, sign JWT, return token
//
// Register in cmd/main.go:
//
//	app.RegisterController(oauth.NewController(oauthManager))
//
// Override the path prefix:
//
//	app.RegisterController(oauth.NewController(oauthManager, "/sign-in"))
type Controller struct {
	mgr    *OAuth
	prefix string
}

var _ contracts.Controller[httpx.Route] = (*Controller)(nil)

// NewController creates a Controller that registers OAuth routes for all providers
// configured in oauthManager. The optional prefix argument overrides the default "/auth".
func NewController(mgr *OAuth, prefix ...string) *Controller {
	p := "/auth"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &Controller{mgr: mgr, prefix: p}
}

// Routes returns the login and callback routes for every configured provider.
func (c *Controller) Routes() []httpx.Route {
	var routes []httpx.Route

	for _, name := range []ProviderName{ProviderGoogle, ProviderGitHub, ProviderGitLab} {
		if _, ok := c.mgr.providers[name]; !ok {
			continue
		}

		base := c.prefix + "/" + string(name)

		routes = append(routes,
			httpx.GET(base, c.mgr.LoginHandler(name)).
				Tag("OAuth").
				Describe("Redirect to "+string(name)+" OAuth2 authorization page"),

			httpx.GET(base+"/callback", c.mgr.CallbackHandler(name)).
				Tag("OAuth").
				WithQueryParam("code", "string", true, "Authorization code returned by the provider").
				Describe("Exchange authorization code and return a signed JWT"),
		)
	}

	return routes
}
