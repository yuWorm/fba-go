package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

func (h Handler) SidebarMenus(c fiber.Ctx) error {
	menus, err := h.menus.Sidebar(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(menus))
}
