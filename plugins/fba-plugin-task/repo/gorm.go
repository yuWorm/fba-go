package repo

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/model"
	"gorm.io/gorm"
)

type GORMRepository struct {
	provider db.Provider
}

func NewGORMRepository(provider db.Provider) *GORMRepository {
	return &GORMRepository{provider: provider}
}

func (r *GORMRepository) AllSchedulers(ctx context.Context) ([]model.TaskScheduler, error) {
	var items []model.TaskScheduler
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetScheduler(ctx context.Context, id int) (model.TaskScheduler, error) {
	var item model.TaskScheduler
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListSchedulers(ctx context.Context, filter SchedulerFilter, page int, size int) ([]model.TaskScheduler, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.TaskScheduler{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	return paginate[model.TaskScheduler](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) CreateScheduler(ctx context.Context, param dto.SchedulerParam) error {
	item := schedulerFromParam(0, param)
	item.ID = 0
	return r.provider.Write().WithContext(ctx).Create(&item).Error
}

func (r *GORMRepository) UpdateScheduler(ctx context.Context, id int, param dto.SchedulerParam) error {
	item := schedulerFromParam(id, param)
	updates := map[string]any{
		"name":            item.Name,
		"task":            item.Task,
		"args":            item.Args,
		"kwargs":          item.Kwargs,
		"queue":           item.Queue,
		"exchange":        item.Exchange,
		"routing_key":     item.RoutingKey,
		"expire_seconds":  item.ExpireSeconds,
		"type":            item.Type,
		"interval_every":  item.IntervalEvery,
		"interval_period": item.IntervalPeriod,
		"crontab":         item.Crontab,
		"one_off":         item.OneOff,
		"remark":          item.Remark,
	}
	result := r.provider.Write().WithContext(ctx).Model(&model.TaskScheduler{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) ToggleSchedulerStatus(ctx context.Context, id int) error {
	item, err := r.GetScheduler(ctx, id)
	if err != nil {
		return err
	}
	result := r.provider.Write().WithContext(ctx).Model(&model.TaskScheduler{}).Where("id = ?", id).Update("enabled", !item.Enabled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) DeleteScheduler(ctx context.Context, id int) error {
	result := r.provider.Write().WithContext(ctx).Delete(&model.TaskScheduler{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) GetTaskResult(ctx context.Context, id int) (model.TaskResult, error) {
	var item model.TaskResult
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListTaskResults(ctx context.Context, filter ResultFilter, page int, size int) ([]model.TaskResult, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.TaskResult{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.TaskID != "" {
		query = query.Where("task_id LIKE ?", "%"+filter.TaskID+"%")
	}
	return paginate[model.TaskResult](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) DeleteTaskResults(ctx context.Context, ids []int) error {
	return r.provider.Write().WithContext(ctx).Delete(&model.TaskResult{}, ids).Error
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
