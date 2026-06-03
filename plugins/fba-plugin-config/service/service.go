package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	coreerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-config/dto"
	"github.com/yuWorm/fba-plugin-config/repo"
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

func (s *Service) All(ctx context.Context, typeName string) ([]dto.ConfigDetail, error) {
	items, err := s.repo.All(ctx, typeName)
	if err != nil {
		return nil, err
	}
	return dto.ConfigListFromModel(items), nil
}

func (s *Service) Get(ctx context.Context, id int) (dto.ConfigDetail, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return dto.ConfigDetail{}, mapServiceError(err, "参数配置不存在")
	}
	return dto.ConfigFromModel(item), nil
}

func (s *Service) List(ctx context.Context, filter repo.ConfigFilter, page int, size int, basePath string) (pagination.PageData[dto.ConfigDetail], error) {
	items, total, err := s.repo.List(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.ConfigDetail]{}, err
	}
	return pagination.NewPageData(dto.ConfigListFromModel(items), total, page, size, basePath), nil
}

func (s *Service) Create(ctx context.Context, param dto.ConfigParam) error {
	if err := s.ensureKeyAvailable(ctx, 0, param.Key); err != nil {
		return err
	}
	return s.repo.Create(ctx, param)
}

func (s *Service) Update(ctx context.Context, id int, param dto.ConfigParam) error {
	if _, err := s.repo.Get(ctx, id); err != nil {
		return mapServiceError(err, "参数配置不存在")
	}
	if err := s.ensureKeyAvailable(ctx, id, param.Key); err != nil {
		return err
	}
	return s.repo.Update(ctx, id, param)
}

func (s *Service) BulkUpdate(ctx context.Context, params []dto.ConfigBulkParam) error {
	seenKeys := make(map[string]int, len(params))
	for _, param := range params {
		if _, err := s.repo.Get(ctx, param.ID); err != nil {
			return mapServiceError(err, "参数配置不存在")
		}
		if previousID, ok := seenKeys[param.Key]; ok && previousID != param.ID {
			return coreerrors.New(http.StatusConflict, http.StatusConflict, "参数配置键名重复", nil)
		}
		seenKeys[param.Key] = param.ID
		if err := s.ensureKeyAvailable(ctx, param.ID, param.Key); err != nil {
			return err
		}
	}
	return s.repo.BulkUpdate(ctx, params)
}

func (s *Service) Delete(ctx context.Context, ids []int) error {
	return s.repo.Delete(ctx, ids)
}

func (s *Service) ensureKeyAvailable(ctx context.Context, currentID int, key string) error {
	existing, err := s.repo.GetByKey(ctx, key)
	if err == nil {
		if existing.ID != currentID {
			return conflictError(key)
		}
		return nil
	}
	if errors.Is(err, repo.ErrNotFound) {
		return nil
	}
	return err
}

func mapServiceError(err error, message string) error {
	if errors.Is(err, repo.ErrNotFound) {
		return coreerrors.New(http.StatusNotFound, http.StatusNotFound, message, err)
	}
	return err
}

func conflictError(key string) error {
	return coreerrors.New(http.StatusConflict, http.StatusConflict, fmt.Sprintf("参数配置 %s 已存在", key), nil)
}
