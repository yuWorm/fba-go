package repo

import (
	"context"

	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/model"
)

type DictTypeFilter struct {
	Name string
	Code string
}

type DictDataFilter struct {
	TypeCode string
	Label    string
	Value    string
	Status   *int
	TypeID   *int
}

type Repository interface {
	AllTypes(ctx context.Context) ([]model.DictType, error)
	GetType(ctx context.Context, id int) (model.DictType, error)
	ListTypes(ctx context.Context, filter DictTypeFilter, page int, size int) ([]model.DictType, int64, error)
	CreateType(ctx context.Context, param dto.DictTypeParam) error
	UpdateType(ctx context.Context, id int, param dto.DictTypeParam) error
	DeleteTypes(ctx context.Context, ids []int) error

	AllData(ctx context.Context) ([]model.DictData, error)
	GetData(ctx context.Context, id int) (model.DictData, error)
	DataByTypeCode(ctx context.Context, code string) ([]model.DictData, error)
	ListData(ctx context.Context, filter DictDataFilter, page int, size int) ([]model.DictData, int64, error)
	CreateData(ctx context.Context, param dto.DictDataParam) error
	UpdateData(ctx context.Context, id int, param dto.DictDataParam) error
	DeleteData(ctx context.Context, ids []int) error
}

type Seed struct {
	Types []model.DictType
	Data  []model.DictData
}

func SeedData() Seed {
	return Seed{
		Types: model.SeedDictTypes(),
		Data:  model.SeedDictData(),
	}
}
