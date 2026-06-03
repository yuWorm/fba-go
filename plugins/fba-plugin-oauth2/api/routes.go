package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/oauth2/github", "Get Github OAuth2 URL", h.GetGithubAuthURL),
		plugin.GET("/oauth2/github/callback", "Github OAuth2 callback", h.GithubCallback),
		plugin.GET("/oauth2/google", "Get Google OAuth2 URL", h.GetGoogleAuthURL),
		plugin.GET("/oauth2/google/callback", "Google OAuth2 callback", h.GoogleCallback),
		plugin.GET("/oauth2/me/bindings", "Get OAuth2 bindings", h.GetBindings, plugin.Auth()),
		plugin.GET("/oauth2/me/binding", "Get OAuth2 binding URL", h.GetBindingAuthURL, plugin.Auth()),
		plugin.DELETE("/oauth2/me/unbinding", "Unbind OAuth2 account", h.Unbind, plugin.Auth()),
	}
}
