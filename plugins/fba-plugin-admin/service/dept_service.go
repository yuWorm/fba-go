package service

import (
	"context"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type DeptService struct {
	repo repo.Repository
}

func NewDeptService(repository repo.Repository) *DeptService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &DeptService{repo: repository}
}

func (s *DeptService) Get(ctx context.Context, id int) (dto.DeptDetail, error) {
	item, err := s.repo.GetDept(ctx, id)
	if err != nil {
		return dto.DeptDetail{}, err
	}
	return dto.DeptFromModel(item), nil
}

func (s *DeptService) Tree(ctx context.Context, filter repo.DeptFilter) ([]dto.DeptDetail, error) {
	items, err := s.repo.ListDepts(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildDeptTree(items), nil
}

func (s *DeptService) Create(ctx context.Context, param dto.DeptParam) error {
	return s.repo.CreateDept(ctx, param)
}

func (s *DeptService) Update(ctx context.Context, id int, param dto.DeptParam) error {
	return s.repo.UpdateDept(ctx, id, param)
}

func (s *DeptService) Delete(ctx context.Context, id int) error {
	return s.repo.DeleteDept(ctx, id)
}

func buildDeptTree(items []model.Dept) []dto.DeptDetail {
	byID := make(map[int]model.Dept, len(items))
	childrenByParent := make(map[int][]model.Dept, len(items))
	for _, item := range items {
		byID[item.ID] = item
		if item.ParentID != nil {
			childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], item)
		}
	}

	var buildNode func(model.Dept, map[int]bool) dto.DeptDetail
	buildNode = func(item model.Dept, visiting map[int]bool) dto.DeptDetail {
		detail := dto.DeptFromModel(item)
		children := childrenByParent[item.ID]
		if len(children) == 0 {
			return detail
		}

		detail.Children = make([]dto.DeptDetail, 0, len(children))
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

	roots := make([]dto.DeptDetail, 0, len(items))
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
