package repo

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/model"
	"gorm.io/gorm"
)

type GORMRepository struct {
	provider db.Provider
}

func NewGORMRepository(provider db.Provider) *GORMRepository {
	return &GORMRepository{provider: provider}
}

func (r *GORMRepository) AllTypes(ctx context.Context) ([]model.DictType, error) {
	var items []model.DictType
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetType(ctx context.Context, id int) (model.DictType, error) {
	var item model.DictType
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListTypes(ctx context.Context, filter DictTypeFilter, page int, size int) ([]model.DictType, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.DictType{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Code != "" {
		query = query.Where("code LIKE ?", "%"+filter.Code+"%")
	}
	return paginate[model.DictType](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) CreateType(ctx context.Context, param dto.DictTypeParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.DictType{
		Name:   param.Name,
		Code:   param.Code,
		Remark: param.Remark,
	}).Error
}

func (r *GORMRepository) UpdateType(ctx context.Context, id int, param dto.DictTypeParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.DictType{}).Where("id = ?", id).Updates(map[string]any{
		"name":   param.Name,
		"code":   param.Code,
		"remark": param.Remark,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) DeleteTypes(ctx context.Context, ids []int) error {
	return r.provider.Write().WithContext(ctx).Delete(&model.DictType{}, ids).Error
}

func (r *GORMRepository) AllData(ctx context.Context) ([]model.DictData, error) {
	var items []model.DictData
	err := r.provider.Read().WithContext(ctx).Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetData(ctx context.Context, id int) (model.DictData, error) {
	var item model.DictData
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) DataByTypeCode(ctx context.Context, code string) ([]model.DictData, error) {
	var items []model.DictData
	err := r.provider.Read().WithContext(ctx).Where("type_code = ?", code).Order("sort ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items, nil
}

func (r *GORMRepository) ListData(ctx context.Context, filter DictDataFilter, page int, size int) ([]model.DictData, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.DictData{})
	if filter.TypeCode != "" {
		query = query.Where("type_code LIKE ?", "%"+filter.TypeCode+"%")
	}
	if filter.Label != "" {
		query = query.Where("label LIKE ?", "%"+filter.Label+"%")
	}
	if filter.Value != "" {
		query = query.Where("value LIKE ?", "%"+filter.Value+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.TypeID != nil {
		query = query.Where("type_id = ?", *filter.TypeID)
	}
	return paginate[model.DictData](query.Order("sort ASC, id ASC"), page, size)
}

func (r *GORMRepository) CreateData(ctx context.Context, param dto.DictDataParam) error {
	dictType, err := r.GetType(ctx, param.TypeID)
	if err != nil {
		return err
	}
	return r.provider.Write().WithContext(ctx).Create(&model.DictData{
		TypeID:   param.TypeID,
		TypeCode: dictType.Code,
		Label:    param.Label,
		Value:    param.Value,
		Color:    param.Color,
		Sort:     param.Sort,
		Status:   param.Status,
		Remark:   param.Remark,
	}).Error
}

func (r *GORMRepository) UpdateData(ctx context.Context, id int, param dto.DictDataParam) error {
	dictType, err := r.GetType(ctx, param.TypeID)
	if err != nil {
		return err
	}
	result := r.provider.Write().WithContext(ctx).Model(&model.DictData{}).Where("id = ?", id).Updates(map[string]any{
		"type_id":   param.TypeID,
		"type_code": dictType.Code,
		"label":     param.Label,
		"value":     param.Value,
		"color":     param.Color,
		"sort":      param.Sort,
		"status":    param.Status,
		"remark":    param.Remark,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) DeleteData(ctx context.Context, ids []int) error {
	return r.provider.Write().WithContext(ctx).Delete(&model.DictData{}, ids).Error
}

func paginate[T any](query *gorm.DB, page int, size int) ([]T, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []T
	err := query.Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func mapGORMError(err error) error {
	if err == nil {
		return nil
	}
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
