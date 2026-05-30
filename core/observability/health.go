package observability

import "github.com/gofiber/fiber/v3"

func RegisterCoreRoutes(app *fiber.App, readiness *Readiness) {
	if readiness == nil {
		readiness = NewReadiness()
	}

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/readyz", func(c fiber.Ctx) error {
		result := readiness.Check(c.Context())
		if !result.Ready {
			return c.Status(fiber.StatusServiceUnavailable).JSON(result)
		}
		return c.JSON(result)
	})
	app.Get("/metrics", MetricsHandler())
}
