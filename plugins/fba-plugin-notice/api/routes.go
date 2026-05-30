package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/notices/:pk", "Get notice", h.GetNotice, plugin.Auth()),
		plugin.GET("/sys/notices", "List notices", h.ListNotices, plugin.Auth()),
		plugin.POST("/sys/notices", "Create notice", h.CreateNotice, plugin.Auth(), plugin.Perm("sys:notice:add")),
		plugin.PUT("/sys/notices/:pk", "Update notice", h.UpdateNotice, plugin.Auth(), plugin.Perm("sys:notice:edit")),
		plugin.DELETE("/sys/notices", "Delete notices", h.DeleteNotices, plugin.Auth(), plugin.Perm("sys:notice:del")),
	}
}
