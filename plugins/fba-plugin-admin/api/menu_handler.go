package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

func (Handler) SidebarMenus(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{}))
}
