package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

// Until real JWT user context is wired into the plugin host, the seeded admin user is the current user.
const currentUserID = 1

func (h Handler) CurrentUser(c fiber.Ctx) error {
	user, err := h.users.Current(c.RequestCtx(), currentUserID)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(user))
}
