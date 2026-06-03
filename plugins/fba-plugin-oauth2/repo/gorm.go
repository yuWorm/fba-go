package repo

import (
	"context"
	"errors"
	"time"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-oauth2/model"
	"gorm.io/gorm"
)

type GORMRepository struct {
	provider db.Provider
}

func NewGORMRepository(provider db.Provider) *GORMRepository {
	return &GORMRepository{provider: provider}
}

func (r *GORMRepository) GetBySID(ctx context.Context, sid string, source string) (model.UserSocial, error) {
	var item model.UserSocial
	err := r.provider.Write().WithContext(ctx).Where("sid = ? AND source = ? AND deleted = 0", sid, source).First(&item).Error
	return item, normalizeError(err)
}

func (r *GORMRepository) CheckBinding(ctx context.Context, userID int, source string) (model.UserSocial, error) {
	var item model.UserSocial
	err := r.provider.Write().WithContext(ctx).Where("user_id = ? AND source = ? AND deleted = 0", userID, source).First(&item).Error
	return item, normalizeError(err)
}

func (r *GORMRepository) ListByUserID(ctx context.Context, userID int) ([]model.UserSocial, error) {
	var items []model.UserSocial
	err := r.provider.Write().WithContext(ctx).Where("user_id = ? AND deleted = 0", userID).Order("id asc").Find(&items).Error
	return items, err
}

func (r *GORMRepository) Create(ctx context.Context, param CreateUserSocialParam) (model.UserSocial, error) {
	item := model.UserSocial{
		SID:    param.SID,
		Source: param.Source,
		UserID: param.UserID,
	}
	if err := r.provider.Write().WithContext(ctx).Create(&item).Error; err != nil {
		return model.UserSocial{}, err
	}
	return item, nil
}

func (r *GORMRepository) Delete(ctx context.Context, userID int, source string) (int, error) {
	now := time.Now()
	result := r.provider.Write().WithContext(ctx).
		Model(&model.UserSocial{}).
		Where("user_id = ? AND source = ? AND deleted = 0", userID, source).
		Updates(map[string]any{"deleted": gorm.Expr("id"), "deleted_time": now})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, ErrNotFound
	}
	return int(result.RowsAffected), nil
}

func normalizeError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
