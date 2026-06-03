package repo

import (
	"context"
	"errors"

	"github.com/yuWorm/fba-plugin-config/dto"
	"github.com/yuWorm/fba-plugin-config/model"
)

var ErrNotFound = errors.New("not found")

type ConfigFilter struct {
	Name string
	Type string
}

type Repository interface {
	All(ctx context.Context, typeName string) ([]model.Config, error)
	Get(ctx context.Context, id int) (model.Config, error)
	GetByKey(ctx context.Context, key string) (model.Config, error)
	List(ctx context.Context, filter ConfigFilter, page int, size int) ([]model.Config, int64, error)
	Create(ctx context.Context, param dto.ConfigParam) error
	Update(ctx context.Context, id int, param dto.ConfigParam) error
	BulkUpdate(ctx context.Context, params []dto.ConfigBulkParam) error
	Delete(ctx context.Context, ids []int) error
}

type Seed struct {
	Configs []model.Config
}

func SeedData() Seed {
	return Seed{Configs: model.SeedConfigs()}
}
