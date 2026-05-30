package service

import (
	"context"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type DataScopeService struct {
	repo repo.Repository
}

func NewDataScopeService(repository repo.Repository) *DataScopeService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &DataScopeService{repo: repository}
}

func (s *DataScopeService) All(ctx context.Context) ([]dto.DataScopeDetail, error) {
	items, err := s.repo.AllDataScopes(ctx)
	if err != nil {
		return nil, err
	}
	return dto.DataScopesFromModel(items), nil
}

func (s *DataScopeService) Get(ctx context.Context, id int) (dto.DataScopeDetail, error) {
	item, err := s.repo.GetDataScope(ctx, id)
	if err != nil {
		return dto.DataScopeDetail{}, err
	}
	return dto.DataScopeFromModel(item), nil
}

func (s *DataScopeService) Rules(ctx context.Context, id int) (dto.DataScopeWithRelationDetail, error) {
	scope, rules, err := s.repo.DataScopeRules(ctx, id)
	if err != nil {
		return dto.DataScopeWithRelationDetail{}, err
	}
	return dto.DataScopeWithRules(scope, rules), nil
}

func (s *DataScopeService) List(ctx context.Context, filter repo.DataScopeFilter, page int, size int, basePath string) (pagination.PageData[dto.DataScopeDetail], error) {
	items, total, err := s.repo.ListDataScopes(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.DataScopeDetail]{}, err
	}
	return pagination.NewPageData(dto.DataScopesFromModel(items), total, page, size, basePath), nil
}

func (s *DataScopeService) Create(ctx context.Context, param dto.DataScopeParam) error {
	return s.repo.CreateDataScope(ctx, param)
}

func (s *DataScopeService) Update(ctx context.Context, id int, param dto.DataScopeParam) error {
	return s.repo.UpdateDataScope(ctx, id, param)
}

func (s *DataScopeService) UpdateRules(ctx context.Context, id int, ruleIDs []int) error {
	return s.repo.UpdateDataScopeRules(ctx, id, ruleIDs)
}

func (s *DataScopeService) Delete(ctx context.Context, ids []int) error {
	return s.repo.DeleteDataScopes(ctx, ids)
}
