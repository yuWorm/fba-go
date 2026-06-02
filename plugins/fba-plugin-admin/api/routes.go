package api

import "github.com/yuWorm/fba-go/core/plugin"

func AuthRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/auth/captcha", "Get captcha", h.Captcha),
		plugin.POST("/auth/login/swagger", "Swagger login", h.LoginSwagger),
		plugin.POST("/auth/login", "Login", h.Login),
		plugin.POST("/auth/refresh", "Refresh token", h.Refresh),
		plugin.POST("/auth/logout", "Logout", h.Logout),
		plugin.GET("/auth/codes", "Permission codes", h.Codes, plugin.Auth()),
	}
}

func UserRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/users/me", "Current user", h.CurrentUser, plugin.Auth()),
		plugin.GET("/sys/users/:pk", "Get user", h.GetUser, plugin.Auth()),
		plugin.GET("/sys/users/:pk/roles", "Get user roles", h.GetUserRoles, plugin.Auth()),
		plugin.GET("/sys/users", "List users", h.ListUsers, plugin.Auth()),
		plugin.POST("/sys/users", "Create user", h.CreateUser, plugin.Auth()),
		plugin.PUT("/sys/users/me/password", "Update current user password", h.UpdateCurrentUserPassword, plugin.Auth()),
		plugin.PUT("/sys/users/me/nickname", "Update current user nickname", h.UpdateCurrentUserNickname, plugin.Auth()),
		plugin.PUT("/sys/users/me/avatar", "Update current user avatar", h.UpdateCurrentUserAvatar, plugin.Auth()),
		plugin.PUT("/sys/users/me/email", "Update current user email", h.UpdateCurrentUserEmail, plugin.Auth()),
		plugin.PUT("/sys/users/:pk/permissions", "Update user permissions", h.UpdateUserPermission, plugin.Auth()),
		plugin.PUT("/sys/users/:pk/password", "Reset user password", h.ResetUserPassword, plugin.Auth()),
		plugin.PUT("/sys/users/:pk", "Update user", h.UpdateUser, plugin.Auth()),
		plugin.DELETE("/sys/users/:pk", "Delete user", h.DeleteUser, plugin.Auth(), plugin.Perm("sys:user:del")),
	}
}

func RoleRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/roles/all", "Get all roles", h.GetAllRoles, plugin.Auth()),
		plugin.GET("/sys/roles/:pk/menus", "Get role menus", h.GetRoleMenus, plugin.Auth()),
		plugin.GET("/sys/roles/:pk/scopes", "Get role scopes", h.GetRoleScopes, plugin.Auth()),
		plugin.GET("/sys/roles/:pk", "Get role", h.GetRole, plugin.Auth()),
		plugin.GET("/sys/roles", "List roles", h.ListRoles, plugin.Auth()),
		plugin.POST("/sys/roles", "Create role", h.CreateRole, plugin.Auth(), plugin.Perm("sys:role:add")),
		plugin.PUT("/sys/roles/:pk", "Update role", h.UpdateRole, plugin.Auth(), plugin.Perm("sys:role:edit")),
		plugin.PUT("/sys/roles/:pk/menus", "Update role menus", h.UpdateRoleMenus, plugin.Auth(), plugin.Perm("sys:role:menu:edit")),
		plugin.PUT("/sys/roles/:pk/scopes", "Update role scopes", h.UpdateRoleScopes, plugin.Auth(), plugin.Perm("sys:role:scope:edit")),
		plugin.DELETE("/sys/roles", "Delete roles", h.DeleteRoles, plugin.Auth(), plugin.Perm("sys:role:del")),
	}
}

func MenuRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/menus/sidebar", "Sidebar menus", h.SidebarMenus, plugin.Auth()),
		plugin.GET("/sys/menus/:pk", "Get menu", h.GetMenu, plugin.Auth()),
		plugin.GET("/sys/menus", "List menus", h.ListMenus, plugin.Auth()),
		plugin.POST("/sys/menus", "Create menu", h.CreateMenu, plugin.Auth(), plugin.Perm("sys:menu:add")),
		plugin.PUT("/sys/menus/:pk", "Update menu", h.UpdateMenu, plugin.Auth(), plugin.Perm("sys:menu:edit")),
		plugin.DELETE("/sys/menus/:pk", "Delete menu", h.DeleteMenu, plugin.Auth(), plugin.Perm("sys:menu:del")),
	}
}

func DeptRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/depts/:pk", "Get dept", h.GetDept, plugin.Auth()),
		plugin.GET("/sys/depts", "List depts", h.ListDepts, plugin.Auth()),
		plugin.POST("/sys/depts", "Create dept", h.CreateDept, plugin.Auth(), plugin.Perm("sys:dept:add")),
		plugin.PUT("/sys/depts/:pk", "Update dept", h.UpdateDept, plugin.Auth(), plugin.Perm("sys:dept:edit")),
		plugin.DELETE("/sys/depts/:pk", "Delete dept", h.DeleteDept, plugin.Auth(), plugin.Perm("sys:dept:del")),
	}
}

func DataRuleRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/data-rules/models", "Data rule models", h.DataRuleModels, plugin.Auth()),
		plugin.GET("/sys/data-rules/models/:model/columns", "Data rule model columns", h.DataRuleModelColumns, plugin.Auth()),
		plugin.GET("/sys/data-rules/value-template-variables", "Data rule value template variables", h.DataRuleValueTemplateVariables, plugin.Auth()),
		plugin.GET("/sys/data-rules/all", "Get all data rules", h.GetAllDataRules, plugin.Auth()),
		plugin.GET("/sys/data-rules/:pk", "Get data rule", h.GetDataRule, plugin.Auth()),
		plugin.GET("/sys/data-rules", "List data rules", h.ListDataRules, plugin.Auth()),
		plugin.POST("/sys/data-rules", "Create data rule", h.CreateDataRule, plugin.Auth(), plugin.Perm("data:rule:add")),
		plugin.PUT("/sys/data-rules/:pk", "Update data rule", h.UpdateDataRule, plugin.Auth(), plugin.Perm("data:rule:edit")),
		plugin.DELETE("/sys/data-rules", "Delete data rules", h.DeleteDataRules, plugin.Auth(), plugin.Perm("data:rule:del")),
	}
}

func DataScopeRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/data-scopes/all", "Get all data scopes", h.GetAllDataScopes, plugin.Auth()),
		plugin.GET("/sys/data-scopes/:pk/rules", "Get data scope rules", h.GetDataScopeRules, plugin.Auth()),
		plugin.GET("/sys/data-scopes/:pk", "Get data scope", h.GetDataScope, plugin.Auth()),
		plugin.GET("/sys/data-scopes", "List data scopes", h.ListDataScopes, plugin.Auth()),
		plugin.POST("/sys/data-scopes", "Create data scope", h.CreateDataScope, plugin.Auth(), plugin.Perm("data:scope:add")),
		plugin.PUT("/sys/data-scopes/:pk", "Update data scope", h.UpdateDataScope, plugin.Auth(), plugin.Perm("data:scope:edit")),
		plugin.PUT("/sys/data-scopes/:pk/rules", "Update data scope rules", h.UpdateDataScopeRules, plugin.Auth(), plugin.Perm("data:scope:rule:edit")),
		plugin.DELETE("/sys/data-scopes", "Delete data scopes", h.DeleteDataScopes, plugin.Auth(), plugin.Perm("data:scope:del")),
	}
}

func FileRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.POST("/sys/files/upload", "Upload file", h.UploadFile, plugin.Auth(), plugin.Perm("sys:file:upload")),
	}
}

func PluginRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/sys/plugins", "List plugins", h.ListPlugins, plugin.Auth()),
		plugin.GET("/sys/plugins/changed", "Plugin changed", h.PluginChanged, plugin.Auth()),
		plugin.POST("/sys/plugins", "Install plugin", h.InstallPlugin, plugin.Auth()),
		plugin.DELETE("/sys/plugins/:plugin", "Uninstall plugin", h.UninstallPlugin, plugin.Auth()),
		plugin.PUT("/sys/plugins/:plugin/status", "Update plugin status", h.UpdatePluginStatus, plugin.Auth()),
		plugin.GET("/sys/plugins/:plugin", "Download plugin", h.DownloadPlugin, plugin.Auth()),
	}
}

func LogRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/logs/login", "List login logs", h.ListLoginLogs, plugin.Auth()),
		plugin.DELETE("/logs/login", "Delete login logs", h.DeleteLoginLogs, plugin.Auth(), plugin.Perm("log:login:del")),
		plugin.DELETE("/logs/login/all", "Clear login logs", h.DeleteAllLoginLogs, plugin.Auth(), plugin.Perm("log:login:clear")),
		plugin.GET("/logs/opera", "List operation logs", h.ListOperaLogs, plugin.Auth()),
		plugin.DELETE("/logs/opera", "Delete operation logs", h.DeleteOperaLogs, plugin.Auth(), plugin.Perm("log:opera:del")),
		plugin.DELETE("/logs/opera/all", "Clear operation logs", h.DeleteAllOperaLogs, plugin.Auth(), plugin.Perm("log:opera:clear")),
	}
}

func MonitorRoutes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/monitors/server", "Server monitor", h.ServerMonitor, plugin.Auth()),
		plugin.GET("/monitors/redis", "Redis monitor", h.RedisMonitor, plugin.Auth()),
		plugin.GET("/monitors/sessions", "Online sessions", h.ListSessions, plugin.Auth()),
		plugin.DELETE("/monitors/sessions/:pk", "Delete online session", h.DeleteSession, plugin.Auth()),
	}
}
