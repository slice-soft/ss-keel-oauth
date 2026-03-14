package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// gitlabEndpoint is the OAuth2 endpoint for gitlab.com.
// Self-hosted GitLab instances require a different base URL; update ProviderConfig.RedirectURL
// accordingly and fork the provider if you need a custom token endpoint.
var gitlabEndpoint = oauth2.Endpoint{
	AuthURL:  "https://gitlab.com/oauth/authorize",
	TokenURL: "https://gitlab.com/oauth/token",
}

type gitlabProvider struct {
	cfg *oauth2.Config
}

var _ provider = (*gitlabProvider)(nil)

func newGitLabProvider(pc *ProviderConfig) *gitlabProvider {
	scopes := pc.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read_user"}
	}
	return &gitlabProvider{
		cfg: &oauth2.Config{
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			RedirectURL:  pc.RedirectURL,
			Scopes:       scopes,
			Endpoint:     gitlabEndpoint,
		},
	}
}

func (p *gitlabProvider) providerName() ProviderName { return ProviderGitLab }
func (p *gitlabProvider) config() *oauth2.Config      { return p.cfg }

type gitlabUserResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func (p *gitlabProvider) userInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.cfg.Client(ctx, token)
	resp, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		return nil, fmt.Errorf("gitlab: fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitlab: user responded %d", resp.StatusCode)
	}

	var u gitlabUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("gitlab: decode user: %w", err)
	}

	return &UserInfo{
		Provider:  ProviderGitLab,
		ID:        fmt.Sprintf("%d", u.ID),
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	}, nil
}
