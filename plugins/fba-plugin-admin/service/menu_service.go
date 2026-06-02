package service

import (
	"context"
	stderrors "errors"
	"net/http"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type MenuService struct {
	repo repo.Repository
}

func NewMenuService(repository repo.Repository) *MenuService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &MenuService{repo: repository}
}

func (s *MenuService) Get(ctx context.Context, id int) (dto.MenuDetail, error) {
	item, err := s.repo.GetMenu(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.MenuDetail{}, menuNotFound("菜单不存在", err)
		}
		return dto.MenuDetail{}, err
	}
	return dto.MenuFromModel(item), nil
}

func (s *MenuService) Tree(ctx context.Context, filter repo.MenuFilter) ([]dto.MenuDetail, error) {
	items, err := s.repo.ListMenus(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(items), nil
}

func (s *MenuService) Sidebar(ctx context.Context) ([]dto.SidebarMenu, error) {
	items, err := s.repo.SidebarMenus(ctx)
	if err != nil {
		return nil, err
	}
	return buildSidebarTree(items), nil
}

func (s *MenuService) Create(ctx context.Context, param dto.MenuParam) error {
	if _, err := s.repo.GetMenuByTitle(ctx, param.Title); err == nil {
		return menuConflict("菜单标题已存在", nil)
	} else if !stderrors.Is(err, repo.ErrNotFound) {
		return err
	}
	if err := s.ensureParent(ctx, param.ParentID); err != nil {
		return err
	}
	return s.repo.CreateMenu(ctx, param)
}

func (s *MenuService) Update(ctx context.Context, id int, param dto.MenuParam) error {
	menu, err := s.ensureMenu(ctx, id)
	if err != nil {
		return err
	}
	if menu.Title != param.Title {
		if _, err := s.repo.GetMenuByTitle(ctx, param.Title); err == nil {
			return menuConflict("菜单标题已存在", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	if err := s.ensureParent(ctx, param.ParentID); err != nil {
		return err
	}
	if param.ParentID != nil && *param.ParentID == menu.ID {
		return menuForbidden("禁止关联自身为父级", nil)
	}
	return s.repo.UpdateMenu(ctx, id, param)
}

func (s *MenuService) Delete(ctx context.Context, id int) error {
	hasChildren, err := s.repo.MenuHasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return menuConflict("菜单下存在子菜单，无法删除", nil)
	}
	return s.repo.DeleteMenu(ctx, id)
}

func (s *MenuService) ensureMenu(ctx context.Context, id int) (model.Menu, error) {
	menu, err := s.repo.GetMenu(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return model.Menu{}, menuNotFound("菜单不存在", err)
		}
		return model.Menu{}, err
	}
	return menu, nil
}

func (s *MenuService) ensureParent(ctx context.Context, parentID *int) error {
	// Python treats a missing or zero parent_id as "no parent"; keep that compatibility for request bodies.
	if parentID == nil || *parentID == 0 {
		return nil
	}
	if _, err := s.repo.GetMenu(ctx, *parentID); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return menuNotFound("父级菜单不存在", err)
		}
		return err
	}
	return nil
}

func menuNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func menuConflict(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}

func menuForbidden(message string, cause error) error {
	return fbaerrors.New(http.StatusForbidden, http.StatusForbidden, message, cause)
}

func buildMenuTree(items []model.Menu) []dto.MenuDetail {
	byID := make(map[int]model.Menu, len(items))
	childrenByParent := make(map[int][]model.Menu, len(items))
	for _, item := range items {
		byID[item.ID] = item
		if item.ParentID != nil {
			childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], item)
		}
	}

	var buildNode func(model.Menu, map[int]bool) dto.MenuDetail
	buildNode = func(item model.Menu, visiting map[int]bool) dto.MenuDetail {
		detail := dto.MenuFromModel(item)
		children := childrenByParent[item.ID]
		if len(children) == 0 {
			return detail
		}

		detail.Children = make([]dto.MenuDetail, 0, len(children))
		visiting[item.ID] = true
		defer delete(visiting, item.ID)
		for _, child := range children {
			// Build values from the leaves upward so grandchildren are not lost through stale value copies.
			if visiting[child.ID] {
				continue
			}
			detail.Children = append(detail.Children, buildNode(child, visiting))
		}
		return detail
	}

	roots := make([]dto.MenuDetail, 0, len(items))
	for _, item := range items {
		// Keep filtered or orphaned children visible as roots, matching the Python tree helper.
		parentExists := false
		if item.ParentID != nil {
			_, parentExists = byID[*item.ParentID]
		}
		if item.ParentID == nil || !parentExists {
			roots = append(roots, buildNode(item, map[int]bool{}))
		}
	}
	return roots
}

func buildSidebarTree(items []model.Menu) []dto.SidebarMenu {
	byID := make(map[int]model.Menu, len(items))
	childrenByParent := make(map[int][]model.Menu, len(items))
	for _, item := range items {
		byID[item.ID] = item
		if item.ParentID != nil {
			childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], item)
		}
	}

	var buildNode func(model.Menu, map[int]bool) dto.SidebarMenu
	buildNode = func(item model.Menu, visiting map[int]bool) dto.SidebarMenu {
		sidebar := dto.SidebarMenuFromModel(item)
		children := childrenByParent[item.ID]
		if len(children) == 0 {
			return sidebar
		}

		sidebar.Children = make([]dto.SidebarMenu, 0, len(children))
		visiting[item.ID] = true
		defer delete(visiting, item.ID)
		for _, child := range children {
			// Build values from the leaves upward so grandchildren are not lost through stale value copies.
			if visiting[child.ID] {
				continue
			}
			sidebar.Children = append(sidebar.Children, buildNode(child, visiting))
		}
		return sidebar
	}

	roots := make([]dto.SidebarMenu, 0, len(items))
	for _, item := range items {
		// Sidebar filtering can remove parents; promote orphaned nodes instead of dropping them.
		parentExists := false
		if item.ParentID != nil {
			_, parentExists = byID[*item.ParentID]
		}
		if item.ParentID == nil || !parentExists {
			roots = append(roots, buildNode(item, map[int]bool{}))
		}
	}
	return roots
}
