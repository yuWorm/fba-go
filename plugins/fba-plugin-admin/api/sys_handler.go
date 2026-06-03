package api

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

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
	if err := h.users.UpdatePermission(c.RequestCtx(), id, c.Query("type"), currentUserID(c)); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateCurrentUserPassword(c fiber.Ctx) error {
	var param dto.UserPasswordParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.UpdatePassword(c.RequestCtx(), currentUserID(c), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) ResetUserPassword(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.UserResetPasswordParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.ResetPassword(c.RequestCtx(), id, param.Password); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateCurrentUserNickname(c fiber.Ctx) error {
	var param dto.UserNicknameParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.UpdateNickname(c.RequestCtx(), currentUserID(c), param.Nickname); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateCurrentUserAvatar(c fiber.Ctx) error {
	var param dto.UserAvatarParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.UpdateAvatar(c.RequestCtx(), currentUserID(c), param.Avatar); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateCurrentUserEmail(c fiber.Ctx) error {
	var param dto.UserEmailParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.users.UpdateEmail(c.RequestCtx(), currentUserID(c), param.Captcha, param.Email); err != nil {
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
	depts, err := h.depts.TreeForUser(c.RequestCtx(), repo.DeptFilter{
		Name:   c.Query("name"),
		Leader: c.Query("leader"),
		Phone:  c.Query("phone"),
		Status: intPtrQuery(c, "status"),
	}, currentUser(c))
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

func (h Handler) DataRuleModels(c fiber.Ctx) error {
	models, err := h.dataRules.Models(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(models))
}

func (h Handler) DataRuleModelColumns(c fiber.Ctx) error {
	columns, err := h.dataRules.Columns(c.RequestCtx(), c.Params("model"))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(columns))
}

func (h Handler) DataRuleValueTemplateVariables(c fiber.Ctx) error {
	variables, err := h.dataRules.ValueTemplateVariables(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(variables))
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

func (h Handler) UploadFile(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}
	uploaded, err := h.files.Upload(c.RequestCtx(), file.Filename, file.Size)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(uploaded))
}

func (h Handler) ListPlugins(c fiber.Ctx) error {
	plugins, err := h.plugins.All(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(plugins))
}

func (h Handler) PluginChanged(c fiber.Ctx) error {
	changed, err := h.plugins.Changed(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(changed))
}

func (h Handler) InstallPlugin(c fiber.Ctx) error {
	return h.plugins.Install(c.RequestCtx(), c.Query("type"), c.Query("repo_url"))
}

func (h Handler) UninstallPlugin(c fiber.Ctx) error {
	pluginName := c.Params("plugin")
	if err := h.plugins.Uninstall(c.RequestCtx(), pluginName); err != nil {
		return err
	}
	return c.JSON(response.Response[any]{
		Code: 200,
		Msg:  fmt.Sprintf("插件 %s 卸载成功，请根据插件说明（README.md）移除相关配置并重启服务", pluginName),
		Data: nil,
	})
}

func (h Handler) UpdatePluginStatus(c fiber.Ctx) error {
	if err := h.plugins.ToggleStatus(c.RequestCtx(), c.Params("plugin")); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DownloadPlugin(c fiber.Ctx) error {
	body, err := h.plugins.Download(c.RequestCtx(), c.Params("plugin"))
	if err != nil {
		return err
	}
	return c.SendString(body)
}

func (h Handler) ListLoginLogs(c fiber.Ctx) error {
	page, size := pageParams(c)
	logs, err := h.logs.ListLogin(c.RequestCtx(), repo.LogFilter{
		Username: c.Query("username"),
		Status:   intPtrQuery(c, "status"),
		IP:       c.Query("ip"),
	}, page, size, "/api/v1/logs/login")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(logs))
}

func (h Handler) DeleteLoginLogs(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	count, err := h.logs.DeleteLogin(c.RequestCtx(), param.PKs)
	if err != nil {
		return err
	}
	// Python returns a business failure envelope, not an HTTP error, when no log rows are deleted.
	if count == 0 {
		return c.JSON(response.Fail[any](nil))
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteAllLoginLogs(c fiber.Ctx) error {
	if err := h.logs.ClearLogin(c.RequestCtx()); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) ListOperaLogs(c fiber.Ctx) error {
	page, size := pageParams(c)
	logs, err := h.logs.ListOpera(c.RequestCtx(), repo.LogFilter{
		Username: c.Query("username"),
		Status:   intPtrQuery(c, "status"),
		IP:       c.Query("ip"),
	}, page, size, "/api/v1/logs/opera")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(logs))
}

func (h Handler) DeleteOperaLogs(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	count, err := h.logs.DeleteOpera(c.RequestCtx(), param.PKs)
	if err != nil {
		return err
	}
	// Python returns a business failure envelope, not an HTTP error, when no log rows are deleted.
	if count == 0 {
		return c.JSON(response.Fail[any](nil))
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteAllOperaLogs(c fiber.Ctx) error {
	if err := h.logs.ClearOpera(c.RequestCtx()); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) ServerMonitor(c fiber.Ctx) error {
	monitor, err := h.monitors.Server(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(monitor))
}

func (h Handler) RedisMonitor(c fiber.Ctx) error {
	monitor, err := h.monitors.Redis(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(monitor))
}

func (h Handler) ListSessions(c fiber.Ctx) error {
	sessions, err := h.monitors.Sessions(c.RequestCtx(), c.Query("username"))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(sessions))
}

func (h Handler) DeleteSession(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.monitors.DeleteSession(c.RequestCtx(), id, c.Query("session_uuid")); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
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
