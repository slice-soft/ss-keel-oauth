package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubProvider struct {
	cfg *oauth2.Config
}

var _ provider = (*githubProvider)(nil)

func newGitHubProvider(pc *ProviderConfig) *githubProvider {
	scopes := pc.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:user", "user:email"}
	}
	return &githubProvider{
		cfg: &oauth2.Config{
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			RedirectURL:  pc.RedirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *githubProvider) providerName() ProviderName { return ProviderGitHub }
func (p *githubProvider) config() *oauth2.Config      { return p.cfg }

type githubUserResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

type githubEmailEntry struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *githubProvider) userInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.cfg.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("github: fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: user responded %d", resp.StatusCode)
	}

	var u githubUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("github: decode user: %w", err)
	}

	email := u.Email
	if email == "" {
		email, _ = primaryGitHubEmail(client)
	}

	return &UserInfo{
		Provider:  ProviderGitHub,
		ID:        fmt.Sprintf("%d", u.ID),
		Email:     email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	}, nil
}

// primaryGitHubEmail fetches the verified primary email from the /user/emails endpoint.
// GitHub may return an empty email on /user when the user has set their email to private.
func primaryGitHubEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("github: fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github: emails responded %d", resp.StatusCode)
	}

	var emails []githubEmailEntry
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("github: decode emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}
