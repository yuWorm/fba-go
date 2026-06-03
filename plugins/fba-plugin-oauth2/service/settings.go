package service

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	SourceGithub = "Github"
	SourceGoogle = "Google"

	AuthTypeLogin   = "login"
	AuthTypeBinding = "binding"
)

type Settings struct {
	FrontendBindingRedirectURI string
	FrontendLoginRedirectURI   string
	GithubRedirectURI          string
	GoogleRedirectURI          string
	StateExpire                time.Duration
	StateRedisPrefix           string
	GithubClientID             string
	GithubClientSecret         string
	GoogleClientID             string
	GoogleClientSecret         string
	HTTPClient                 *http.Client
}

func DefaultSettings() Settings {
	return Settings{
		FrontendBindingRedirectURI: getenv("OAUTH2_FRONTEND_BINDING_REDIRECT_URI", "http://localhost:5173/profile"),
		FrontendLoginRedirectURI:   getenv("OAUTH2_FRONTEND_LOGIN_REDIRECT_URI", "http://localhost:5173/oauth2/callback"),
		GithubRedirectURI:          getenv("OAUTH2_GITHUB_REDIRECT_URI", "http://127.0.0.1:8000/api/v1/oauth2/github/callback"),
		GoogleRedirectURI:          getenv("OAUTH2_GOOGLE_REDIRECT_URI", "http://127.0.0.1:8000/api/v1/oauth2/google/callback"),
		StateExpire:                time.Duration(getenvInt("OAUTH2_STATE_EXPIRE_SECONDS", 180)) * time.Second,
		StateRedisPrefix:           getenv("OAUTH2_STATE_REDIS_PREFIX", "fba:oauth2:state"),
		GithubClientID:             os.Getenv("OAUTH2_GITHUB_CLIENT_ID"),
		GithubClientSecret:         os.Getenv("OAUTH2_GITHUB_CLIENT_SECRET"),
		GoogleClientID:             os.Getenv("OAUTH2_GOOGLE_CLIENT_ID"),
		GoogleClientSecret:         os.Getenv("OAUTH2_GOOGLE_CLIENT_SECRET"),
		HTTPClient:                 &http.Client{Timeout: 10 * time.Second},
	}
}

func (s Settings) withDefaults() Settings {
	defaults := DefaultSettings()
	if s.FrontendBindingRedirectURI == "" {
		s.FrontendBindingRedirectURI = defaults.FrontendBindingRedirectURI
	}
	if s.FrontendLoginRedirectURI == "" {
		s.FrontendLoginRedirectURI = defaults.FrontendLoginRedirectURI
	}
	if s.GithubRedirectURI == "" {
		s.GithubRedirectURI = defaults.GithubRedirectURI
	}
	if s.GoogleRedirectURI == "" {
		s.GoogleRedirectURI = defaults.GoogleRedirectURI
	}
	if s.StateExpire == 0 {
		s.StateExpire = defaults.StateExpire
	}
	if s.StateRedisPrefix == "" {
		s.StateRedisPrefix = defaults.StateRedisPrefix
	}
	if s.HTTPClient == nil {
		s.HTTPClient = defaults.HTTPClient
	}
	return s
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
