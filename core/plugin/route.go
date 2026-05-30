package plugin

import "github.com/gofiber/fiber/v3"

type Route struct {
	Method       string
	Path         string
	Summary      string
	Tags         []string
	Permission   string
	AuthRequired bool
	Handler      fiber.Handler
}

type RouteOption func(*Route)

func Auth() RouteOption {
	return func(route *Route) {
		route.AuthRequired = true
	}
}

func Perm(permission string) RouteOption {
	return func(route *Route) {
		route.Permission = permission
	}
}

func Tags(tags ...string) RouteOption {
	return func(route *Route) {
		route.Tags = append([]string(nil), tags...)
	}
}

func GET(path string, summary string, handler fiber.Handler, opts ...RouteOption) Route {
	return newRoute("GET", path, summary, handler, opts...)
}

func POST(path string, summary string, handler fiber.Handler, opts ...RouteOption) Route {
	return newRoute("POST", path, summary, handler, opts...)
}

func PUT(path string, summary string, handler fiber.Handler, opts ...RouteOption) Route {
	return newRoute("PUT", path, summary, handler, opts...)
}

func DELETE(path string, summary string, handler fiber.Handler, opts ...RouteOption) Route {
	return newRoute("DELETE", path, summary, handler, opts...)
}

func RegisterRoutes(ctx Context, groups ...[]Route) error {
	for _, group := range groups {
		for _, route := range group {
			if err := ctx.Route(route); err != nil {
				return err
			}
		}
	}
	return nil
}

func newRoute(method string, path string, summary string, handler fiber.Handler, opts ...RouteOption) Route {
	route := Route{
		Method:  method,
		Path:    path,
		Summary: summary,
		Handler: handler,
	}
	for _, opt := range opts {
		opt(&route)
	}
	return route
}
