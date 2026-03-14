package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleProvider struct {
	cfg *oauth2.Config
}

var _ provider = (*googleProvider)(nil)

func newGoogleProvider(pc *ProviderConfig) *googleProvider {
	scopes := pc.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "email", "profile"}
	}
	return &googleProvider{
		cfg: &oauth2.Config{
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			RedirectURL:  pc.RedirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *googleProvider) providerName() ProviderName { return ProviderGoogle }
func (p *googleProvider) config() *oauth2.Config      { return p.cfg }

type googleUserResponse struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func (p *googleProvider) userInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google: fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google: user info responded %d", resp.StatusCode)
	}

	var u googleUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("google: decode user info: %w", err)
	}

	return &UserInfo{
		Provider:  ProviderGoogle,
		ID:        u.Sub,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.Picture,
	}, nil
}
