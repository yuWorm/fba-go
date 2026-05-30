package repo

import (
	"context"
	"errors"

	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/model"
)

var ErrNotFound = errors.New("not found")

type SchedulerFilter struct {
	Name string
	Type *int
}

type ResultFilter struct {
	Name   string
	TaskID string
}

type Repository interface {
	AllSchedulers(ctx context.Context) ([]model.TaskScheduler, error)
	GetScheduler(ctx context.Context, id int) (model.TaskScheduler, error)
	ListSchedulers(ctx context.Context, filter SchedulerFilter, page int, size int) ([]model.TaskScheduler, int64, error)
	CreateScheduler(ctx context.Context, param dto.SchedulerParam) error
	UpdateScheduler(ctx context.Context, id int, param dto.SchedulerParam) error
	ToggleSchedulerStatus(ctx context.Context, id int) error
	DeleteScheduler(ctx context.Context, id int) error

	GetTaskResult(ctx context.Context, id int) (model.TaskResult, error)
	ListTaskResults(ctx context.Context, filter ResultFilter, page int, size int) ([]model.TaskResult, int64, error)
	DeleteTaskResults(ctx context.Context, ids []int) error
}

type Seed struct {
	Schedulers []model.TaskScheduler
	Results    []model.TaskResult
}

func SeedData() Seed {
	return Seed{
		Schedulers: model.SeedSchedulers(),
		Results:    model.SeedTaskResults(),
	}
}
