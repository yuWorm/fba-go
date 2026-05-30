package repo

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/yuWorm/fba-plugin-notice/dto"
	"github.com/yuWorm/fba-plugin-notice/model"
)

var ErrNotFound = errors.New("not found")

type MemoryRepository struct {
	mu      sync.RWMutex
	notices []model.Notice
	nextID  int
}

func NewMemoryRepository(seed Seed) *MemoryRepository {
	nextID := 1
	for _, item := range seed.Notices {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	return &MemoryRepository{
		notices: append([]model.Notice(nil), seed.Notices...),
		nextID:  nextID,
	}
}

func (r *MemoryRepository) Get(_ context.Context, id int) (model.Notice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.notices {
		if item.ID == id {
			return item, nil
		}
	}
	return model.Notice{}, ErrNotFound
}

func (r *MemoryRepository) List(_ context.Context, filter NoticeFilter, page int, size int) ([]model.Notice, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Notice, 0, len(r.notices))
	for _, item := range r.notices {
		if filter.Title != "" && !strings.Contains(item.Title, filter.Title) {
			continue
		}
		if filter.Type != nil && item.Type != *filter.Type {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) Create(_ context.Context, param dto.NoticeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notices = append(r.notices, model.Notice{
		ID:          r.next(),
		Title:       param.Title,
		Type:        param.Type,
		Status:      param.Status,
		Content:     param.Content,
		CreatedTime: model.SeedNotices()[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) Update(_ context.Context, id int, param dto.NoticeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.notices {
		if r.notices[i].ID == id {
			r.notices[i].Title = param.Title
			r.notices[i].Type = param.Type
			r.notices[i].Status = param.Status
			r.notices[i].Content = param.Content
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) Delete(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notices = deleteByIDs(r.notices, ids, func(item model.Notice) int { return item.ID })
	return nil
}

func (r *MemoryRepository) next() int {
	id := r.nextID
	r.nextID++
	return id
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

func deleteByIDs[T any](items []T, ids []int, idFunc func(T) int) []T {
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	result := items[:0]
	for _, item := range items {
		if _, ok := idSet[idFunc(item)]; !ok {
			result = append(result, item)
		}
	}
	return result
}
