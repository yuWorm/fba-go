package repo

import (
	"context"
	"sync"
	"time"

	"github.com/yuWorm/fba-plugin-oauth2/model"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	items  []model.UserSocial
	nextID int
}

func NewMemoryRepository(seed []model.UserSocial) *MemoryRepository {
	nextID := 1
	for _, item := range seed {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	return &MemoryRepository{
		items:  append([]model.UserSocial(nil), seed...),
		nextID: nextID,
	}
}

func (r *MemoryRepository) GetBySID(_ context.Context, sid string, source string) (model.UserSocial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.items {
		if item.SID == sid && item.Source == source && item.Deleted == 0 {
			return item, nil
		}
	}
	return model.UserSocial{}, ErrNotFound
}

func (r *MemoryRepository) CheckBinding(_ context.Context, userID int, source string) (model.UserSocial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.items {
		if item.UserID == userID && item.Source == source && item.Deleted == 0 {
			return item, nil
		}
	}
	return model.UserSocial{}, ErrNotFound
}

func (r *MemoryRepository) ListByUserID(_ context.Context, userID int) ([]model.UserSocial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.UserSocial, 0)
	for _, item := range r.items {
		if item.UserID == userID && item.Deleted == 0 {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *MemoryRepository) Create(_ context.Context, param CreateUserSocialParam) (model.UserSocial, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	item := model.UserSocial{
		ID:          r.next(),
		SID:         param.SID,
		Source:      param.Source,
		UserID:      param.UserID,
		Deleted:     0,
		CreatedTime: now,
		UpdatedTime: now,
	}
	r.items = append(r.items, item)
	return item, nil
}

func (r *MemoryRepository) Delete(_ context.Context, userID int, source string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i := range r.items {
		if r.items[i].UserID != userID || r.items[i].Source != source || r.items[i].Deleted != 0 {
			continue
		}
		// Python logical deletion writes a non-zero deleted flag derived from the row id.
		r.items[i].Deleted = r.items[i].ID
		r.items[i].DeletedTime = &now
		r.items[i].UpdatedTime = now
		return 1, nil
	}
	return 0, ErrNotFound
}

func (r *MemoryRepository) next() int {
	id := r.nextID
	r.nextID++
	return id
}
