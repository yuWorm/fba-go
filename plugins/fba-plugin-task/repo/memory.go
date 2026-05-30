package repo

import (
	"context"
	"strings"
	"sync"

	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/model"
)

type MemoryRepository struct {
	mu              sync.RWMutex
	schedulers      []model.TaskScheduler
	results         []model.TaskResult
	nextSchedulerID int
}

func NewMemoryRepository(seed Seed) *MemoryRepository {
	nextID := 1
	for _, item := range seed.Schedulers {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	return &MemoryRepository{
		schedulers:      append([]model.TaskScheduler(nil), seed.Schedulers...),
		results:         append([]model.TaskResult(nil), seed.Results...),
		nextSchedulerID: nextID,
	}
}

func (r *MemoryRepository) AllSchedulers(context.Context) ([]model.TaskScheduler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.TaskScheduler(nil), r.schedulers...), nil
}

func (r *MemoryRepository) GetScheduler(_ context.Context, id int) (model.TaskScheduler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.schedulers {
		if item.ID == id {
			return item, nil
		}
	}
	return model.TaskScheduler{}, ErrNotFound
}

func (r *MemoryRepository) ListSchedulers(_ context.Context, filter SchedulerFilter, page int, size int) ([]model.TaskScheduler, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.TaskScheduler, 0, len(r.schedulers))
	for _, item := range r.schedulers {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Type != nil && item.Type != *filter.Type {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateScheduler(_ context.Context, param dto.SchedulerParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schedulers = append(r.schedulers, schedulerFromParam(r.nextScheduler(), param))
	return nil
}

func (r *MemoryRepository) UpdateScheduler(_ context.Context, id int, param dto.SchedulerParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.schedulers {
		if r.schedulers[i].ID == id {
			updated := schedulerFromParam(id, param)
			updated.Enabled = r.schedulers[i].Enabled
			updated.TotalRunCount = r.schedulers[i].TotalRunCount
			updated.CreatedTime = r.schedulers[i].CreatedTime
			r.schedulers[i] = updated
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) ToggleSchedulerStatus(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.schedulers {
		if r.schedulers[i].ID == id {
			r.schedulers[i].Enabled = !r.schedulers[i].Enabled
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteScheduler(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.schedulers {
		if item.ID == id {
			r.schedulers = append(r.schedulers[:i], r.schedulers[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) GetTaskResult(_ context.Context, id int) (model.TaskResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.results {
		if item.ID == id {
			return item, nil
		}
	}
	return model.TaskResult{}, ErrNotFound
}

func (r *MemoryRepository) ListTaskResults(_ context.Context, filter ResultFilter, page int, size int) ([]model.TaskResult, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.TaskResult, 0, len(r.results))
	for _, item := range r.results {
		if filter.TaskID != "" && !strings.Contains(item.TaskID, filter.TaskID) {
			continue
		}
		if filter.Name != "" && (item.Name == nil || !strings.Contains(*item.Name, filter.Name)) {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) DeleteTaskResults(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	results := r.results[:0]
	for _, item := range r.results {
		if _, ok := idSet[item.ID]; !ok {
			results = append(results, item)
		}
	}
	r.results = results
	return nil
}

func (r *MemoryRepository) nextScheduler() int {
	id := r.nextSchedulerID
	r.nextSchedulerID++
	return id
}

func schedulerFromParam(id int, param dto.SchedulerParam) model.TaskScheduler {
	crontab := param.Crontab
	if crontab == "" {
		crontab = "* * * * *"
	}
	return model.TaskScheduler{
		ID:             id,
		Name:           param.Name,
		Task:           param.Task,
		Args:           dto.EncodeJSON(param.Args),
		Kwargs:         dto.EncodeJSON(param.Kwargs),
		Queue:          param.Queue,
		Exchange:       param.Exchange,
		RoutingKey:     param.RoutingKey,
		ExpireSeconds:  param.ExpireSeconds,
		Type:           param.Type,
		IntervalEvery:  param.IntervalEvery,
		IntervalPeriod: param.IntervalPeriod,
		Crontab:        crontab,
		OneOff:         param.OneOff,
		Remark:         param.Remark,
		Enabled:        true,
		CreatedTime:    model.SeedSchedulers()[0].CreatedTime,
	}
}

func pageSlice[T any](items []T, page int, size int) []T {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(items) {
		return []T{}
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return append([]T(nil), items[start:end]...)
}
