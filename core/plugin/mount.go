package plugin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-go/core/response"
)

const CurrentUserLocalKey = "fba.current_user"

type Authenticator interface {
	Authenticate(c fiber.Ctx) (*rbac.CurrentUser, error)
}

type AuthenticatorFunc func(c fiber.Ctx) (*rbac.CurrentUser, error)

func (fn AuthenticatorFunc) Authenticate(c fiber.Ctx) (*rbac.CurrentUser, error) {
	return fn(c)
}

type MountOption func(*mountOptions)

type mountOptions struct {
	authenticator Authenticator
	container     *di.Container
}

func WithAuthenticator(authenticator Authenticator) MountOption {
	return func(opts *mountOptions) {
		opts.authenticator = authenticator
	}
}

func WithContainer(container *di.Container) MountOption {
	return func(opts *mountOptions) {
		opts.container = container
	}
}

func MountRoutes(router fiber.Router, routes []Route, options ...MountOption) {
	opts := mountOptions{}
	for _, option := range options {
		option(&opts)
	}
	if opts.authenticator == nil && opts.container != nil {
		var authenticator Authenticator
		if opts.container.Resolve(&authenticator) {
			opts.authenticator = authenticator
		}
	}

	for _, item := range routes {
		route := item
		handler := route.Handler
		if route.AuthRequired || route.Permission != "" {
			handler = authHandler(route, opts.authenticator)
		}
		router.Add([]string{strings.ToUpper(route.Method)}, route.Path, handler)
	}
}

func authHandler(route Route, authenticator Authenticator) fiber.Handler {
	return func(c fiber.Ctx) error {
		if authenticator == nil {
			return authFailure(c, http.StatusUnauthorized, "未认证")
		}
		user, err := authenticator.Authenticate(c)
		if err != nil {
			return authFailure(c, http.StatusUnauthorized, "未认证")
		}
		if err := rbac.Authorize(user, rbac.RouteAccess{
			Method:      route.Method,
			Permission:  route.Permission,
			Whitelisted: !route.AuthRequired && route.Permission == "",
		}); err != nil {
			return authFailure(c, authStatus(err), authMessage(err))
		}
		c.Locals(CurrentUserLocalKey, user)
		return route.Handler(c)
	}
}

func authStatus(err error) int {
	if errors.Is(err, rbac.ErrUnauthenticated) {
		return http.StatusUnauthorized
	}
	return http.StatusForbidden
}

func authMessage(err error) string {
	switch {
	case errors.Is(err, rbac.ErrUnauthenticated):
		return "未认证"
	case errors.Is(err, rbac.ErrNoEnabledRole):
		return "无可用角色"
	case errors.Is(err, rbac.ErrStaffRequired):
		return "需要管理员权限"
	default:
		return "无权限"
	}
}

func authFailure(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(response.Error(status, message, middleware.RequestIDFromCtx(c)))
}
