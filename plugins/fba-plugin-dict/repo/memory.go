package repo

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/model"
)

var ErrNotFound = errors.New("not found")

type MemoryRepository struct {
	mu     sync.RWMutex
	types  []model.DictType
	data   []model.DictData
	nextID int
}

func NewMemoryRepository(seed Seed) *MemoryRepository {
	nextID := 1
	for _, item := range seed.Types {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	for _, item := range seed.Data {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	return &MemoryRepository{
		types:  append([]model.DictType(nil), seed.Types...),
		data:   append([]model.DictData(nil), seed.Data...),
		nextID: nextID,
	}
}

func (r *MemoryRepository) AllTypes(context.Context) ([]model.DictType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.DictType(nil), r.types...), nil
}

func (r *MemoryRepository) GetType(_ context.Context, id int) (model.DictType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.types {
		if item.ID == id {
			return item, nil
		}
	}
	return model.DictType{}, ErrNotFound
}

func (r *MemoryRepository) ListTypes(_ context.Context, filter DictTypeFilter, page int, size int) ([]model.DictType, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.DictType, 0, len(r.types))
	for _, item := range r.types {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Code != "" && !strings.Contains(item.Code, filter.Code) {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateType(_ context.Context, param dto.DictTypeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types = append(r.types, model.DictType{
		ID:          r.next(),
		Name:        param.Name,
		Code:        param.Code,
		Remark:      param.Remark,
		CreatedTime: model.SeedDictTypes()[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateType(_ context.Context, id int, param dto.DictTypeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.types {
		if r.types[i].ID == id {
			r.types[i].Name = param.Name
			r.types[i].Code = param.Code
			r.types[i].Remark = param.Remark
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteTypes(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types = deleteByIDs(r.types, ids, func(item model.DictType) int { return item.ID })
	return nil
}

func (r *MemoryRepository) AllData(context.Context) ([]model.DictData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.DictData(nil), r.data...), nil
}

func (r *MemoryRepository) GetData(_ context.Context, id int) (model.DictData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.data {
		if item.ID == id {
			return item, nil
		}
	}
	return model.DictData{}, ErrNotFound
}

func (r *MemoryRepository) DataByTypeCode(_ context.Context, code string) ([]model.DictData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.DictData, 0, len(r.data))
	for _, item := range r.data {
		if item.TypeCode == code {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items, nil
}

func (r *MemoryRepository) ListData(_ context.Context, filter DictDataFilter, page int, size int) ([]model.DictData, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.DictData, 0, len(r.data))
	for _, item := range r.data {
		if filter.TypeCode != "" && !strings.Contains(item.TypeCode, filter.TypeCode) {
			continue
		}
		if filter.Label != "" && !strings.Contains(item.Label, filter.Label) {
			continue
		}
		if filter.Value != "" && !strings.Contains(item.Value, filter.Value) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		if filter.TypeID != nil && item.TypeID != *filter.TypeID {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateData(_ context.Context, param dto.DictDataParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	typeCode, err := r.typeCode(param.TypeID)
	if err != nil {
		return err
	}
	r.data = append(r.data, model.DictData{
		ID:          r.next(),
		TypeID:      param.TypeID,
		TypeCode:    typeCode,
		Label:       param.Label,
		Value:       param.Value,
		Color:       param.Color,
		Sort:        param.Sort,
		Status:      param.Status,
		Remark:      param.Remark,
		CreatedTime: model.SeedDictData()[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateData(_ context.Context, id int, param dto.DictDataParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	typeCode, err := r.typeCode(param.TypeID)
	if err != nil {
		return err
	}
	for i := range r.data {
		if r.data[i].ID == id {
			r.data[i].TypeID = param.TypeID
			r.data[i].TypeCode = typeCode
			r.data[i].Label = param.Label
			r.data[i].Value = param.Value
			r.data[i].Color = param.Color
			r.data[i].Sort = param.Sort
			r.data[i].Status = param.Status
			r.data[i].Remark = param.Remark
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteData(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = deleteByIDs(r.data, ids, func(item model.DictData) int { return item.ID })
	return nil
}

func (r *MemoryRepository) next() int {
	id := r.nextID
	r.nextID++
	return id
}

func (r *MemoryRepository) typeCode(typeID int) (string, error) {
	for _, item := range r.types {
		if item.ID == typeID {
			return item.Code, nil
		}
	}
	return "", ErrNotFound
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
