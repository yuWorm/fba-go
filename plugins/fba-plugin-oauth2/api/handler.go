package api

import (
	"net/url"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-go/core/response"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	"github.com/yuWorm/fba-plugin-oauth2/service"
)

const (
	refreshCookieName          = "fba_refresh_token"
	refreshCookieMaxAgeSeconds = 60 * 60 * 24 * 7
	defaultCurrentUserID       = 1
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) Handler {
	if svc == nil {
		svc = service.New(service.Options{})
	}
	return Handler{service: svc}
}

func (h Handler) GetGithubAuthURL(c fiber.Ctx) error {
	return h.authURL(c, service.SourceGithub)
}

func (h Handler) GetGoogleAuthURL(c fiber.Ctx) error {
	return h.authURL(c, service.SourceGoogle)
}

func (h Handler) authURL(c fiber.Ctx, source string) error {
	authURL, err := h.service.LoginAuthURL(c.RequestCtx(), source)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(authURL))
}

func (h Handler) GithubCallback(c fiber.Ctx) error {
	return h.callback(c, service.SourceGithub)
}

func (h Handler) GoogleCallback(c fiber.Ctx) error {
	return h.callback(c, service.SourceGoogle)
}

func (h Handler) callback(c fiber.Ctx, source string) error {
	result, err := h.service.Callback(c.RequestCtx(), source, c.Query("code"), c.Query("state"), requestMetadata(c))
	if err != nil {
		return err
	}
	if result.Binding {
		return c.Redirect().Status(fiber.StatusFound).To(h.service.Settings().FrontendBindingRedirectURI)
	}
	setRefreshCookie(c, result.Refresh)
	redirectURL, err := loginRedirectURL(h.service.Settings().FrontendLoginRedirectURI, result.AccessToken, result.SessionUUID)
	if err != nil {
		return err
	}
	return c.Redirect().Status(fiber.StatusFound).To(redirectURL)
}

func (h Handler) GetBindings(c fiber.Ctx) error {
	bindings, err := h.service.Bindings(c.RequestCtx(), currentUserID(c))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(bindings))
}

func (h Handler) GetBindingAuthURL(c fiber.Ctx) error {
	authURL, err := h.service.BindingAuthURL(c.RequestCtx(), currentUserID(c), c.Query("source"))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(authURL))
}

func (h Handler) Unbind(c fiber.Ctx) error {
	if err := h.service.Unbind(c.RequestCtx(), currentUserID(c), c.Query("source")); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func loginRedirectURL(base string, accessToken string, sessionUUID string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("access_token", accessToken)
	query.Set("session_uuid", sessionUUID)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func setRefreshCookie(c fiber.Ctx, value string) {
	if value == "" {
		value = "fixture-refresh-token"
	}
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		MaxAge:   refreshCookieMaxAgeSeconds,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func requestMetadata(c fiber.Ctx) adminservice.RequestMetadata {
	return adminservice.RequestMetadata{
		IP:        c.IP(),
		UserAgent: optionalString(c.Get("User-Agent")),
	}
}

func currentUserID(c fiber.Ctx) int {
	user, ok := c.Locals(plugin.CurrentUserLocalKey).(*rbac.CurrentUser)
	if !ok || user == nil || user.ID <= 0 {
		return defaultCurrentUserID
	}
	return int(user.ID)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
