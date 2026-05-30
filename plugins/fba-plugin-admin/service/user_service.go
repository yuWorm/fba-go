package service

import (
	"context"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type UserService struct {
	repo repo.Repository
}

func NewUserService(repository repo.Repository) *UserService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &UserService{repo: repository}
}

func (s *UserService) Get(ctx context.Context, id int) (dto.UserWithRelationDetail, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	return s.withRelations(ctx, user)
}

func (s *UserService) List(ctx context.Context, filter repo.UserFilter, page int, size int, basePath string) (pagination.PageData[dto.UserWithRelationDetail], error) {
	users, total, err := s.repo.ListUsers(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.UserWithRelationDetail]{}, err
	}
	items := make([]dto.UserWithRelationDetail, 0, len(users))
	for _, user := range users {
		detail, err := s.withRelations(ctx, user)
		if err != nil {
			return pagination.PageData[dto.UserWithRelationDetail]{}, err
		}
		items = append(items, detail)
	}
	return pagination.NewPageData(items, total, page, size, basePath), nil
}

func (s *UserService) Create(ctx context.Context, param dto.UserCreateParam) (dto.UserWithRelationDetail, error) {
	user, err := s.repo.CreateUser(ctx, param)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	return s.withRelations(ctx, user)
}

func (s *UserService) Update(ctx context.Context, id int, param dto.UserUpdateParam) error {
	return s.repo.UpdateUser(ctx, id, param)
}

func (s *UserService) UpdatePermission(ctx context.Context, id int, permissionType string) error {
	return s.repo.UpdateUserPermission(ctx, id, permissionType)
}

func (s *UserService) Delete(ctx context.Context, id int) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) Roles(ctx context.Context, id int) ([]dto.RoleDetail, error) {
	roles, err := s.repo.UserRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.RolesFromModel(roles), nil
}

func (s *UserService) withRelations(ctx context.Context, user model.User) (dto.UserWithRelationDetail, error) {
	var dept *model.Dept
	if user.DeptID != nil {
		item, err := s.repo.GetDept(ctx, *user.DeptID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		dept = &item
	}
	roles, err := s.repo.UserRoles(ctx, user.ID)
	if err != nil {
		return dto.UserWithRelationDetail{}, err
	}
	roleDetails := make([]dto.RoleWithRelationDetail, 0, len(roles))
	// Python get_join returns each user role with its menu and data-scope relations.
	for _, role := range roles {
		menus, err := s.repo.RoleMenus(ctx, role.ID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		scopes, err := s.repo.RoleScopes(ctx, role.ID)
		if err != nil {
			return dto.UserWithRelationDetail{}, err
		}
		roleDetails = append(roleDetails, dto.RoleWithRelations(role, menus, scopes))
	}
	return dto.UserWithRelations(user, dept, roleDetails), nil
}
