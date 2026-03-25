package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/contracts"
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
	events    chan contracts.PanelEvent
}

// New creates an OAuth manager from the given Config.
// Only providers with a complete ProviderConfig are activated.
// Panics if Config.Signer is nil.
func New(cfg Config) *OAuth {
	if cfg.Signer == nil {
		panic("ss-keel-oauth: Config.Signer is required — pass a TokenSigner (e.g., from ss-keel-jwt)")
	}

	o := &OAuth{
		cfg:       cfg,
		providers: make(map[ProviderName]provider),
		events:    make(chan contracts.PanelEvent, 256),
	}

	if providerConfigReady(cfg.Google) {
		o.providers[ProviderGoogle] = newGoogleProvider(cfg.Google)
	} else if cfg.Google != nil && cfg.Logger != nil {
		cfg.Logger.Warn("oauth[%s]: provider config incomplete; skipping route registration", ProviderGoogle)
	}
	if providerConfigReady(cfg.GitHub) {
		o.providers[ProviderGitHub] = newGitHubProvider(cfg.GitHub)
	} else if cfg.GitHub != nil && cfg.Logger != nil {
		cfg.Logger.Warn("oauth[%s]: provider config incomplete; skipping route registration", ProviderGitHub)
	}
	if providerConfigReady(cfg.GitLab) {
		o.providers[ProviderGitLab] = newGitLabProvider(cfg.GitLab)
	} else if cfg.GitLab != nil && cfg.Logger != nil {
		cfg.Logger.Warn("oauth[%s]: provider config incomplete; skipping route registration", ProviderGitLab)
	}

	return o
}

func providerConfigReady(pc *ProviderConfig) bool {
	if pc == nil {
		return false
	}
	return strings.TrimSpace(pc.ClientID) != "" &&
		strings.TrimSpace(pc.ClientSecret) != "" &&
		strings.TrimSpace(pc.RedirectURL) != ""
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
		o.tryEmit(contracts.PanelEvent{
			Timestamp: time.Now(),
			AddonID:   "oauth",
			Label:     "login",
			Level:     "info",
			Detail: map[string]any{
				"provider": string(name),
				"result":   "flow_started",
			},
		})
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
			o.tryEmit(contracts.PanelEvent{
				Timestamp: time.Now(),
				AddonID:   "oauth",
				Label:     "callback",
				Level:     "error",
				Detail: map[string]any{
					"provider": string(name),
					"result":   "error",
					"error":    "missing oauth code",
				},
			})
			return fiber.NewError(fiber.StatusBadRequest, "missing oauth code")
		}

		token, err := p.config().Exchange(context.Background(), code)
		if err != nil {
			o.logWarn("CallbackHandler[%s]: exchange failed: %v", name, err)
			o.tryEmit(contracts.PanelEvent{
				Timestamp: time.Now(),
				AddonID:   "oauth",
				Label:     "callback",
				Level:     "error",
				Detail: map[string]any{
					"provider": string(name),
					"result":   "error",
					"error":    "exchange failed",
				},
			})
			return fiber.NewError(fiber.StatusUnauthorized, "oauth exchange failed")
		}

		userInfo, err := p.userInfo(context.Background(), token)
		if err != nil {
			o.logWarn("CallbackHandler[%s]: user info failed: %v", name, err)
			o.tryEmit(contracts.PanelEvent{
				Timestamp: time.Now(),
				AddonID:   "oauth",
				Label:     "callback",
				Level:     "error",
				Detail: map[string]any{
					"provider": string(name),
					"result":   "error",
					"error":    "failed to fetch user info",
				},
			})
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
			o.tryEmit(contracts.PanelEvent{
				Timestamp: time.Now(),
				AddonID:   "oauth",
				Label:     "callback",
				Level:     "error",
				Detail: map[string]any{
					"provider": string(name),
					"result":   "error",
					"error":    "token signing failed",
				},
			})
			return fiber.NewError(fiber.StatusInternalServerError, "token signing failed")
		}

		o.tryEmit(contracts.PanelEvent{
			Timestamp: time.Now(),
			AddonID:   "oauth",
			Label:     "callback",
			Level:     "info",
			Detail: map[string]any{
				"provider": string(name),
				"user_id":  userInfo.ID,
				"result":   "ok",
			},
		})

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

// tryEmit sends a PanelEvent to the events channel without blocking.
// Events are silently dropped when the channel buffer is full.
func (o *OAuth) tryEmit(e contracts.PanelEvent) {
	select {
	case o.events <- e:
	default:
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
