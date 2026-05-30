package service

import (
	"context"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-notice/dto"
	"github.com/yuWorm/fba-plugin-notice/repo"
)

type Service struct {
	repo repo.Repository
}

func New(repository repo.Repository) *Service {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &Service{repo: repository}
}

func (s *Service) Get(ctx context.Context, id int) (dto.NoticeDetail, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return dto.NoticeDetail{}, err
	}
	return dto.NoticeFromModel(item), nil
}

func (s *Service) List(ctx context.Context, filter repo.NoticeFilter, page int, size int, basePath string) (pagination.PageData[dto.NoticeDetail], error) {
	items, total, err := s.repo.List(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.NoticeDetail]{}, err
	}
	return pagination.NewPageData(dto.NoticesFromModel(items), total, page, size, basePath), nil
}

func (s *Service) Create(ctx context.Context, param dto.NoticeParam) error {
	return s.repo.Create(ctx, param)
}

func (s *Service) Update(ctx context.Context, id int, param dto.NoticeParam) error {
	return s.repo.Update(ctx, id, param)
}

func (s *Service) Delete(ctx context.Context, ids []int) error {
	return s.repo.Delete(ctx, ids)
}
