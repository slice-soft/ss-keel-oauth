package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"golang.org/x/oauth2"
)

// OAuth is the main entry point for the ss-keel-oauth addon.
//
// Typical setup in cmd/main.go:
//
//	oauthManager := oauth.New(oauth.Config{
//	    Google: &oauth.ProviderConfig{...},
//	    Signer: jwtAddon,
//	    Logger: appLogger,
//	})
//	app.RegisterController(oauth.NewController(oauthManager))
type OAuth struct {
	cfg       Config
	providers map[ProviderName]provider
}

// New creates an OAuth manager from the given Config.
// Only providers with a non-nil ProviderConfig are activated.
// Panics if Config.Signer is nil.
func New(cfg Config) *OAuth {
	if cfg.Signer == nil {
		panic("ss-keel-oauth: Config.Signer is required — pass a TokenSigner (e.g., from ss-keel-jwt)")
	}

	o := &OAuth{
		cfg:       cfg,
		providers: make(map[ProviderName]provider),
	}

	if cfg.Google != nil {
		o.providers[ProviderGoogle] = newGoogleProvider(cfg.Google)
	}
	if cfg.GitHub != nil {
		o.providers[ProviderGitHub] = newGitHubProvider(cfg.GitHub)
	}
	if cfg.GitLab != nil {
		o.providers[ProviderGitLab] = newGitLabProvider(cfg.GitLab)
	}

	return o
}

// LoginHandler returns a handler that redirects the browser to the
// provider's OAuth2 authorization page.
//
// Use NewController to register all routes at once, or call this directly
// if you need a custom route path:
//
//	httpx.GET("/sign-in/google", oauthManager.LoginHandler(oauth.ProviderGoogle))
func (o *OAuth) LoginHandler(name ProviderName) func(*httpx.Ctx) error {
	p, ok := o.providers[name]
	if !ok {
		return providerNotConfigured(name)
	}

	return func(c *httpx.Ctx) error {
		state, err := generateState()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to generate oauth state")
		}
		url := p.config().AuthCodeURL(state, oauth2.AccessTypeOnline)
		return c.Redirect(url)
	}
}

// CallbackHandler returns a handler that:
//  1. Exchanges the authorization code for an OAuth2 token.
//  2. Fetches the normalized UserInfo from the provider.
//  3. Signs a JWT via Config.Signer — subject: "<provider>:<user-id>".
//  4. Returns JSON {"token": "<jwt>"} or redirects to Config.RedirectOnSuccess?token=<jwt>.
//
// Use NewController to register all routes at once, or call this directly
// if you need a custom route path:
//
//	httpx.GET("/sign-in/google/callback", oauthManager.CallbackHandler(oauth.ProviderGoogle))
func (o *OAuth) CallbackHandler(name ProviderName) func(*httpx.Ctx) error {
	p, ok := o.providers[name]
	if !ok {
		return providerNotConfigured(name)
	}

	return func(c *httpx.Ctx) error {
		code := c.Query("code")
		if code == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing oauth code")
		}

		token, err := p.config().Exchange(context.Background(), code)
		if err != nil {
			o.logWarn("CallbackHandler[%s]: exchange failed: %v", name, err)
			return fiber.NewError(fiber.StatusUnauthorized, "oauth exchange failed")
		}

		userInfo, err := p.userInfo(context.Background(), token)
		if err != nil {
			o.logWarn("CallbackHandler[%s]: user info failed: %v", name, err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch user info")
		}

		subject := fmt.Sprintf("%s:%s", name, userInfo.ID)
		claims := map[string]any{
			"email":      userInfo.Email,
			"name":       userInfo.Name,
			"avatar_url": userInfo.AvatarURL,
			"provider":   string(userInfo.Provider),
		}

		jwt, err := o.cfg.Signer.Sign(subject, claims)
		if err != nil {
			o.logWarn("CallbackHandler[%s]: sign failed: %v", name, err)
			return fiber.NewError(fiber.StatusInternalServerError, "token signing failed")
		}

		if o.cfg.RedirectOnSuccess != "" {
			param := o.cfg.RedirectTokenParam
			if param == "" {
				param = "token"
			}
			return c.Redirect(o.cfg.RedirectOnSuccess + "?" + param + "=" + jwt)
		}
		return c.JSON(fiber.Map{"token": jwt})
	}
}

func (o *OAuth) logWarn(format string, args ...interface{}) {
	if o.cfg.Logger != nil {
		o.cfg.Logger.Warn(format, args...)
	}
}

func providerNotConfigured(name ProviderName) func(*httpx.Ctx) error {
	return func(c *httpx.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("oauth provider %q not configured", name))
	}
}

// generateState returns a cryptographically random hex string used as the
// OAuth2 state parameter to prevent CSRF attacks.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
