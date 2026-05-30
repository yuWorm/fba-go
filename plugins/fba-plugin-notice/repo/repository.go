package repo

import (
	"context"

	"github.com/yuWorm/fba-plugin-notice/dto"
	"github.com/yuWorm/fba-plugin-notice/model"
)

type NoticeFilter struct {
	Title  string
	Type   *int
	Status *int
}

type Repository interface {
	Get(ctx context.Context, id int) (model.Notice, error)
	List(ctx context.Context, filter NoticeFilter, page int, size int) ([]model.Notice, int64, error)
	Create(ctx context.Context, param dto.NoticeParam) error
	Update(ctx context.Context, id int, param dto.NoticeParam) error
	Delete(ctx context.Context, ids []int) error
}

type Seed struct {
	Notices []model.Notice
}

func SeedData() Seed {
	return Seed{Notices: model.SeedNotices()}
}
