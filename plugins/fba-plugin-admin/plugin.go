package admin

import (
	"github.com/yuWorm/fba-go/core/plugin"
	adminapi "github.com/yuWorm/fba-plugin-admin/api"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "admin",
		Name:              "Admin Plugin",
		Version:           "0.1.0",
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	handler := adminapi.NewHandler()

	for _, route := range []plugin.Route{
		{Method: "GET", Path: "/auth/captcha", Summary: "Get captcha", Handler: handler.Captcha},
		{Method: "POST", Path: "/auth/login/swagger", Summary: "Swagger login", Handler: handler.LoginSwagger},
		{Method: "POST", Path: "/auth/login", Summary: "Login", Handler: handler.Login},
		{Method: "POST", Path: "/auth/refresh", Summary: "Refresh token", Handler: handler.Refresh},
		{Method: "POST", Path: "/auth/logout", Summary: "Logout", Handler: handler.Logout},
		{Method: "GET", Path: "/auth/codes", Summary: "Permission codes", AuthRequired: true, Handler: handler.Codes},
		{Method: "GET", Path: "/sys/users/me", Summary: "Current user", AuthRequired: true, Handler: handler.CurrentUser},
		{Method: "GET", Path: "/sys/users/:pk", Summary: "Get user", AuthRequired: true, Handler: handler.GetUser},
		{Method: "GET", Path: "/sys/users/:pk/roles", Summary: "Get user roles", AuthRequired: true, Handler: handler.GetUserRoles},
		{Method: "GET", Path: "/sys/users", Summary: "List users", AuthRequired: true, Handler: handler.ListUsers},
		{Method: "GET", Path: "/sys/roles/all", Summary: "Get all roles", AuthRequired: true, Handler: handler.GetAllRoles},
		{Method: "GET", Path: "/sys/roles/:pk/menus", Summary: "Get role menus", AuthRequired: true, Handler: handler.GetRoleMenus},
		{Method: "GET", Path: "/sys/roles/:pk/scopes", Summary: "Get role scopes", AuthRequired: true, Handler: handler.GetRoleScopes},
		{Method: "GET", Path: "/sys/roles/:pk", Summary: "Get role", AuthRequired: true, Handler: handler.GetRole},
		{Method: "GET", Path: "/sys/roles", Summary: "List roles", AuthRequired: true, Handler: handler.ListRoles},
		{Method: "POST", Path: "/sys/roles", Summary: "Create role", AuthRequired: true, Permission: "sys:role:add", Handler: handler.CreateRole},
		{Method: "PUT", Path: "/sys/roles/:pk", Summary: "Update role", AuthRequired: true, Permission: "sys:role:edit", Handler: handler.UpdateRole},
		{Method: "PUT", Path: "/sys/roles/:pk/menus", Summary: "Update role menus", AuthRequired: true, Permission: "sys:role:menu:edit", Handler: handler.UpdateRoleMenus},
		{Method: "PUT", Path: "/sys/roles/:pk/scopes", Summary: "Update role scopes", AuthRequired: true, Handler: handler.UpdateRoleScopes},
		{Method: "DELETE", Path: "/sys/roles", Summary: "Delete roles", AuthRequired: true, Permission: "sys:role:del", Handler: handler.DeleteRoles},
		{Method: "GET", Path: "/sys/menus/sidebar", Summary: "Sidebar menus", AuthRequired: true, Handler: handler.SidebarMenus},
		{Method: "GET", Path: "/sys/menus/:pk", Summary: "Get menu", AuthRequired: true, Handler: handler.GetMenu},
		{Method: "GET", Path: "/sys/menus", Summary: "List menus", AuthRequired: true, Handler: handler.ListMenus},
		{Method: "POST", Path: "/sys/menus", Summary: "Create menu", AuthRequired: true, Permission: "sys:menu:add", Handler: handler.CreateMenu},
		{Method: "PUT", Path: "/sys/menus/:pk", Summary: "Update menu", AuthRequired: true, Permission: "sys:menu:edit", Handler: handler.UpdateMenu},
		{Method: "DELETE", Path: "/sys/menus/:pk", Summary: "Delete menu", AuthRequired: true, Permission: "sys:menu:del", Handler: handler.DeleteMenu},
		{Method: "GET", Path: "/sys/depts/:pk", Summary: "Get dept", AuthRequired: true, Handler: handler.GetDept},
		{Method: "GET", Path: "/sys/depts", Summary: "List depts", AuthRequired: true, Handler: handler.ListDepts},
		{Method: "POST", Path: "/sys/depts", Summary: "Create dept", AuthRequired: true, Handler: handler.CreateDept},
		{Method: "PUT", Path: "/sys/depts/:pk", Summary: "Update dept", AuthRequired: true, Handler: handler.UpdateDept},
		{Method: "DELETE", Path: "/sys/depts/:pk", Summary: "Delete dept", AuthRequired: true, Handler: handler.DeleteDept},
		{Method: "GET", Path: "/sys/data-rules/models", Summary: "Data rule models", AuthRequired: true, Handler: handler.DataRuleModels},
		{Method: "GET", Path: "/sys/data-rules/models/:model/columns", Summary: "Data rule model columns", AuthRequired: true, Handler: handler.DataRuleModelColumns},
		{Method: "GET", Path: "/sys/data-rules/value-template-variables", Summary: "Data rule value template variables", AuthRequired: true, Handler: handler.DataRuleValueTemplateVariables},
		{Method: "GET", Path: "/sys/data-rules/all", Summary: "Get all data rules", AuthRequired: true, Handler: handler.GetAllDataRules},
		{Method: "GET", Path: "/sys/data-rules/:pk", Summary: "Get data rule", AuthRequired: true, Handler: handler.GetDataRule},
		{Method: "GET", Path: "/sys/data-rules", Summary: "List data rules", AuthRequired: true, Handler: handler.ListDataRules},
		{Method: "GET", Path: "/sys/data-scopes/all", Summary: "Get all data scopes", AuthRequired: true, Handler: handler.GetAllDataScopes},
		{Method: "GET", Path: "/sys/data-scopes/:pk/rules", Summary: "Get data scope rules", AuthRequired: true, Handler: handler.GetDataScopeRules},
		{Method: "GET", Path: "/sys/data-scopes/:pk", Summary: "Get data scope", AuthRequired: true, Handler: handler.GetDataScope},
		{Method: "GET", Path: "/sys/data-scopes", Summary: "List data scopes", AuthRequired: true, Handler: handler.ListDataScopes},
		{Method: "GET", Path: "/sys/plugins", Summary: "List plugins", AuthRequired: true, Handler: handler.ListPlugins},
		{Method: "GET", Path: "/sys/plugins/changed", Summary: "Plugin changed", AuthRequired: true, Handler: handler.PluginChanged},
		{Method: "GET", Path: "/sys/plugins/:plugin", Summary: "Download plugin", AuthRequired: true, Handler: handler.DownloadPlugin},
		{Method: "GET", Path: "/logs/login", Summary: "List login logs", AuthRequired: true, Handler: handler.ListLoginLogs},
		{Method: "GET", Path: "/logs/opera", Summary: "List operation logs", AuthRequired: true, Handler: handler.ListOperaLogs},
		{Method: "GET", Path: "/monitors/server", Summary: "Server monitor", AuthRequired: true, Handler: handler.ServerMonitor},
		{Method: "GET", Path: "/monitors/redis", Summary: "Redis monitor", AuthRequired: true, Handler: handler.RedisMonitor},
		{Method: "GET", Path: "/monitors/sessions", Summary: "Online sessions", AuthRequired: true, Handler: handler.ListSessions},
	} {
		if err := ctx.Route(route); err != nil {
			return err
		}
	}

	return nil
}
