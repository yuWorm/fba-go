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
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.DataScopeDetail{}, dataScopeNotFound("数据范围不存在", err)
		}
		return dto.DataScopeDetail{}, err
	}
	return dto.DataScopeFromModel(item), nil
}

func (s *DataScopeService) Rules(ctx context.Context, id int) (dto.DataScopeWithRelationDetail, error) {
	scope, rules, err := s.repo.DataScopeRules(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.DataScopeWithRelationDetail{}, dataScopeNotFound("数据范围不存在", err)
		}
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
	if _, err := s.repo.GetDataScopeByName(ctx, param.Name); err == nil {
		return dataScopeConflict("数据范围已存在", nil)
	} else if !stderrors.Is(err, repo.ErrNotFound) {
		return err
	}
	return s.repo.CreateDataScope(ctx, param)
}

func (s *DataScopeService) Update(ctx context.Context, id int, param dto.DataScopeParam) error {
	scope, err := s.ensureDataScope(ctx, id)
	if err != nil {
		return err
	}
	if scope.Name != param.Name {
		if _, err := s.repo.GetDataScopeByName(ctx, param.Name); err == nil {
			return dataScopeConflict("数据范围已存在", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	return s.repo.UpdateDataScope(ctx, id, param)
}

func (s *DataScopeService) UpdateRules(ctx context.Context, id int, ruleIDs []int) error {
	if _, err := s.ensureDataScope(ctx, id); err != nil {
		return err
	}
	// Python validates every submitted rule ID before replacing the relation set.
	if err := s.ensureDataRules(ctx, ruleIDs); err != nil {
		return err
	}
	return s.repo.UpdateDataScopeRules(ctx, id, ruleIDs)
}

func (s *DataScopeService) Delete(ctx context.Context, ids []int) error {
	return s.repo.DeleteDataScopes(ctx, ids)
}

func (s *DataScopeService) ensureDataScope(ctx context.Context, id int) (model.DataScope, error) {
	scope, err := s.repo.GetDataScope(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return model.DataScope{}, dataScopeNotFound("数据范围不存在", err)
		}
		return model.DataScope{}, err
	}
	return scope, nil
}

func (s *DataScopeService) ensureDataRules(ctx context.Context, ids []int) error {
	for _, id := range uniqueIDs(ids) {
		if _, err := s.repo.GetDataRule(ctx, id); err != nil {
			if stderrors.Is(err, repo.ErrNotFound) {
				return dataScopeNotFound("数据规则不存在", err)
			}
			return err
		}
	}
	return nil
}

func dataScopeNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func dataScopeConflict(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}
