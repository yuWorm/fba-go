package repo

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/yuWorm/fba-plugin-config/dto"
	"github.com/yuWorm/fba-plugin-config/model"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	configs []model.Config
	nextID  int
}

func NewMemoryRepository(seed Seed) *MemoryRepository {
	nextID := 1
	for _, item := range seed.Configs {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	return &MemoryRepository{
		configs: append([]model.Config(nil), seed.Configs...),
		nextID:  nextID,
	}
}

func (r *MemoryRepository) All(_ context.Context, typeName string) ([]model.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.Config, 0, len(r.configs))
	for _, item := range r.configs {
		if item.Deleted != 0 {
			continue
		}
		if typeName != "" && (item.Type == nil || *item.Type != typeName) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MemoryRepository) Get(_ context.Context, id int) (model.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.configs {
		if item.ID == id && item.Deleted == 0 {
			return item, nil
		}
	}
	return model.Config{}, ErrNotFound
}

func (r *MemoryRepository) GetByKey(_ context.Context, key string) (model.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.configs {
		if item.Key == key && item.Deleted == 0 {
			return item, nil
		}
	}
	return model.Config{}, ErrNotFound
}

func (r *MemoryRepository) List(_ context.Context, filter ConfigFilter, page int, size int) ([]model.Config, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.Config, 0, len(r.configs))
	for _, item := range r.configs {
		if item.Deleted != 0 {
			continue
		}
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Type != "" && (item.Type == nil || *item.Type != filter.Type) {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) Create(_ context.Context, param dto.ConfigParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs = append(r.configs, model.Config{
		ID:          r.next(),
		Name:        param.Name,
		Type:        cloneString(param.Type),
		Key:         param.Key,
		Value:       param.Value,
		IsFrontend:  param.IsFrontend,
		Remark:      cloneString(param.Remark),
		CreatedTime: model.SeedConfigs()[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) Update(_ context.Context, id int, param dto.ConfigParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.configs {
		if r.configs[i].ID == id && r.configs[i].Deleted == 0 {
			r.applyParam(i, param)
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) BulkUpdate(_ context.Context, params []dto.ConfigBulkParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, param := range params {
		found := false
		for i := range r.configs {
			if r.configs[i].ID == param.ID && r.configs[i].Deleted == 0 {
				r.applyParam(i, param.ConfigParam)
				found = true
				break
			}
		}
		if !found {
			return ErrNotFound
		}
	}
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	deleted := 0
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for i := range r.configs {
		if _, ok := idSet[r.configs[i].ID]; ok && r.configs[i].Deleted == 0 {
			r.configs[i].Deleted = r.configs[i].ID
			r.configs[i].DeletedTime = &now
			deleted++
		}
	}
	if deleted == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MemoryRepository) next() int {
	id := r.nextID
	r.nextID++
	return id
}

func (r *MemoryRepository) applyParam(index int, param dto.ConfigParam) {
	r.configs[index].Name = param.Name
	r.configs[index].Type = cloneString(param.Type)
	r.configs[index].Key = param.Key
	r.configs[index].Value = param.Value
	r.configs[index].IsFrontend = param.IsFrontend
	r.configs[index].Remark = cloneString(param.Remark)
	now := time.Now()
	r.configs[index].UpdatedTime = &now
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
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
