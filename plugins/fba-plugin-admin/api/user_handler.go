package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

func (Handler) CurrentUser(c fiber.Ctx) error {
	return c.JSON(response.Success(fiber.Map{
		"id":           1,
		"username":     "admin",
		"nickname":     "Admin",
		"is_superuser": true,
		"is_staff":     true,
		"roles":        []fiber.Map{},
		"menus":        []fiber.Map{},
		"depts":        []fiber.Map{},
	}))
}
