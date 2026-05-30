package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

const refreshCookieName = "fba_refresh_token"

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func (Handler) Captcha(c fiber.Ctx) error {
	return c.JSON(response.Success(fiber.Map{
		"uuid": "fixture-captcha",
		"img":  "",
	}))
}

func (Handler) Login(c fiber.Ctx) error {
	setRefreshCookie(c)
	return c.JSON(response.Success(fiber.Map{
		"access_token":                   "fixture-access-token",
		"access_token_expire_time":       "2099-01-01 00:00:00",
		"session_uuid":                   "fixture-session",
		"password_expire_days_remaining": nil,
		"user": fiber.Map{
			"id":       1,
			"username": "admin",
			"nickname": "Admin",
		},
	}))
}

func (Handler) Refresh(c fiber.Ctx) error {
	setRefreshCookie(c)
	return c.JSON(response.Success(fiber.Map{
		"access_token":             "fixture-access-token-refreshed",
		"access_token_expire_time": "2099-01-01 00:00:00",
		"session_uuid":             "fixture-session",
	}))
}

func (Handler) Logout(c fiber.Ctx) error {
	c.ClearCookie(refreshCookieName)
	return c.JSON(response.Success[any](nil))
}

func (Handler) Codes(c fiber.Ctx) error {
	return c.JSON(response.Success([]string{
		"sys:user:view",
		"sys:menu:view",
	}))
}

func setRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "fixture-refresh-token",
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  time.Now().Add(24 * time.Hour),
	})
}
