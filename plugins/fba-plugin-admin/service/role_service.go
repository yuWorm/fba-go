package service

import (
	"context"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type RoleService struct {
	repo repo.Repository
}

func NewRoleService(repository repo.Repository) *RoleService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &RoleService{repo: repository}
}

func (s *RoleService) All(ctx context.Context) ([]dto.RoleDetail, error) {
	items, err := s.repo.AllRoles(ctx)
	if err != nil {
		return nil, err
	}
	return dto.RolesFromModel(items), nil
}

func (s *RoleService) Get(ctx context.Context, id int) (dto.RoleWithRelationDetail, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return dto.RoleWithRelationDetail{}, err
	}
	menus, err := s.repo.RoleMenus(ctx, id)
	if err != nil {
		return dto.RoleWithRelationDetail{}, err
	}
	scopes, err := s.repo.RoleScopes(ctx, id)
	if err != nil {
		return dto.RoleWithRelationDetail{}, err
	}
	return dto.RoleWithRelations(role, menus, scopes), nil
}

func (s *RoleService) List(ctx context.Context, filter repo.RoleFilter, page int, size int, basePath string) (pagination.PageData[dto.RoleDetail], error) {
	items, total, err := s.repo.ListRoles(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.RoleDetail]{}, err
	}
	return pagination.NewPageData(dto.RolesFromModel(items), total, page, size, basePath), nil
}

func (s *RoleService) Create(ctx context.Context, param dto.RoleParam) error {
	return s.repo.CreateRole(ctx, param)
}

func (s *RoleService) Update(ctx context.Context, id int, param dto.RoleParam) error {
	return s.repo.UpdateRole(ctx, id, param)
}

func (s *RoleService) Delete(ctx context.Context, ids []int) error {
	return s.repo.DeleteRoles(ctx, ids)
}

func (s *RoleService) MenuTree(ctx context.Context, roleID int) ([]dto.MenuDetail, error) {
	menus, err := s.repo.RoleMenus(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return dto.MenusFromModel(menus), nil
}

func (s *RoleService) Scopes(ctx context.Context, roleID int) ([]int, error) {
	return s.repo.RoleScopeIDs(ctx, roleID)
}

func (s *RoleService) UpdateMenus(ctx context.Context, roleID int, menuIDs []int) error {
	return s.repo.UpdateRoleMenus(ctx, roleID, menuIDs)
}

func (s *RoleService) UpdateScopes(ctx context.Context, roleID int, scopeIDs []int) error {
	return s.repo.UpdateRoleScopes(ctx, roleID, scopeIDs)
}
