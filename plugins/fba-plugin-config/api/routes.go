package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/configs/all", "Get all configs", h.GetAllConfigs, plugin.Auth()),
		plugin.GET("/sys/configs/:pk", "Get config", h.GetConfig, plugin.Auth()),
		plugin.GET("/sys/configs", "List configs", h.ListConfigs, plugin.Auth()),
		plugin.POST("/sys/configs", "Create config", h.CreateConfig, plugin.Auth(), plugin.Perm("sys:config:add")),
		plugin.PUT("/sys/configs", "Update configs", h.BulkUpdateConfigs, plugin.Auth(), plugin.Perm("sys.config.edits")),
		plugin.PUT("/sys/configs/:pk", "Update config", h.UpdateConfig, plugin.Auth(), plugin.Perm("sys:config:edit")),
		plugin.DELETE("/sys/configs", "Delete configs", h.DeleteConfigs, plugin.Auth(), plugin.Perm("sys:config:del")),
	}
}
