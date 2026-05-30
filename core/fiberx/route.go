package fiberx

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
