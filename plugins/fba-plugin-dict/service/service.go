package service

import (
	"context"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/repo"
)

type Service struct {
	repo        repo.Repository
	invalidator CacheInvalidator
}

func New(repository repo.Repository, invalidator CacheInvalidator) *Service {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	if invalidator == nil {
		invalidator = NoopInvalidator{}
	}
	return &Service{
		repo:        repository,
		invalidator: invalidator,
	}
}

func (s *Service) AllTypes(ctx context.Context) ([]dto.DictTypeDetail, error) {
	items, err := s.repo.AllTypes(ctx)
	if err != nil {
		return nil, err
	}
	return dto.DictTypesFromModel(items), nil
}

func (s *Service) GetType(ctx context.Context, id int) (dto.DictTypeDetail, error) {
	item, err := s.repo.GetType(ctx, id)
	if err != nil {
		return dto.DictTypeDetail{}, err
	}
	return dto.DictTypeFromModel(item), nil
}

func (s *Service) ListTypes(ctx context.Context, filter repo.DictTypeFilter, page int, size int, basePath string) (pagination.PageData[dto.DictTypeDetail], error) {
	items, total, err := s.repo.ListTypes(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.DictTypeDetail]{}, err
	}
	return pagination.NewPageData(dto.DictTypesFromModel(items), total, page, size, basePath), nil
}

func (s *Service) CreateType(ctx context.Context, param dto.DictTypeParam) error {
	if err := s.repo.CreateType(ctx, param); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}

func (s *Service) UpdateType(ctx context.Context, id int, param dto.DictTypeParam) error {
	if err := s.repo.UpdateType(ctx, id, param); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}

func (s *Service) DeleteTypes(ctx context.Context, ids []int) error {
	if err := s.repo.DeleteTypes(ctx, ids); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}

func (s *Service) AllData(ctx context.Context) ([]dto.DictDataDetail, error) {
	items, err := s.repo.AllData(ctx)
	if err != nil {
		return nil, err
	}
	return dto.DictDataListFromModel(items), nil
}

func (s *Service) GetData(ctx context.Context, id int) (dto.DictDataDetail, error) {
	item, err := s.repo.GetData(ctx, id)
	if err != nil {
		return dto.DictDataDetail{}, err
	}
	return dto.DictDataFromModel(item), nil
}

func (s *Service) GetDataByTypeCode(ctx context.Context, code string) ([]dto.DictDataDetail, error) {
	items, err := s.repo.DataByTypeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return dto.DictDataListFromModel(items), nil
}

func (s *Service) ListData(ctx context.Context, filter repo.DictDataFilter, page int, size int, basePath string) (pagination.PageData[dto.DictDataDetail], error) {
	items, total, err := s.repo.ListData(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.DictDataDetail]{}, err
	}
	return pagination.NewPageData(dto.DictDataListFromModel(items), total, page, size, basePath), nil
}

func (s *Service) CreateData(ctx context.Context, param dto.DictDataParam) error {
	if err := s.repo.CreateData(ctx, param); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}

func (s *Service) UpdateData(ctx context.Context, id int, param dto.DictDataParam) error {
	if err := s.repo.UpdateData(ctx, id, param); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}

func (s *Service) DeleteData(ctx context.Context, ids []int) error {
	if err := s.repo.DeleteData(ctx, ids); err != nil {
		return err
	}
	return s.invalidator.InvalidateDict(ctx)
}
