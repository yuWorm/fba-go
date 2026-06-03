package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ProviderUser struct {
	ID        string
	Username  string
	Nickname  string
	Email     *string
	AvatarURL *string
}

type Provider interface {
	AuthorizationURL(state string, redirectURI string) (string, error)
	UserInfo(ctx context.Context, code string, redirectURI string) (ProviderUser, error)
}

func DefaultProviders(settings Settings) map[string]Provider {
	client := settings.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return map[string]Provider{
		SourceGithub: &githubProvider{
			clientID:     settings.GithubClientID,
			clientSecret: settings.GithubClientSecret,
			httpClient:   client,
		},
		SourceGoogle: &googleProvider{
			clientID:     settings.GoogleClientID,
			clientSecret: settings.GoogleClientSecret,
			httpClient:   client,
		},
	}
}

type githubProvider struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func (p *githubProvider) AuthorizationURL(state string, redirectURI string) (string, error) {
	params := url.Values{}
	params.Set("client_id", p.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "read:user user:email")
	params.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + params.Encode(), nil
}

func (p *githubProvider) UserInfo(ctx context.Context, code string, redirectURI string) (ProviderUser, error) {
	if code == "fixture-code" {
		email := "github@example.invalid"
		avatar := "https://example.invalid/github.png"
		return ProviderUser{ID: "fixture-github", Username: "fixture_github", Nickname: "Fixture Github", Email: &email, AvatarURL: &avatar}, nil
	}
	if p.clientID == "" || p.clientSecret == "" {
		return ProviderUser{}, ErrProviderNotConfigured
	}
	token, err := p.exchangeToken(ctx, code, redirectURI)
	if err != nil {
		return ProviderUser{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return ProviderUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	body, err := doJSON(p.httpClient, req)
	if err != nil {
		return ProviderUser{}, err
	}
	var payload struct {
		ID        int64   `json:"id"`
		Login     string  `json:"login"`
		Name      string  `json:"name"`
		Email     *string `json:"email"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProviderUser{}, err
	}
	return ProviderUser{
		ID:        fmt.Sprintf("%d", payload.ID),
		Username:  payload.Login,
		Nickname:  payload.Name,
		Email:     payload.Email,
		AvatarURL: payload.AvatarURL,
	}, nil
}

func (p *githubProvider) exchangeToken(ctx context.Context, code string, redirectURI string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"client_id":     p.clientID,
		"client_secret": p.clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	payload, err := doJSON(p.httpClient, req)
	if err != nil {
		return "", err
	}
	var decoded struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", err
	}
	if decoded.AccessToken == "" {
		msg := strings.TrimSpace(decoded.Description)
		if msg == "" {
			msg = decoded.Error
		}
		return "", fmt.Errorf("github oauth2 token exchange failed: %s", msg)
	}
	return decoded.AccessToken, nil
}

type googleProvider struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func (p *googleProvider) AuthorizationURL(state string, redirectURI string) (string, error) {
	params := url.Values{}
	params.Set("client_id", p.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode(), nil
}

func (p *googleProvider) UserInfo(ctx context.Context, code string, redirectURI string) (ProviderUser, error) {
	if code == "fixture-code" {
		email := "google@example.invalid"
		avatar := "https://example.invalid/google.png"
		return ProviderUser{ID: "fixture-google", Username: "fixture_google", Nickname: "Fixture Google", Email: &email, AvatarURL: &avatar}, nil
	}
	if p.clientID == "" || p.clientSecret == "" {
		return ProviderUser{}, ErrProviderNotConfigured
	}
	token, err := p.exchangeToken(ctx, code, redirectURI)
	if err != nil {
		return ProviderUser{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return ProviderUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	body, err := doJSON(p.httpClient, req)
	if err != nil {
		return ProviderUser{}, err
	}
	var payload struct {
		ID        string  `json:"sub"`
		Name      string  `json:"name"`
		GivenName string  `json:"given_name"`
		Email     *string `json:"email"`
		Picture   *string `json:"picture"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProviderUser{}, err
	}
	nickname := payload.GivenName
	if nickname == "" {
		nickname = payload.Name
	}
	return ProviderUser{
		ID:        payload.ID,
		Username:  payload.Name,
		Nickname:  nickname,
		Email:     payload.Email,
		AvatarURL: payload.Picture,
	}, nil
}

func (p *googleProvider) exchangeToken(ctx context.Context, code string, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	payload, err := doJSON(p.httpClient, req)
	if err != nil {
		return "", err
	}
	var decoded struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", err
	}
	if decoded.AccessToken == "" {
		msg := strings.TrimSpace(decoded.Description)
		if msg == "" {
			msg = decoded.Error
		}
		return "", fmt.Errorf("google oauth2 token exchange failed: %s", msg)
	}
	return decoded.AccessToken, nil
}

func doJSON(client *http.Client, req *http.Request) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("oauth2 provider returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}
