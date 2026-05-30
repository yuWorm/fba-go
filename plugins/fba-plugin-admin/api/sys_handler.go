package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

const fixtureTime = "2026-05-30 00:00:00"

func (h Handler) GetUser(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	user, err := h.users.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(user))
}

func (h Handler) GetUserRoles(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	roles, err := h.users.Roles(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(roles))
}

func (h Handler) ListUsers(c fiber.Ctx) error {
	page, size := pageParams(c)
	users, err := h.users.List(c.RequestCtx(), repo.UserFilter{
		Dept:     intPtrQuery(c, "dept"),
		Username: c.Query("username"),
		Phone:    c.Query("phone"),
		Status:   intPtrQuery(c, "status"),
	}, page, size, "/api/v1/sys/users")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(users))
}

func (h Handler) CreateUser(c fiber.Ctx) error {
	var param dto.UserCreateParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	user, err := h.users.Create(c.RequestCtx(), param)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(user))
}

func (h Handler) UpdateUser(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.UserUpdateParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateUserPermission(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.users.UpdatePermission(c.RequestCtx(), id, c.Query("type")); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateCurrentUserPassword(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) ResetUserPassword(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateCurrentUserNickname(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateCurrentUserAvatar(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateCurrentUserEmail(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteUser(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.users.Delete(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) GetAllRoles(c fiber.Ctx) error {
	roles, err := h.roles.All(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(roles))
}

func (h Handler) GetRoleMenus(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	menus, err := h.roles.MenuTree(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(menus))
}

func (h Handler) GetRoleScopes(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	scopes, err := h.roles.Scopes(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(scopes))
}

func (h Handler) GetRole(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	role, err := h.roles.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(role))
}

func (h Handler) ListRoles(c fiber.Ctx) error {
	page, size := pageParams(c)
	roles, err := h.roles.List(c.RequestCtx(), repo.RoleFilter{
		Name:   c.Query("name"),
		Status: intPtrQuery(c, "status"),
	}, page, size, "/api/v1/sys/roles")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(roles))
}

func (h Handler) CreateRole(c fiber.Ctx) error {
	var param dto.RoleParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.roles.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateRole(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.RoleParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.roles.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateRoleMenus(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.RoleMenuParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.roles.UpdateMenus(c.RequestCtx(), id, param.Menus); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateRoleScopes(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.RoleScopeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.roles.UpdateScopes(c.RequestCtx(), id, param.Scopes); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteRoles(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.roles.Delete(c.RequestCtx(), param.PKs); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) GetMenu(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	menu, err := h.menus.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(menu))
}

func (h Handler) ListMenus(c fiber.Ctx) error {
	menus, err := h.menus.Tree(c.RequestCtx(), repo.MenuFilter{
		Title:  c.Query("title"),
		Status: intPtrQuery(c, "status"),
	})
	if err != nil {
		return err
	}
	return c.JSON(response.Success(menus))
}

func (h Handler) CreateMenu(c fiber.Ctx) error {
	var param dto.MenuParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.menus.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateMenu(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.MenuParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.menus.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteMenu(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.menus.Delete(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) GetDept(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	dept, err := h.depts.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(dept))
}

func (h Handler) ListDepts(c fiber.Ctx) error {
	depts, err := h.depts.Tree(c.RequestCtx(), repo.DeptFilter{
		Name:   c.Query("name"),
		Leader: c.Query("leader"),
		Phone:  c.Query("phone"),
		Status: intPtrQuery(c, "status"),
	})
	if err != nil {
		return err
	}
	return c.JSON(response.Success(depts))
}

func (h Handler) CreateDept(c fiber.Ctx) error {
	var param dto.DeptParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.depts.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDept(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DeptParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.depts.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteDept(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.depts.Delete(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) DataRuleModels(c fiber.Ctx) error {
	return c.JSON(response.Success([]string{"user", "role", "dept"}))
}

func (Handler) DataRuleModelColumns(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{
		{"key": "id", "comment": "ID"},
		{"key": "dept_id", "comment": "部门 ID"},
	}))
}

func (Handler) DataRuleValueTemplateVariables(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{
		{"key": "user_id", "comment": "当前用户 ID"},
		{"key": "dept_id", "comment": "当前部门 ID"},
	}))
}

func (h Handler) GetAllDataRules(c fiber.Ctx) error {
	rules, err := h.dataRules.All(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(rules))
}

func (h Handler) GetDataRule(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	rule, err := h.dataRules.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(rule))
}

func (h Handler) ListDataRules(c fiber.Ctx) error {
	page, size := pageParams(c)
	rules, err := h.dataRules.List(c.RequestCtx(), repo.DataRuleFilter{
		Name: c.Query("name"),
	}, page, size, "/api/v1/sys/data-rules")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(rules))
}

func (h Handler) CreateDataRule(c fiber.Ctx) error {
	var param dto.DataRuleParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataRules.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDataRule(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DataRuleParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataRules.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteDataRules(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataRules.Delete(c.RequestCtx(), param.PKs); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) GetAllDataScopes(c fiber.Ctx) error {
	scopes, err := h.dataScopes.All(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(scopes))
}

func (h Handler) GetDataScope(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	scope, err := h.dataScopes.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(scope))
}

func (h Handler) GetDataScopeRules(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	scope, err := h.dataScopes.Rules(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(scope))
}

func (h Handler) ListDataScopes(c fiber.Ctx) error {
	page, size := pageParams(c)
	scopes, err := h.dataScopes.List(c.RequestCtx(), repo.DataScopeFilter{
		Name:   c.Query("name"),
		Status: intPtrQuery(c, "status"),
	}, page, size, "/api/v1/sys/data-scopes")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(scopes))
}

func (h Handler) CreateDataScope(c fiber.Ctx) error {
	var param dto.DataScopeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataScopes.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDataScope(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DataScopeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataScopes.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDataScopeRules(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DataScopeRuleParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataScopes.UpdateRules(c.RequestCtx(), id, param.Rules); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteDataScopes(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.dataScopes.Delete(c.RequestCtx(), param.PKs); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) UploadFile(c fiber.Ctx) error {
	return c.JSON(response.Success(fiber.Map{"url": "/static/upload/contract.txt"}))
}

func (Handler) ListPlugins(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{fixturePlugin()}))
}

func (Handler) PluginChanged(c fiber.Ctx) error {
	return c.JSON(response.Success(false))
}

func (Handler) InstallPlugin(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UninstallPlugin(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdatePluginStatus(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) DownloadPlugin(c fiber.Ctx) error {
	return c.SendString("plugin fixture")
}

func (Handler) ListLoginLogs(c fiber.Ctx) error {
	return c.JSON(response.Success(pagination.NewPageData([]fiber.Map{}, 0, 1, 20, "/api/v1/logs/login")))
}

func (Handler) DeleteLoginLogs(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) DeleteAllLoginLogs(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) ListOperaLogs(c fiber.Ctx) error {
	return c.JSON(response.Success(pagination.NewPageData([]fiber.Map{}, 0, 1, 20, "/api/v1/logs/opera")))
}

func (Handler) DeleteOperaLogs(c fiber.Ctx) error {
	if err := bindBody(c); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (Handler) DeleteAllOperaLogs(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) ServerMonitor(c fiber.Ctx) error {
	return c.JSON(response.Success(fiber.Map{
		"cpu": fiber.Map{
			"physical_num": 1,
			"logical_num":  1,
			"max_freq":     0,
			"min_freq":     0,
			"current_freq": 0,
			"usage":        0,
		},
		"mem": fiber.Map{
			"total": 0,
			"used":  0,
			"free":  0,
			"usage": 0,
		},
		"sys": fiber.Map{
			"name": "fba-go",
			"os":   "go",
			"ip":   "127.0.0.1",
			"arch": "unknown",
		},
		"disk": []fiber.Map{},
		"service": fiber.Map{
			"name":      "fba-go",
			"version":   "0.1.0",
			"home":      "",
			"startup":   fixtureTime,
			"elapsed":   "0s",
			"cpu_usage": "0%",
			"mem_vms":   "0B",
			"mem_rss":   "0B",
			"mem_free":  "0B",
		},
	}))
}

func (Handler) RedisMonitor(c fiber.Ctx) error {
	return c.JSON(response.Success(fiber.Map{
		"info": fiber.Map{
			"redis_version":             "",
			"redis_mode":                "",
			"role":                      "",
			"tcp_port":                  "",
			"uptime":                    "0s",
			"connected_clients":         "0",
			"blocked_clients":           "0",
			"used_memory_human":         "0B",
			"used_memory_rss_human":     "0B",
			"maxmemory_human":           "0B",
			"mem_fragmentation_ratio":   "0",
			"instantaneous_ops_per_sec": "0",
			"total_commands_processed":  "0",
			"rejected_connections":      "0",
			"keys_num":                  "0",
		},
		"stats": []fiber.Map{},
	}))
}

func (Handler) ListSessions(c fiber.Ctx) error {
	return c.JSON(response.Success([]fiber.Map{}))
}

func (Handler) DeleteSession(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func bindBody(c fiber.Ctx) error {
	var body map[string]any
	return c.Bind().Body(&body)
}

func parseID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	return id, nil
}

func pageParams(c fiber.Ctx) (int, int) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(c.Query("size", "20"))
	if err != nil || size < 1 {
		size = 20
	}
	return page, size
}

func intPtrQuery(c fiber.Ctx, name string) *int {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}

func fixtureUser() fiber.Map {
	return fiber.Map{
		"dept_id":         nil,
		"username":        "admin",
		"nickname":        "Admin",
		"avatar":          nil,
		"email":           nil,
		"phone":           nil,
		"id":              1,
		"uuid":            "fixture-user",
		"status":          1,
		"is_superuser":    true,
		"is_staff":        true,
		"is_multi_login":  true,
		"join_time":       fixtureTime,
		"last_login_time": nil,
		"dept":            nil,
		"roles":           []fiber.Map{fixtureRole()},
	}
}

func fixtureRole() fiber.Map {
	return fiber.Map{
		"name":             "admin",
		"status":           1,
		"is_filter_scopes": true,
		"remark":           nil,
		"id":               1,
		"created_time":     fixtureTime,
		"updated_time":     nil,
	}
}

func fixtureMenu() fiber.Map {
	return fiber.Map{
		"title":        "仪表盘",
		"name":         "Dashboard",
		"path":         "/dashboard",
		"parent_id":    nil,
		"sort":         0,
		"icon":         "lucide:layout-dashboard",
		"type":         1,
		"component":    "Layout",
		"perms":        nil,
		"status":       1,
		"display":      1,
		"cache":        1,
		"link":         nil,
		"remark":       nil,
		"id":           1,
		"created_time": fixtureTime,
		"updated_time": nil,
		"children":     []fiber.Map{},
	}
}

func fixtureDept() fiber.Map {
	return fiber.Map{
		"name":         "总部",
		"parent_id":    nil,
		"sort":         0,
		"leader":       nil,
		"phone":        nil,
		"email":        nil,
		"status":       1,
		"id":           1,
		"deleted":      0,
		"created_time": fixtureTime,
		"updated_time": nil,
		"deleted_time": nil,
		"children":     []fiber.Map{},
	}
}

func fixtureDataRule() fiber.Map {
	return fiber.Map{
		"name":         "本人数据",
		"model":        "user",
		"column":       "id",
		"operator":     0,
		"expression":   0,
		"value":        "{{ user_id }}",
		"id":           1,
		"created_time": fixtureTime,
		"updated_time": nil,
	}
}

func fixtureDataScope() fiber.Map {
	return fiber.Map{
		"name":         "本人数据范围",
		"status":       1,
		"id":           1,
		"created_time": fixtureTime,
		"updated_time": nil,
	}
}

func fixturePlugin() fiber.Map {
	return fiber.Map{
		"name":        "dict",
		"version":     "0.0.8",
		"description": "Dictionary data plugin",
		"status":      true,
	}
}
