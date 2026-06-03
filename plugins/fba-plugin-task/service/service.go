package service

import (
	"context"
	stderrors "errors"
	"net/http"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/pagination"
	coretask "github.com/yuWorm/fba-go/core/task"
	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/repo"
)

type Service struct {
	repo     repo.Repository
	registry *coretask.Registry
	executor Executor
	leader   LeaderLease
}

func New(repository repo.Repository, registry *coretask.Registry, executor Executor, leader LeaderLease) *Service {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	if executor == nil {
		executor = NoopExecutor{}
	}
	if leader == nil {
		leader = NoopLeaderLease{}
	}
	return &Service{repo: repository, registry: registry, executor: executor, leader: leader}
}

func (s *Service) RegisteredTasks() []dto.RegisteredTask {
	if s.registry == nil {
		return []dto.RegisteredTask{{Name: "task_demo", Task: "task_demo"}}
	}
	definitions := s.registry.All()
	items := make([]dto.RegisteredTask, 0, len(definitions))
	for _, definition := range definitions {
		name := definition.Name
		if name == "" {
			name = definition.Type
		}
		items = append(items, dto.RegisteredTask{Name: name, Task: definition.Type})
	}
	return items
}

func (s *Service) CancelTask(ctx context.Context, taskID string) error {
	return s.executor.Cancel(ctx, taskID)
}

func (s *Service) AllSchedulers(ctx context.Context) ([]dto.SchedulerDetail, error) {
	items, err := s.repo.AllSchedulers(ctx)
	if err != nil {
		return nil, err
	}
	return dto.SchedulersFromModel(items), nil
}

func (s *Service) GetScheduler(ctx context.Context, id int) (dto.SchedulerDetail, error) {
	item, err := s.repo.GetScheduler(ctx, id)
	if err != nil {
		return dto.SchedulerDetail{}, taskSchedulerNotFound(err)
	}
	return dto.SchedulerFromModel(item), nil
}

func (s *Service) ListSchedulers(ctx context.Context, filter repo.SchedulerFilter, page int, size int, basePath string) (pagination.PageData[dto.SchedulerDetail], error) {
	items, total, err := s.repo.ListSchedulers(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.SchedulerDetail]{}, err
	}
	return pagination.NewPageData(dto.SchedulersFromModel(items), total, page, size, basePath), nil
}

func (s *Service) CreateScheduler(ctx context.Context, param dto.SchedulerParam) error {
	if err := s.repo.CreateScheduler(ctx, param); err != nil {
		return err
	}
	return s.reloadIfLeader(ctx)
}

func (s *Service) UpdateScheduler(ctx context.Context, id int, param dto.SchedulerParam) error {
	if _, err := s.repo.GetScheduler(ctx, id); err != nil {
		return taskSchedulerNotFound(err)
	}
	if err := s.repo.UpdateScheduler(ctx, id, param); err != nil {
		return err
	}
	return s.reloadIfLeader(ctx)
}

func (s *Service) ToggleSchedulerStatus(ctx context.Context, id int) error {
	if _, err := s.repo.GetScheduler(ctx, id); err != nil {
		return taskSchedulerNotFound(err)
	}
	if err := s.repo.ToggleSchedulerStatus(ctx, id); err != nil {
		return err
	}
	return s.reloadIfLeader(ctx)
}

func (s *Service) DeleteScheduler(ctx context.Context, id int) error {
	if _, err := s.repo.GetScheduler(ctx, id); err != nil {
		return taskSchedulerNotFound(err)
	}
	if err := s.repo.DeleteScheduler(ctx, id); err != nil {
		return err
	}
	return s.reloadIfLeader(ctx)
}

func (s *Service) ExecuteScheduler(ctx context.Context, id int) error {
	scheduler, err := s.repo.GetScheduler(ctx, id)
	if err != nil {
		return taskSchedulerNotFound(err)
	}
	detail := dto.SchedulerFromModel(scheduler)
	return s.executor.Execute(ctx, detail.Task, detail.Args, detail.Kwargs)
}

func (s *Service) GetTaskResult(ctx context.Context, id int) (dto.TaskResultDetail, error) {
	item, err := s.repo.GetTaskResult(ctx, id)
	if err != nil {
		return dto.TaskResultDetail{}, taskResultNotFound(err)
	}
	return dto.TaskResultFromModel(item), nil
}

func (s *Service) ListTaskResults(ctx context.Context, filter repo.ResultFilter, page int, size int, basePath string) (pagination.PageData[dto.TaskResultDetail], error) {
	items, total, err := s.repo.ListTaskResults(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.TaskResultDetail]{}, err
	}
	return pagination.NewPageData(dto.TaskResultsFromModel(items), total, page, size, basePath), nil
}

func (s *Service) DeleteTaskResults(ctx context.Context, ids []int) error {
	return s.repo.DeleteTaskResults(ctx, ids)
}

func (s *Service) reloadIfLeader(ctx context.Context) error {
	acquired, err := s.leader.Acquire(ctx)
	if err != nil || !acquired {
		return err
	}
	defer func() {
		_ = s.leader.Release(ctx)
	}()
	return s.executor.Reload(ctx)
}

func taskSchedulerNotFound(err error) error {
	if stderrors.Is(err, repo.ErrNotFound) {
		return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, "任务调度不存在", err)
	}
	return err
}

func taskResultNotFound(err error) error {
	if stderrors.Is(err, repo.ErrNotFound) {
		return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, "任务结果不存在", err)
	}
	return err
}
