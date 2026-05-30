package observability

import "github.com/gofiber/fiber/v3"

func MetricsHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/plain; version=0.0.4; charset=utf-8")
		return c.SendString("# HELP fba_info FBA core build information\n# TYPE fba_info gauge\nfba_info 1\n")
	}
}
