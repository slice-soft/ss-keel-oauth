package oauth_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-oauth/oauth"
)

// stubSigner is a test double that returns a fixed token or a configured error.
type stubSigner struct{ err error }

func (s *stubSigner) Sign(_ string, _ map[string]any) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return "signed-token", nil
}

func TestNew_PanicsWithNilSigner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when Signer is nil")
		}
	}()
	oauth.New(oauth.Config{}) //nolint:exhaustruct
}

func TestLoginHandler_UnconfiguredProvider(t *testing.T) {
	o := oauth.New(oauth.Config{Signer: &stubSigner{}})
	app := fiber.New()
	app.Get("/auth/google", httpx.WrapHandler(o.LoginHandler(oauth.ProviderGoogle)))

	req := httptest.NewRequest("GET", "/auth/google", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404 got %d", resp.StatusCode)
	}
}

func TestCallbackHandler_UnconfiguredProvider(t *testing.T) {
	o := oauth.New(oauth.Config{Signer: &stubSigner{}})
	app := fiber.New()
	app.Get("/auth/github/callback", httpx.WrapHandler(o.CallbackHandler(oauth.ProviderGitHub)))

	req := httptest.NewRequest("GET", "/auth/github/callback?code=abc", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404 got %d", resp.StatusCode)
	}
}

func TestCallbackHandler_MissingCode(t *testing.T) {
	o := oauth.New(oauth.Config{
		Google: &oauth.ProviderConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "http://localhost/auth/google/callback",
		},
		Signer: &stubSigner{},
	})
	app := fiber.New()
	app.Get("/auth/google/callback", httpx.WrapHandler(o.CallbackHandler(oauth.ProviderGoogle)))

	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400 got %d", resp.StatusCode)
	}
}

func TestNewController_DefaultPrefix(t *testing.T) {
	o := oauth.New(oauth.Config{
		Google: &oauth.ProviderConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/auth/google/callback",
		},
		Signer: &stubSigner{},
	})

	ctrl := oauth.NewController(o)
	routes := ctrl.Routes()

	if len(routes) != 2 {
		t.Errorf("expected 2 routes (login + callback) got %d", len(routes))
	}
	if routes[0].Path() != "/auth/google" {
		t.Errorf("expected /auth/google got %s", routes[0].Path())
	}
	if routes[1].Path() != "/auth/google/callback" {
		t.Errorf("expected /auth/google/callback got %s", routes[1].Path())
	}
}

func TestNewController_CustomPrefix(t *testing.T) {
	o := oauth.New(oauth.Config{
		GitHub: &oauth.ProviderConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/sign-in/github/callback",
		},
		Signer: &stubSigner{},
	})

	ctrl := oauth.NewController(o, "/sign-in")
	routes := ctrl.Routes()

	if len(routes) != 2 {
		t.Errorf("expected 2 routes got %d", len(routes))
	}
	if routes[0].Path() != "/sign-in/github" {
		t.Errorf("expected /sign-in/github got %s", routes[0].Path())
	}
}
