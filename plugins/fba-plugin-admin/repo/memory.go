package repo

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
)

var ErrNotFound = errors.New("not found")

type MemoryRepository struct {
	mu         sync.RWMutex
	roles      []model.Role
	menus      []model.Menu
	scopes     []model.DataScope
	roleMenus  map[int][]int
	roleScopes map[int][]int
	nextRoleID int
	nextMenuID int
}

func NewMemoryRepository(seed model.Seed) *MemoryRepository {
	nextRoleID := 1
	for _, item := range seed.Roles {
		if item.ID >= nextRoleID {
			nextRoleID = item.ID + 1
		}
	}
	nextMenuID := 1
	for _, item := range seed.Menus {
		if item.ID >= nextMenuID {
			nextMenuID = item.ID + 1
		}
	}
	return &MemoryRepository{
		roles:      append([]model.Role(nil), seed.Roles...),
		menus:      append([]model.Menu(nil), seed.Menus...),
		scopes:     append([]model.DataScope(nil), seed.DataScopes...),
		roleMenus:  cloneIDMap(seed.RoleMenus),
		roleScopes: cloneIDMap(seed.RoleScopes),
		nextRoleID: nextRoleID,
		nextMenuID: nextMenuID,
	}
}

func (r *MemoryRepository) AllRoles(context.Context) ([]model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Role(nil), r.roles...), nil
}

func (r *MemoryRepository) GetRole(_ context.Context, id int) (model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.roles {
		if item.ID == id {
			return item, nil
		}
	}
	return model.Role{}, ErrNotFound
}

