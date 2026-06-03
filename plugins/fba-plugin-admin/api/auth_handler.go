package api

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"github.com/yuWorm/fba-plugin-admin/service"
)

const refreshCookieName = "fba_refresh_token"
const refreshCookieMaxAgeSeconds = 60 * 60 * 24 * 7

type Handler struct {
	auth       *service.AuthService
	users      *service.UserService
	roles      *service.RoleService
	menus      *service.MenuService
	depts      *service.DeptService
	dataRules  *service.DataRuleService
	dataScopes *service.DataScopeService
	plugins    *service.PluginService
	logs       *service.LogService
	files      *service.FileService
	monitors   *service.MonitorService
}

type HandlerOptions struct {
	Config         config.Options
	ConfigProvider service.AdminConfigProvider
	Redis          service.RedisClient
}

func NewHandler() Handler {
	repository := repo.NewMemoryRepository(repo.SeedData())
	return NewHandlerWithRepository(repository)
}

func NewHandlerWithRepository(repository repo.Repository) Handler {
	return NewHandlerWithOptions(repository, config.Options{})
}

func NewHandlerWithOptions(repository repo.Repository, opts config.Options) Handler {
	return NewHandlerWithAdminOptions(repository, HandlerOptions{Config: opts})
}

func NewHandlerWithAdminOptions(repository repo.Repository, opts HandlerOptions) Handler {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return Handler{
		auth: service.NewAuthServiceWithOptions(repository, service.AuthServiceOptions{
			ConfigProvider: opts.ConfigProvider,
			Redis:          opts.Redis,
			RedisKeyPrefix: opts.Config.Redis.KeyPrefix,
		}),
		users: service.NewUserServiceWithOptions(repository, service.UserServiceOptions{
			ConfigProvider: opts.ConfigProvider,
			Redis:          opts.Redis,
		}),
		roles:      service.NewRoleService(repository),
		menus:      service.NewMenuService(repository),
		depts:      service.NewDeptService(repository),
		dataRules:  service.NewDataRuleService(repository),
		dataScopes: service.NewDataScopeService(repository),
		plugins:    service.NewPluginServiceWithConfig(repository, opts.Config),
		logs:       service.NewLogService(repository),
		files:      service.NewFileService(),
		monitors:   service.NewMonitorServiceWithRedis(repository, opts.Redis),
	}
}

func (h Handler) Captcha(c fiber.Ctx) error {
	captcha, err := h.auth.Captcha(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(captcha))
}

func (h Handler) LoginSwagger(c fiber.Ctx) error {
	username, password := basicCredentials(c.Get("Authorization"))
	token, err := h.auth.SwaggerLogin(c.RequestCtx(), username, password)
	if err != nil {
		return err
	}
	return c.JSON(token)
}

func (h Handler) Login(c fiber.Ctx) error {
	var param dto.AuthLoginParam
	if c.HasBody() {
		if err := c.Bind().Body(&param); err != nil {
			return err
		}
	}
	token, refresh, err := h.auth.Login(c.RequestCtx(), param, requestMetadata(c))
	if err != nil {
		return err
	}
	setRefreshCookie(c, refresh)
	return c.JSON(response.Success(token))
}

func (h Handler) Refresh(c fiber.Ctx) error {
	token, refresh, err := h.auth.Refresh(c.RequestCtx(), c.Cookies(refreshCookieName))
	if err != nil {
		return err
	}
	setRefreshCookie(c, refresh)
	return c.JSON(response.Success(token))
}

func (h Handler) Logout(c fiber.Ctx) error {
	if err := h.auth.Logout(c.RequestCtx(), c.Get("Authorization")); err != nil {
		return err
	}
	c.ClearCookie(refreshCookieName)
	return c.JSON(response.Success[any](nil))
}

func requestMetadata(c fiber.Ctx) service.RequestMetadata {
	return service.RequestMetadata{
		IP:        c.IP(),
		UserAgent: optionalString(c.Get("User-Agent")),
	}
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (h Handler) Codes(c fiber.Ctx) error {
	codes, err := h.auth.Codes(c.RequestCtx(), currentUser(c))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(codes))
}

func (h Handler) Authenticate(c fiber.Ctx) (*rbac.CurrentUser, error) {
	return h.auth.Authenticate(c.RequestCtx(), c.Get("Authorization"))
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

func basicCredentials(header string) (string, string) {
	if !strings.HasPrefix(strings.ToLower(header), "basic ") {
		return "", ""
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[6:]))
	if err != nil {
		return "", ""
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return "", ""
	}
	return username, password
}
