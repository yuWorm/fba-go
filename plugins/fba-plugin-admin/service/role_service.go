package service

import (
	"context"
	stderrors "errors"
	"net/http"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
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
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.RoleWithRelationDetail{}, roleNotFound("角色不存在", err)
		}
		return dto.RoleWithRelationDetail{}, err
	}
	menus, err := s.repo.RoleMenus(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.RoleWithRelationDetail{}, roleNotFound("角色不存在", err)
		}
		return dto.RoleWithRelationDetail{}, err
	}
	scopes, err := s.repo.RoleScopes(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.RoleWithRelationDetail{}, roleNotFound("角色不存在", err)
		}
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
	if _, err := s.repo.GetRoleByName(ctx, param.Name); err == nil {
		return roleConflict("角色已存在", nil)
	} else if !stderrors.Is(err, repo.ErrNotFound) {
		return err
	}
	return s.repo.CreateRole(ctx, param)
}

func (s *RoleService) Update(ctx context.Context, id int, param dto.RoleParam) error {
	role, err := s.ensureRole(ctx, id)
	if err != nil {
		return err
	}
	if role.Name != param.Name {
		if _, err := s.repo.GetRoleByName(ctx, param.Name); err == nil {
			return roleConflict("角色已存在", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	return s.repo.UpdateRole(ctx, id, param)
}

func (s *RoleService) Delete(ctx context.Context, ids []int) error {
	return s.repo.DeleteRoles(ctx, ids)
}

func (s *RoleService) MenuTree(ctx context.Context, roleID int) ([]dto.MenuDetail, error) {
	if _, err := s.ensureRole(ctx, roleID); err != nil {
		return nil, err
	}
	menus, err := s.repo.RoleMenus(ctx, roleID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return nil, roleNotFound("角色不存在", err)
		}
		return nil, err
	}
	return dto.MenusFromModel(menus), nil
}

func (s *RoleService) Scopes(ctx context.Context, roleID int) ([]int, error) {
	if _, err := s.ensureRole(ctx, roleID); err != nil {
		return nil, err
	}
	scopeIDs, err := s.repo.RoleScopeIDs(ctx, roleID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return nil, roleNotFound("角色不存在", err)
		}
		return nil, err
	}
	return scopeIDs, nil
}

func (s *RoleService) UpdateMenus(ctx context.Context, roleID int, menuIDs []int) error {
	if _, err := s.ensureRole(ctx, roleID); err != nil {
		return err
	}
	// Python role relation writes are strict: any unknown menu ID fails the request instead of being silently ignored.
	if err := s.ensureMenus(ctx, menuIDs); err != nil {
		return err
	}
	return s.repo.UpdateRoleMenus(ctx, roleID, menuIDs)
}

func (s *RoleService) UpdateScopes(ctx context.Context, roleID int, scopeIDs []int) error {
	if _, err := s.ensureRole(ctx, roleID); err != nil {
		return err
	}
	// Keep data-scope assignment aligned with Python's all-or-error behavior.
	if err := s.ensureScopes(ctx, scopeIDs); err != nil {
		return err
	}
	return s.repo.UpdateRoleScopes(ctx, roleID, scopeIDs)
}

func (s *RoleService) ensureRole(ctx context.Context, id int) (model.Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return model.Role{}, roleNotFound("角色不存在", err)
		}
		return model.Role{}, err
	}
	return role, nil
}

func (s *RoleService) ensureMenus(ctx context.Context, ids []int) error {
	for _, id := range uniqueIDs(ids) {
		if _, err := s.repo.GetMenu(ctx, id); err != nil {
			if stderrors.Is(err, repo.ErrNotFound) {
				return roleNotFound("菜单不存在", err)
			}
			return err
		}
	}
	return nil
}

func (s *RoleService) ensureScopes(ctx context.Context, ids []int) error {
	for _, id := range uniqueIDs(ids) {
		if _, err := s.repo.GetDataScope(ctx, id); err != nil {
			if stderrors.Is(err, repo.ErrNotFound) {
				return roleNotFound("数据范围不存在", err)
			}
			return err
		}
	}
	return nil
}

func uniqueIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func roleNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func roleConflict(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}
