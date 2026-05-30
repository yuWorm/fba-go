package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

const refreshCookieName = "fba_refresh_token"
const refreshCookieMaxAgeSeconds = 60 * 60 * 24 * 7

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

type captchaDetail struct {
	IsEnabled     bool   `json:"is_enabled"`
	ExpireSeconds int    `json:"expire_seconds"`
	UUID          string `json:"uuid"`
	Image         string `json:"image"`
}

type userInfoDetail struct {
	DeptID        *int    `json:"dept_id"`
	Username      string  `json:"username"`
	Nickname      string  `json:"nickname"`
	Avatar        *string `json:"avatar"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	ID            int     `json:"id"`
	UUID          string  `json:"uuid"`
	Status        int     `json:"status"`
	IsSuperuser   bool    `json:"is_superuser"`
	IsStaff       bool    `json:"is_staff"`
	IsMultiLogin  bool    `json:"is_multi_login"`
	JoinTime      string  `json:"join_time"`
	LastLoginTime *string `json:"last_login_time"`
}

type currentUserInfo struct {
	userInfoDetail
	Dept  *string  `json:"dept"`
	Roles []string `json:"roles"`
}

type swaggerToken struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	User        userInfoDetail `json:"user"`
}

type accessTokenBase struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpireTime string `json:"access_token_expire_time"`
	SessionUUID           string `json:"session_uuid"`
}

type loginToken struct {
	accessTokenBase
	PasswordExpireDaysRemaining *int           `json:"password_expire_days_remaining"`
	User                        userInfoDetail `json:"user"`
}

func (Handler) Captcha(c fiber.Ctx) error {
	return c.JSON(response.Success(captchaDetail{
		IsEnabled:     true,
		ExpireSeconds: 300,
		UUID:          "fixture-captcha",
		Image:         "",
	}))
}

func (Handler) LoginSwagger(c fiber.Ctx) error {
	return c.JSON(swaggerToken{
		AccessToken: "fixture-access-token",
		TokenType:   "Bearer",
		User:        fixtureUserInfo(),
	})
}

func (Handler) Login(c fiber.Ctx) error {
	setRefreshCookie(c)
	return c.JSON(response.Success(loginToken{
		accessTokenBase: accessTokenBase{
			AccessToken:           "fixture-access-token",
			AccessTokenExpireTime: "2099-01-01 00:00:00",
			SessionUUID:           "fixture-session",
		},
		PasswordExpireDaysRemaining: nil,
		User:                        fixtureUserInfo(),
	}))
}

func (Handler) Refresh(c fiber.Ctx) error {
	setRefreshCookie(c)
	return c.JSON(response.Success(accessTokenBase{
		AccessToken:           "fixture-access-token-refreshed",
		AccessTokenExpireTime: "2099-01-01 00:00:00",
		SessionUUID:           "fixture-session",
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
		MaxAge:   refreshCookieMaxAgeSeconds,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func fixtureUserInfo() userInfoDetail {
	return userInfoDetail{
		DeptID:        nil,
		Username:      "admin",
		Nickname:      "Admin",
		Avatar:        nil,
		Email:         nil,
		Phone:         nil,
		ID:            1,
		UUID:          "fixture-user",
		Status:        1,
		IsSuperuser:   true,
		IsStaff:       true,
		IsMultiLogin:  true,
		JoinTime:      "2026-05-30 00:00:00",
		LastLoginTime: nil,
	}
}
