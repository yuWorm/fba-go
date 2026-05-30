package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-go/core/response"
)

const defaultCurrentUserID = 1

func (h Handler) CurrentUser(c fiber.Ctx) error {
	user, err := h.users.Current(c.RequestCtx(), currentUserID(c))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(user))
}

func currentUserID(c fiber.Ctx) int {
	user, ok := c.Locals(plugin.CurrentUserLocalKey).(*rbac.CurrentUser)
	if !ok || user == nil || user.ID <= 0 {
		return defaultCurrentUserID
	}
	return int(user.ID)
}
