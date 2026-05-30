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
	nodes := make(map[int]*dto.DeptDetail, len(items))
	for _, item := range items {
		detail := dto.DeptFromModel(item)
		detail.Children = []dto.DeptDetail{}
		nodes[item.ID] = &detail
	}

	// Build child links before collecting roots; appending values too early would copy stale children.
	for _, item := range items {
		node := nodes[item.ID]
		if item.ParentID != nil {
			if parent, ok := nodes[*item.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	roots := make([]dto.DeptDetail, 0, len(items))
	for _, item := range items {
		// Keep filtered or orphaned children visible as roots, matching the Python tree helper.
		if item.ParentID == nil || nodes[*item.ParentID] == nil {
			roots = append(roots, *nodes[item.ID])
		}
	}
	return roots
}
