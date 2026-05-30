package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

func (Handler) CurrentUser(c fiber.Ctx) error {
	return c.JSON(response.Success(currentUserInfo{
		userInfoDetail: fixtureUserInfo(),
		Dept:           nil,
		Roles:          []string{},
	}))
}
