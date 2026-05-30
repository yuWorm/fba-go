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

type MenuFilter struct {
	Title  string
	Status *int
}

type Repository interface {
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
}

func SeedData() model.Seed {
	return model.SeedData()
}
