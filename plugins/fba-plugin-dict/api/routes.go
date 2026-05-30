package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/dict-types/all", "Get all dict types", h.GetAllDictTypes, plugin.Auth()),
		plugin.GET("/dict-types/:pk", "Get dict type", h.GetDictType, plugin.Auth()),
		plugin.GET("/dict-types", "List dict types", h.ListDictTypes, plugin.Auth()),
		plugin.POST("/dict-types", "Create dict type", h.CreateDictType, plugin.Auth(), plugin.Perm("dict:type:add")),
		plugin.PUT("/dict-types/:pk", "Update dict type", h.UpdateDictType, plugin.Auth(), plugin.Perm("dict:type:edit")),
		plugin.DELETE("/dict-types", "Delete dict types", h.DeleteDictTypes, plugin.Auth(), plugin.Perm("dict:type:del")),
		plugin.GET("/dict-datas/all", "Get all dict data", h.GetAllDictData, plugin.Auth()),
		plugin.GET("/dict-datas/:pk", "Get dict data", h.GetDictData, plugin.Auth()),
		plugin.GET("/dict-datas/type-codes/:code", "Get dict data by type code", h.GetDictDataByTypeCode, plugin.Auth()),
		plugin.GET("/dict-datas", "List dict data", h.ListDictData, plugin.Auth()),
		plugin.POST("/dict-datas", "Create dict data", h.CreateDictData, plugin.Auth(), plugin.Perm("dict:data:add")),
		plugin.PUT("/dict-datas/:pk", "Update dict data", h.UpdateDictData, plugin.Auth(), plugin.Perm("dict:data:edit")),
		plugin.DELETE("/dict-datas", "Delete dict data", h.DeleteDictData, plugin.Auth(), plugin.Perm("dict:data:del")),
	}
}
