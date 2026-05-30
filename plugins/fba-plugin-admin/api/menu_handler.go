package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
)

func (Handler) SidebarMenus(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{
		{
			"id":        1,
			"name":      "Dashboard",
			"path":      "/dashboard",
			"parent_id": nil,
			"sort":      0,
			"type":      1,
			"component": "Layout",
			"perms":     nil,
			"remark":    nil,
			"children": []fiber.Map{
				{
					"id":        2,
					"name":      "Workbench",
					"path":      "/dashboard/workbench",
					"parent_id": 1,
					"sort":      0,
					"type":      1,
					"component": "/dashboard/workbench/index",
					"perms":     nil,
					"remark":    nil,
					"meta": fiber.Map{
						"title":                    "工作台",
						"icon":                     "lucide:layout-dashboard",
						"iframeSrc":                "",
						"link":                     "",
						"keepAlive":                true,
						"hideInMenu":               false,
						"menuVisibleWithForbidden": false,
					},
				},
			},
			"meta": fiber.Map{
				"title":                    "仪表盘",
				"icon":                     "lucide:layout-dashboard",
				"iframeSrc":                "",
				"link":                     "",
				"keepAlive":                true,
				"hideInMenu":               false,
				"menuVisibleWithForbidden": false,
			},
		},
	}))
}
