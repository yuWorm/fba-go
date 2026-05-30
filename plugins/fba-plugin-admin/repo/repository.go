package repo

import (
	"context"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
)

type RoleFilter struct {
	Name   string
	Status *int
}

type UserFilter struct {
	Dept     *int
	Username string
	Phone    string
	Status   *int
}

type MenuFilter struct {
	Title  string
	Status *int
}

type DeptFilter struct {
	Name   string
	Leader string
	Phone  string
	Status *int
}

type DataRuleFilter struct {
	Name string
}

type DataScopeFilter struct {
	Name   string
	Status *int
}

type Repository interface {
	GetUser(ctx context.Context, id int) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	ListUsers(ctx context.Context, filter UserFilter, page int, size int) ([]model.User, int64, error)
	CreateUser(ctx context.Context, param dto.UserCreateParam) (model.User, error)
	UpdateUser(ctx context.Context, id int, param dto.UserUpdateParam) error
	UpdateUserNickname(ctx context.Context, id int, nickname string) error
	UpdateUserAvatar(ctx context.Context, id int, avatar *string) error
	UpdateUserEmail(ctx context.Context, id int, email *string) error
	ResetUserPassword(ctx context.Context, id int, password string) error
	UpdateUserPermission(ctx context.Context, id int, permissionType string) error
	DeleteUser(ctx context.Context, id int) error
	UserRoles(ctx context.Context, userID int) ([]model.Role, error)
	AllRoles(ctx context.Context) ([]model.Role, error)
	GetRole(ctx context.Context, id int) (model.Role, error)
	ListRoles(ctx context.Context, filter RoleFilter, page int, size int) ([]model.Role, int64, error)
	CreateRole(ctx context.Context, param dto.RoleParam) error
	UpdateRole(ctx context.Context, id int, param dto.RoleParam) error
	DeleteRoles(ctx context.Context, ids []int) error
	RoleMenus(ctx context.Context, roleID int) ([]model.Menu, error)
	UpdateRoleMenus(ctx context.Context, roleID int, menuIDs []int) error
	RoleScopes(ctx context.Context, roleID int) ([]model.DataScope, error)
	RoleScopeIDs(ctx context.Context, roleID int) ([]int, error)
	UpdateRoleScopes(ctx context.Context, roleID int, scopeIDs []int) error
	GetMenu(ctx context.Context, id int) (model.Menu, error)
	ListMenus(ctx context.Context, filter MenuFilter) ([]model.Menu, error)
	SidebarMenus(ctx context.Context) ([]model.Menu, error)
	CreateMenu(ctx context.Context, param dto.MenuParam) error
	UpdateMenu(ctx context.Context, id int, param dto.MenuParam) error
	DeleteMenu(ctx context.Context, id int) error
	GetDept(ctx context.Context, id int) (model.Dept, error)
	ListDepts(ctx context.Context, filter DeptFilter) ([]model.Dept, error)
	CreateDept(ctx context.Context, param dto.DeptParam) error
	UpdateDept(ctx context.Context, id int, param dto.DeptParam) error
	DeleteDept(ctx context.Context, id int) error
	AllDataRules(ctx context.Context) ([]model.DataRule, error)
	GetDataRule(ctx context.Context, id int) (model.DataRule, error)
	ListDataRules(ctx context.Context, filter DataRuleFilter, page int, size int) ([]model.DataRule, int64, error)
	CreateDataRule(ctx context.Context, param dto.DataRuleParam) error
	UpdateDataRule(ctx context.Context, id int, param dto.DataRuleParam) error
	DeleteDataRules(ctx context.Context, ids []int) error
	AllDataScopes(ctx context.Context) ([]model.DataScope, error)
	GetDataScope(ctx context.Context, id int) (model.DataScope, error)
	DataScopeRules(ctx context.Context, id int) (model.DataScope, []model.DataRule, error)
	ListDataScopes(ctx context.Context, filter DataScopeFilter, page int, size int) ([]model.DataScope, int64, error)
	CreateDataScope(ctx context.Context, param dto.DataScopeParam) error
	UpdateDataScope(ctx context.Context, id int, param dto.DataScopeParam) error
	UpdateDataScopeRules(ctx context.Context, id int, ruleIDs []int) error
	DeleteDataScopes(ctx context.Context, ids []int) error
}

func SeedData() model.Seed {
	return model.SeedData()
}
