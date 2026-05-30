package service

import (
	"context"

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
	return s.repo.CreateMenu(ctx, param)
}

func (s *MenuService) Update(ctx context.Context, id int, param dto.MenuParam) error {
	return s.repo.UpdateMenu(ctx, id, param)
}

func (s *MenuService) Delete(ctx context.Context, id int) error {
	return s.repo.DeleteMenu(ctx, id)
}

func buildMenuTree(items []model.Menu) []dto.MenuDetail {
	nodes := make(map[int]*dto.MenuDetail, len(items))
	for _, item := range items {
		detail := dto.MenuFromModel(item)
		detail.Children = []dto.MenuDetail{}
		nodes[item.ID] = &detail
	}

	// Build child links before collecting roots; appending root values too early would copy stale children.
	for _, item := range items {
		node := nodes[item.ID]
		if item.ParentID != nil {
			if parent, ok := nodes[*item.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	roots := make([]dto.MenuDetail, 0, len(items))
	for _, item := range items {
		// Keep filtered or orphaned children visible as roots, matching the Python tree helper.
		if item.ParentID == nil || nodes[*item.ParentID] == nil {
			roots = append(roots, *nodes[item.ID])
		}
	}
	return roots
}

func buildSidebarTree(items []model.Menu) []dto.SidebarMenu {
	nodes := make(map[int]*dto.SidebarMenu, len(items))
	for _, item := range items {
		sidebar := dto.SidebarMenuFromModel(item)
		sidebar.Children = []dto.SidebarMenu{}
		nodes[item.ID] = &sidebar
	}

	// Build child links before collecting roots; appending root values too early would copy stale children.
	for _, item := range items {
		node := nodes[item.ID]
		if item.ParentID != nil {
			if parent, ok := nodes[*item.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	roots := make([]dto.SidebarMenu, 0, len(items))
	for _, item := range items {
		// Sidebar filtering can remove parents; promote orphaned nodes instead of dropping them.
		if item.ParentID == nil || nodes[*item.ParentID] == nil {
			roots = append(roots, *nodes[item.ID])
		}
	}
	return roots
}
