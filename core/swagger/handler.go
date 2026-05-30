package swagger

import "github.com/gofiber/fiber/v3"

func OpenAPIJSONHandler(doc Document) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.JSON(doc)
	}
}

func UIHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(`<!doctype html><html><head><title>FBA API Docs</title></head><body><redoc spec-url="/openapi"></redoc><script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script></body></html>`)
	}
}

func RegisterHandlers(app *fiber.App, doc Document) {
	app.Get("/openapi", OpenAPIJSONHandler(doc))
	app.Get("/swagger/doc.json", OpenAPIJSONHandler(doc))
	app.Get("/docs", UIHandler())
}
