package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.POST("/emails/captcha", "Send email captcha", h.SendCaptcha, plugin.Auth()),
	}
}
