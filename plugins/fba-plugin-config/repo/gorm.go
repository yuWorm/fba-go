package repo

import (
	"context"
	"errors"
	"time"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-config/dto"
	"github.com/yuWorm/fba-plugin-config/model"
	"gorm.io/gorm"
)

type GORMRepository struct {
	provider db.Provider
}

func NewGORMRepository(provider db.Provider) *GORMRepository {
	return &GORMRepository{provider: provider}
}

func (r *GORMRepository) All(ctx context.Context, typeName string) ([]model.Config, error) {
	query := r.provider.Read().WithContext(ctx).Where("deleted = ?", 0)
	if typeName != "" {
		query = query.Where("type = ?", typeName)
	}
	var items []model.Config
	err := query.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) Get(ctx context.Context, id int) (model.Config, error) {
	var item model.Config
	err := r.provider.Read().WithContext(ctx).Where("deleted = ?", 0).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetByKey(ctx context.Context, key string) (model.Config, error) {
	var item model.Config
	err := r.provider.Read().WithContext(ctx).Where("deleted = ? AND key = ?", 0, key).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) List(ctx context.Context, filter ConfigFilter, page int, size int) ([]model.Config, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Config{}).Where("deleted = ?", 0)
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	return paginate[model.Config](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) Create(ctx context.Context, param dto.ConfigParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.Config{
		Name:       param.Name,
		Type:       param.Type,
		Key:        param.Key,
		Value:      param.Value,
		IsFrontend: param.IsFrontend,
		Remark:     param.Remark,
	}).Error
}

func (r *GORMRepository) Update(ctx context.Context, id int, param dto.ConfigParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Config{}).Where("id = ? AND deleted = ?", id, 0).Updates(updateMap(param))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) BulkUpdate(ctx context.Context, params []dto.ConfigBulkParam) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, param := range params {
			result := tx.Model(&model.Config{}).Where("id = ? AND deleted = ?", param.ID, 0).Updates(updateMap(param.ConfigParam))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrNotFound
			}
		}
		return nil
	})
}

func (r *GORMRepository) Delete(ctx context.Context, ids []int) error {
	now := time.Now()
	result := r.provider.Write().WithContext(ctx).Model(&model.Config{}).Where("id IN ? AND deleted = ?", ids, 0).Updates(map[string]any{
		"deleted":      gorm.Expr("id"),
		"deleted_time": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func updateMap(param dto.ConfigParam) map[string]any {
	return map[string]any{
		"name":        param.Name,
		"type":        param.Type,
		"key":         param.Key,
		"value":       param.Value,
		"is_frontend": param.IsFrontend,
		"remark":      param.Remark,
	}
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
