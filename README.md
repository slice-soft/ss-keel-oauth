<img src="https://cdn.slicesoft.dev/boat.svg" width="400" />

# ss-keel-oauth
OAuth2 authentication addon for Keel — Google, GitHub, and GitLab providers with JWT issuance on callback.

[![CI](https://github.com/slice-soft/ss-keel-oauth/actions/workflows/ci.yml/badge.svg)](https://github.com/slice-soft/ss-keel-oauth/actions)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/slice-soft/ss-keel-oauth)](https://goreportcard.com/report/github.com/slice-soft/ss-keel-oauth)
[![Go Reference](https://pkg.go.dev/badge/github.com/slice-soft/ss-keel-oauth.svg)](https://pkg.go.dev/github.com/slice-soft/ss-keel-oauth)
![License](https://img.shields.io/badge/License-MIT-green)
![Made in Colombia](https://img.shields.io/badge/Made%20in-Colombia-FCD116?labelColor=003893)


## OAuth2 authentication addon for Keel

`ss-keel-oauth` adds OAuth2 authentication to any Keel application.
After a successful provider flow the addon signs a JWT and returns it to the client — either as JSON or as a redirect with the token in the query string.

**Supported providers:** Google · GitHub · GitLab

---

## 🚀 Installation

```bash
keel add oauth
```

The Keel CLI will:
1. Add `github.com/slice-soft/ss-keel-oauth` as a dependency.
2. Create `cmd/setup_oauth.go` and inject initialization code into `cmd/main.go`.
3. Add OAuth provider environment variable examples to your `.env`.

Manual install:

```bash
go get github.com/slice-soft/ss-keel-oauth
```

---

## ⚙️ Environment variables

| Variable | Description |
|---|---|
| `OAUTH_GOOGLE_CLIENT_ID` | Google OAuth2 client ID ([console.cloud.google.com](https://console.cloud.google.com/apis/credentials)) |
| `OAUTH_GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret |
| `OAUTH_GITHUB_CLIENT_ID` | GitHub OAuth2 client ID ([github.com/settings/developers](https://github.com/settings/developers)) |
| `OAUTH_GITHUB_CLIENT_SECRET` | GitHub OAuth2 client secret |
| `OAUTH_GITLAB_CLIENT_ID` | GitLab OAuth2 application ID ([gitlab.com/-/user_settings/applications](https://gitlab.com/-/user_settings/applications)) |
| `OAUTH_GITLAB_CLIENT_SECRET` | GitLab OAuth2 client secret |
| `OAUTH_REDIRECT_BASE_URL` | Base URL for building callback URLs (e.g. `http://localhost:7331` in dev, `https://api.myapp.com` in prod) |
| `OAUTH_ROUTE_PREFIX` | Route prefix used to expose OAuth login and callback routes (default: `/auth`) |
| `OAUTH_ENABLED_PROVIDERS` | Optional comma-separated list of enabled providers (`google,github,gitlab`) |

`keel add oauth` generates a `cmd/setup_oauth.go` that reads Google, GitHub, and GitLab credentials from `.env`. Providers are only activated when both client ID and client secret are present. Set `OAUTH_ENABLED_PROVIDERS=google,github` to restrict the exposed routes further.

---

## ⚡️ Quick start

```go
// cmd/setup_oauth.go — created by keel add oauth
package main

import (
    "github.com/slice-soft/ss-keel-core/config"
    "github.com/slice-soft/ss-keel-core/core"
    "github.com/slice-soft/ss-keel-core/logger"
    "github.com/slice-soft/ss-keel-jwt/jwt"
    "github.com/slice-soft/ss-keel-oauth/oauth"
)

// setupOAuth registers the OAuth2 controller for the configured providers.
// jwtProvider is used to sign the JWT returned after a successful OAuth flow.
func setupOAuth(app *core.App, jwtProvider *jwt.JWT, log *logger.Logger) {
    redirectBase := config.GetEnvOrDefault("OAUTH_REDIRECT_BASE_URL", "http://localhost:7331")
    oauthManager := oauth.New(oauth.Config{
        Google: &oauth.ProviderConfig{
            ClientID:     config.GetEnvOrDefault("OAUTH_GOOGLE_CLIENT_ID", ""),
            ClientSecret: config.GetEnvOrDefault("OAUTH_GOOGLE_CLIENT_SECRET", ""),
            RedirectURL:  redirectBase + "/auth/google/callback",
        },
        Signer: jwtProvider,
        Logger: log,
    })
    app.RegisterController(oauth.NewController(oauthManager))
}
```

The generated `cmd/setup_oauth.go` created by `keel add oauth` goes further than this minimal example: it normalizes `OAUTH_ROUTE_PREFIX`, auto-builds callback URLs from `OAUTH_REDIRECT_BASE_URL`, and only registers routes for providers with complete credentials.

---

## 🔌 TokenSigner interface

`ss-keel-oauth` depends on `contracts.TokenSigner` from `ss-keel-core` — not on `ss-keel-jwt` directly.
`ss-keel-jwt` satisfies this interface, but any custom implementation works:

```go
type TokenSigner interface {
    Sign(subject string, data map[string]any) (string, error)
}
```

The signed token's `subject` is `"<provider>:<user-id>"` (e.g. `"google:1234567890"`).
The `data` map includes: `email`, `name`, `avatar_url`, `provider`.

---

## 🔗 Routes

`NewController` registers the following routes for every enabled provider:

| Route | Description |
|---|---|
| `GET /auth/google` | Redirects to Google's authorization page |
| `GET /auth/google/callback` | Exchanges code, signs JWT, returns token |
| `GET /auth/github` | Redirects to GitHub's authorization page |
| `GET /auth/github/callback` | Exchanges code, signs JWT, returns token |
| `GET /auth/gitlab` | Redirects to GitLab's authorization page |
| `GET /auth/gitlab/callback` | Exchanges code, signs JWT, returns token |

---

## ❤️ Callback response modes

**Mode 1 — JSON** (default, recommended for APIs and mobile): the callback returns `{ "token": "<jwt>" }`.

**Mode 2 — Redirect**: set `RedirectOnSuccess` to redirect the browser to your frontend with the token as a query parameter.

```go
oauth.New(oauth.Config{
    Google:            &oauth.ProviderConfig{...},
    Signer:            jwtProvider,
    RedirectOnSuccess: "https://myapp.com/auth/done",
})
```

> Tokens in query strings appear in access logs and browser history. After reading the token, remove it from the URL with `history.replaceState`.

---

## 🤚 CI/CD and releases

- **CI** runs on every pull request targeting `main` via `.github/workflows/ci.yml`.
- **Releases** are created automatically on merge to `main` via `.github/workflows/release.yml` using Release Please.

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for setup and repository-specific rules.
The base workflow, commit conventions, and community standards live in [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md).

## Community

| Document | |
|---|---|
| [CONTRIBUTING.md](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) | Workflow, commit conventions, and PR guidelines |
| [GOVERNANCE.md](https://github.com/slice-soft/ss-community/blob/main/GOVERNANCE.md) | Decision-making, roles, and release process |
| [CODE_OF_CONDUCT.md](https://github.com/slice-soft/ss-community/blob/main/CODE_OF_CONDUCT.md) | Community standards |
| [VERSIONING.md](https://github.com/slice-soft/ss-community/blob/main/VERSIONING.md) | SemVer policy and breaking changes |
| [SECURITY.md](https://github.com/slice-soft/ss-community/blob/main/SECURITY.md) | How to report vulnerabilities |
| [MAINTAINERS.md](https://github.com/slice-soft/ss-community/blob/main/MAINTAINERS.md) | Active maintainers |

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- Website: [keel-go.dev](https://keel-go.dev)
- GitHub: [github.com/slice-soft/ss-keel-oauth](https://github.com/slice-soft/ss-keel-oauth)
- Documentation: [docs.keel-go.dev](https://docs.keel-go.dev)

---

Made by [SliceSoft](https://slicesoft.dev) — Colombia 💙
