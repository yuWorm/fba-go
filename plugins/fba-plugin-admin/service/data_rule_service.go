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

type DataRuleService struct {
	repo repo.Repository
}

func NewDataRuleService(repository repo.Repository) *DataRuleService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &DataRuleService{repo: repository}
}

func (s *DataRuleService) All(ctx context.Context) ([]dto.DataRuleDetail, error) {
	items, err := s.repo.AllDataRules(ctx)
	if err != nil {
		return nil, err
	}
	return dto.DataRulesFromModel(items), nil
}

func (s *DataRuleService) Models(ctx context.Context) ([]string, error) {
	return s.repo.DataRuleModels(ctx)
}

func (s *DataRuleService) Columns(ctx context.Context, modelName string) ([]dto.DataRuleColumnDetail, error) {
	items, err := s.repo.DataRuleModelColumns(ctx, modelName)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return nil, dataRuleNotFound("数据规则可用模型不存在", err)
		}
		return nil, err
	}
	return dto.DataRuleColumnsFromModel(items), nil
}

func (s *DataRuleService) ValueTemplateVariables(ctx context.Context) ([]dto.DataRuleTemplateVariableDetail, error) {
	items, err := s.repo.DataRuleValueTemplateVariables(ctx)
	if err != nil {
		return nil, err
	}
	return dto.DataRuleTemplateVariablesFromModel(items), nil
}

func (s *DataRuleService) Get(ctx context.Context, id int) (dto.DataRuleDetail, error) {
	item, err := s.repo.GetDataRule(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.DataRuleDetail{}, dataRuleNotFound("数据规则不存在", err)
		}
		return dto.DataRuleDetail{}, err
	}
	return dto.DataRuleFromModel(item), nil
}

func (s *DataRuleService) List(ctx context.Context, filter repo.DataRuleFilter, page int, size int, basePath string) (pagination.PageData[dto.DataRuleDetail], error) {
	items, total, err := s.repo.ListDataRules(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.DataRuleDetail]{}, err
	}
	return pagination.NewPageData(dto.DataRulesFromModel(items), total, page, size, basePath), nil
}

func (s *DataRuleService) Create(ctx context.Context, param dto.DataRuleParam) error {
	if _, err := s.repo.GetDataRuleByName(ctx, param.Name); err == nil {
		return dataRuleConflict("数据规则已存在", nil)
	} else if !stderrors.Is(err, repo.ErrNotFound) {
		return err
	}
	return s.repo.CreateDataRule(ctx, param)
}

func (s *DataRuleService) Update(ctx context.Context, id int, param dto.DataRuleParam) error {
	rule, err := s.ensureDataRule(ctx, id)
	if err != nil {
		return err
	}
	if rule.Name != param.Name {
		if _, err := s.repo.GetDataRuleByName(ctx, param.Name); err == nil {
			return dataRuleConflict("数据规则已存在", nil)
		} else if !stderrors.Is(err, repo.ErrNotFound) {
			return err
		}
	}
	return s.repo.UpdateDataRule(ctx, id, param)
}

func (s *DataRuleService) Delete(ctx context.Context, ids []int) error {
	return s.repo.DeleteDataRules(ctx, ids)
}

func (s *DataRuleService) ensureDataRule(ctx context.Context, id int) (model.DataRule, error) {
	rule, err := s.repo.GetDataRule(ctx, id)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return model.DataRule{}, dataRuleNotFound("数据规则不存在", err)
		}
		return model.DataRule{}, err
	}
	return rule, nil
}

func dataRuleNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func dataRuleConflict(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}
