package plugin

import "github.com/gofiber/fiber/v3"

type Route struct {
	Method            string
	Path              string
	Summary           string
	Tags              []string
	Metadata          map[string]string
	Permission        string
	AuthRequired      bool
	SuperuserRequired bool
	Handler           fiber.Handler
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

func Superuser() RouteOption {
	return func(route *Route) {
		route.AuthRequired = true
		route.SuperuserRequired = true
	}
}

func Tags(tags ...string) RouteOption {
	return func(route *Route) {
		route.Tags = append([]string(nil), tags...)
	}
}

// Metadata copies extension metadata onto a route. Framework consumers should
// reserve namespaced keys so independently developed plugins cannot collide.
func Metadata(values map[string]string) RouteOption {
	return func(route *Route) {
		if len(values) == 0 {
			route.Metadata = nil
			return
		}
		route.Metadata = make(map[string]string, len(values))
		for key, value := range values {
			route.Metadata[key] = value
		}
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
