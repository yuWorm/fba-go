package repo

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-notice/dto"
	"github.com/yuWorm/fba-plugin-notice/model"
	"gorm.io/gorm"
)

type GORMRepository struct {
	provider db.Provider
}

func NewGORMRepository(provider db.Provider) *GORMRepository {
	return &GORMRepository{provider: provider}
}

func (r *GORMRepository) Get(ctx context.Context, id int) (model.Notice, error) {
	var item model.Notice
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) List(ctx context.Context, filter NoticeFilter, page int, size int) ([]model.Notice, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Notice{})
	if filter.Title != "" {
		query = query.Where("title LIKE ?", "%"+filter.Title+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return paginate[model.Notice](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) Create(ctx context.Context, param dto.NoticeParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.Notice{
		Title:   param.Title,
		Type:    param.Type,
		Status:  param.Status,
		Content: param.Content,
	}).Error
}

func (r *GORMRepository) Update(ctx context.Context, id int, param dto.NoticeParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Notice{}).Where("id = ?", id).Updates(map[string]any{
		"title":   param.Title,
		"type":    param.Type,
		"status":  param.Status,
		"content": param.Content,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) Delete(ctx context.Context, ids []int) error {
	result := r.provider.Write().WithContext(ctx).Delete(&model.Notice{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
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