func (r *MemoryRepository) ListRoles(_ context.Context, filter RoleFilter, page int, size int) ([]model.Role, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Role, 0, len(r.roles))
	for _, item := range r.roles {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateRole(_ context.Context, param dto.RoleParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles = append(r.roles, model.Role{
		ID:             r.nextRole(),
		Name:           param.Name,
		Status:         param.Status,
		IsFilterScopes: param.IsFilterScopes,
		Remark:         param.Remark,
		CreatedTime:    model.SeedData().Roles[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateRole(_ context.Context, id int, param dto.RoleParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.roles {
		if r.roles[i].ID == id {
			r.roles[i].Name = param.Name
			r.roles[i].Status = param.Status
			r.roles[i].IsFilterScopes = param.IsFilterScopes
			r.roles[i].Remark = param.Remark
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteRoles(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles = deleteByIDs(r.roles, ids, func(item model.Role) int { return item.ID })
	for _, id := range ids {
		delete(r.roleMenus, id)
		delete(r.roleScopes, id)
	}
	return nil
}

func (r *MemoryRepository) RoleMenus(_ context.Context, roleID int) ([]model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.hasRole(roleID) {
		return nil, ErrNotFound
	}
	return menusByIDs(r.menus, r.roleMenus[roleID]), nil
}

func (r *MemoryRepository) UpdateRoleMenus(_ context.Context, roleID int, menuIDs []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.hasRole(roleID) {
		return ErrNotFound
	}
	r.roleMenus[roleID] = filterKnownIDs(menuIDs, func(id int) bool {
		return hasMenu(r.menus, id)
	})
	return nil
}

func (r *MemoryRepository) RoleScopes(_ context.Context, roleID int) ([]model.DataScope, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.hasRole(roleID) {
		return nil, ErrNotFound
	}
	return scopesByIDs(r.scopes, r.roleScopes[roleID]), nil
}

func (r *MemoryRepository) RoleScopeIDs(_ context.Context, roleID int) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.hasRole(roleID) {
		return nil, ErrNotFound
	}
	return append([]int(nil), r.roleScopes[roleID]...), nil
}

func (r *MemoryRepository) UpdateRoleScopes(_ context.Context, roleID int, scopeIDs []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.hasRole(roleID) {
		return ErrNotFound
	}
	r.roleScopes[roleID] = filterKnownIDs(scopeIDs, func(id int) bool {
		return hasScope(r.scopes, id)
	})
	return nil
}

func (r *MemoryRepository) GetMenu(_ context.Context, id int) (model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.menus {
		if item.ID == id {
			return item, nil
		}
	}
	return model.Menu{}, ErrNotFound
}

func (r *MemoryRepository) ListMenus(_ context.Context, filter MenuFilter) ([]model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Menu, 0, len(r.menus))
	for _, item := range r.menus {
		if filter.Title != "" && !strings.Contains(item.Title, filter.Title) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	sortMenus(items)
	return items, nil
}

func (r *MemoryRepository) SidebarMenus(_ context.Context) ([]model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Menu, 0, len(r.menus))
	for _, item := range r.menus {
		if item.Type == 2 {
			continue
		}
		items = append(items, item)
	}
	sortMenus(items)
	return items, nil
}

func (r *MemoryRepository) CreateMenu(_ context.Context, param dto.MenuParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.menus = append(r.menus, model.Menu{
		ID:          r.nextMenu(),
		Title:       param.Title,
		Name:        param.Name,
		Path:        param.Path,
		ParentID:    param.ParentID,
		Sort:        param.Sort,
		Icon:        param.Icon,
		Type:        param.Type,
		Component:   param.Component,
		Perms:       param.Perms,
		Status:      param.Status,
		Display:     param.Display,
		Cache:       param.Cache,
		Link:        param.Link,
		Remark:      param.Remark,
		CreatedTime: model.SeedData().Menus[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateMenu(_ context.Context, id int, param dto.MenuParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.menus {
		if r.menus[i].ID == id {
			r.menus[i].Title = param.Title
			r.menus[i].Name = param.Name
			r.menus[i].Path = param.Path
			r.menus[i].ParentID = param.ParentID
			r.menus[i].Sort = param.Sort
			r.menus[i].Icon = param.Icon
			r.menus[i].Type = param.Type
			r.menus[i].Component = param.Component
			r.menus[i].Perms = param.Perms
			r.menus[i].Status = param.Status
			r.menus[i].Display = param.Display
			r.menus[i].Cache = param.Cache
			r.menus[i].Link = param.Link
			r.menus[i].Remark = param.Remark
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteMenu(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.menus = deleteByIDs(r.menus, []int{id}, func(item model.Menu) int { return item.ID })
	for roleID, menuIDs := range r.roleMenus {
		r.roleMenus[roleID] = deleteInt(menuIDs, id)
	}
	return nil
}

func (r *MemoryRepository) nextRole() int {
	id := r.nextRoleID
	r.nextRoleID++
	return id
}

func (r *MemoryRepository) nextMenu() int {
	id := r.nextMenuID
	r.nextMenuID++
	return id
}

func (r *MemoryRepository) hasRole(id int) bool {
	for _, item := range r.roles {
		if item.ID == id {
			return true
		}
	}
	return false
}

func cloneIDMap(source map[int][]int) map[int][]int {
	result := make(map[int][]int, len(source))
	for id, values := range source {
		result[id] = append([]int(nil), values...)
	}
	return result
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

func menusByIDs(items []model.Menu, ids []int) []model.Menu {
	result := make([]model.Menu, 0, len(ids))
	for _, id := range ids {
		for _, item := range items {
			if item.ID == id {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

func scopesByIDs(items []model.DataScope, ids []int) []model.DataScope {
	result := make([]model.DataScope, 0, len(ids))
	for _, id := range ids {
		for _, item := range items {
			if item.ID == id {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

func filterKnownIDs(ids []int, exists func(int) bool) []int {
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok || !exists(id) {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func hasMenu(items []model.Menu, id int) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasScope(items []model.DataScope, id int) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func sortMenus(items []model.Menu) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort != items[j].Sort {
			return items[i].Sort < items[j].Sort
		}
		return items[i].ID < items[j].ID
	})
}

func deleteInt(items []int, id int) []int {
	result := items[:0]
	for _, item := range items {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}
