package repo

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
)

var ErrNotFound = errors.New("not found")

type MemoryRepository struct {
	mu              sync.RWMutex
	users           []model.User
	roles           []model.Role
	menus           []model.Menu
	depts           []model.Dept
	dataRules       []model.DataRule
	scopes          []model.DataScope
	userRoles       map[int][]int
	scopeRules      map[int][]int
	roleMenus       map[int][]int
	roleScopes      map[int][]int
	nextUserID      int
	nextRoleID      int
	nextMenuID      int
	nextDeptID      int
	nextDataRuleID  int
	nextDataScopeID int
}

func NewMemoryRepository(seed model.Seed) *MemoryRepository {
	nextUserID := 1
	for _, item := range seed.Users {
		if item.ID >= nextUserID {
			nextUserID = item.ID + 1
		}
	}
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
	nextDeptID := 1
	for _, item := range seed.Depts {
		if item.ID >= nextDeptID {
			nextDeptID = item.ID + 1
		}
	}
	nextDataRuleID := 1
	for _, item := range seed.DataRules {
		if item.ID >= nextDataRuleID {
			nextDataRuleID = item.ID + 1
		}
	}
	nextDataScopeID := 1
	for _, item := range seed.DataScopes {
		if item.ID >= nextDataScopeID {
			nextDataScopeID = item.ID + 1
		}
	}
	return &MemoryRepository{
		users:           append([]model.User(nil), seed.Users...),
		roles:           append([]model.Role(nil), seed.Roles...),
		menus:           append([]model.Menu(nil), seed.Menus...),
		depts:           append([]model.Dept(nil), seed.Depts...),
		dataRules:       append([]model.DataRule(nil), seed.DataRules...),
		scopes:          append([]model.DataScope(nil), seed.DataScopes...),
		userRoles:       cloneIDMap(seed.UserRoles),
		scopeRules:      cloneIDMap(seed.ScopeRules),
		roleMenus:       cloneIDMap(seed.RoleMenus),
		roleScopes:      cloneIDMap(seed.RoleScopes),
		nextUserID:      nextUserID,
		nextRoleID:      nextRoleID,
		nextMenuID:      nextMenuID,
		nextDeptID:      nextDeptID,
		nextDataRuleID:  nextDataRuleID,
		nextDataScopeID: nextDataScopeID,
	}
}

func (r *MemoryRepository) GetUser(_ context.Context, id int) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.users {
		if item.ID == id {
			return item, nil
		}
	}
	return model.User{}, ErrNotFound
}

func (r *MemoryRepository) GetUserByUsername(_ context.Context, username string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.users {
		if item.Username == username {
			return item, nil
		}
	}
	return model.User{}, ErrNotFound
}

func (r *MemoryRepository) ListUsers(_ context.Context, filter UserFilter, page int, size int) ([]model.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.User, 0, len(r.users))
	for _, item := range r.users {
		if filter.Dept != nil && (item.DeptID == nil || *item.DeptID != *filter.Dept) {
			continue
		}
		if filter.Username != "" && !strings.Contains(item.Username, filter.Username) {
			continue
		}
		if filter.Phone != "" && (item.Phone == nil || !strings.Contains(*item.Phone, filter.Phone)) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	sortUsers(items)
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateUser(_ context.Context, param dto.UserCreateParam) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	deptID := param.DeptID
	if !hasDept(r.depts, deptID) {
		return model.User{}, ErrNotFound
	}
	for _, roleID := range param.Roles {
		if !hasRole(r.roles, roleID) {
			return model.User{}, ErrNotFound
		}
	}
	nickname := param.Username
	if param.Nickname != nil && *param.Nickname != "" {
		nickname = *param.Nickname
	}
	id := r.nextUser()
	user := model.User{
		ID:           id,
		UUID:         "fixture-user-" + strconv.Itoa(id),
		DeptID:       &deptID,
		Username:     param.Username,
		Nickname:     nickname,
		Password:     param.Password,
		Email:        param.Email,
		Phone:        param.Phone,
		Status:       1,
		IsSuperuser:  false,
		IsStaff:      false,
		IsMultiLogin: false,
		JoinTime:     model.SeedData().Users[0].JoinTime,
	}
	r.users = append(r.users, user)
	r.userRoles[id] = append([]int(nil), param.Roles...)
	return user, nil
}

func (r *MemoryRepository) UpdateUser(_ context.Context, id int, param dto.UserUpdateParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if param.DeptID != nil && !hasDept(r.depts, *param.DeptID) {
		return ErrNotFound
	}
	for _, roleID := range param.Roles {
		if !hasRole(r.roles, roleID) {
			return ErrNotFound
		}
	}
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].DeptID = param.DeptID
			r.users[i].Username = param.Username
			r.users[i].Nickname = param.Nickname
			r.users[i].Avatar = param.Avatar
			r.users[i].Email = param.Email
			r.users[i].Phone = param.Phone
			r.userRoles[id] = append([]int(nil), param.Roles...)
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) UpdateUserNickname(_ context.Context, id int, nickname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].Nickname = nickname
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) UpdateUserAvatar(_ context.Context, id int, avatar *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].Avatar = avatar
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) UpdateUserEmail(_ context.Context, id int, email *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].Email = email
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) ResetUserPassword(_ context.Context, id int, password string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].Password = password
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) UpdateUserPermission(_ context.Context, id int, permissionType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.users {
		if r.users[i].ID == id {
			switch permissionType {
			case "superuser":
				r.users[i].IsSuperuser = !r.users[i].IsSuperuser
			case "staff":
				r.users[i].IsStaff = !r.users[i].IsStaff
			case "status":
				if r.users[i].Status == 1 {
					r.users[i].Status = 0
				} else {
					r.users[i].Status = 1
				}
			case "multi_login":
				r.users[i].IsMultiLogin = !r.users[i].IsMultiLogin
			default:
				return ErrNotFound
			}
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteUser(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = deleteByIDs(r.users, []int{id}, func(item model.User) int { return item.ID })
	delete(r.userRoles, id)
	return nil
}

func (r *MemoryRepository) UserRoles(_ context.Context, userID int) ([]model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !hasUser(r.users, userID) {
		return nil, ErrNotFound
	}
	return rolesByIDs(r.roles, r.userRoles[userID]), nil
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
		for userID, roleIDs := range r.userRoles {
			r.userRoles[userID] = deleteInt(roleIDs, id)
		}
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

func (r *MemoryRepository) GetDept(_ context.Context, id int) (model.Dept, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.depts {
		if item.ID == id {
			return item, nil
		}
	}
	return model.Dept{}, ErrNotFound
}

func (r *MemoryRepository) ListDepts(_ context.Context, filter DeptFilter) ([]model.Dept, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.Dept, 0, len(r.depts))
	for _, item := range r.depts {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Leader != "" && (item.Leader == nil || !strings.Contains(*item.Leader, filter.Leader)) {
			continue
		}
		if filter.Phone != "" && (item.Phone == nil || !strings.HasPrefix(*item.Phone, filter.Phone)) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	sortDepts(items)
	return items, nil
}

func (r *MemoryRepository) CreateDept(_ context.Context, param dto.DeptParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.depts = append(r.depts, model.Dept{
		ID:          r.nextDept(),
		Name:        param.Name,
		ParentID:    param.ParentID,
		Sort:        param.Sort,
		Leader:      param.Leader,
		Phone:       param.Phone,
		Email:       param.Email,
		Status:      param.Status,
		Deleted:     0,
		CreatedTime: model.SeedData().Depts[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateDept(_ context.Context, id int, param dto.DeptParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.depts {
		if r.depts[i].ID == id {
			r.depts[i].Name = param.Name
			r.depts[i].ParentID = param.ParentID
			r.depts[i].Sort = param.Sort
			r.depts[i].Leader = param.Leader
			r.depts[i].Phone = param.Phone
			r.depts[i].Email = param.Email
			r.depts[i].Status = param.Status
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteDept(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.depts = deleteByIDs(r.depts, []int{id}, func(item model.Dept) int { return item.ID })
	return nil
}

func (r *MemoryRepository) AllDataRules(context.Context) ([]model.DataRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.DataRule(nil), r.dataRules...), nil
}

func (r *MemoryRepository) GetDataRule(_ context.Context, id int) (model.DataRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.dataRules {
		if item.ID == id {
			return item, nil
		}
	}
	return model.DataRule{}, ErrNotFound
}

func (r *MemoryRepository) ListDataRules(_ context.Context, filter DataRuleFilter, page int, size int) ([]model.DataRule, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.DataRule, 0, len(r.dataRules))
	for _, item := range r.dataRules {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		items = append(items, item)
	}
	sortDataRules(items)
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateDataRule(_ context.Context, param dto.DataRuleParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataRules = append(r.dataRules, model.DataRule{
		ID:          r.nextDataRule(),
		Name:        param.Name,
		Model:       param.Model,
		Column:      param.Column,
		Operator:    param.Operator,
		Expression:  param.Expression,
		Value:       param.Value,
		CreatedTime: model.SeedData().DataRules[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateDataRule(_ context.Context, id int, param dto.DataRuleParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.dataRules {
		if r.dataRules[i].ID == id {
			r.dataRules[i].Name = param.Name
			r.dataRules[i].Model = param.Model
			r.dataRules[i].Column = param.Column
			r.dataRules[i].Operator = param.Operator
			r.dataRules[i].Expression = param.Expression
			r.dataRules[i].Value = param.Value
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) DeleteDataRules(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataRules = deleteByIDs(r.dataRules, ids, func(item model.DataRule) int { return item.ID })
	return nil
}

func (r *MemoryRepository) AllDataScopes(context.Context) ([]model.DataScope, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]model.DataScope(nil), r.scopes...)
	sortDataScopes(items)
	return items, nil
}

func (r *MemoryRepository) GetDataScope(_ context.Context, id int) (model.DataScope, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.scopes {
		if item.ID == id {
			return item, nil
		}
	}
	return model.DataScope{}, ErrNotFound
}

func (r *MemoryRepository) DataScopeRules(_ context.Context, id int) (model.DataScope, []model.DataRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.scopes {
		if item.ID == id {
			return item, dataRulesByIDs(r.dataRules, r.scopeRules[id]), nil
		}
	}
	return model.DataScope{}, nil, ErrNotFound
}

func (r *MemoryRepository) ListDataScopes(_ context.Context, filter DataScopeFilter, page int, size int) ([]model.DataScope, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.DataScope, 0, len(r.scopes))
	for _, item := range r.scopes {
		if filter.Name != "" && !strings.Contains(item.Name, filter.Name) {
			continue
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	sortDataScopes(items)
	return pageSlice(items, page, size), int64(len(items)), nil
}

func (r *MemoryRepository) CreateDataScope(_ context.Context, param dto.DataScopeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scopes = append(r.scopes, model.DataScope{
		ID:          r.nextDataScope(),
		Name:        param.Name,
		Status:      param.Status,
		CreatedTime: model.SeedData().DataScopes[0].CreatedTime,
	})
	return nil
}

func (r *MemoryRepository) UpdateDataScope(_ context.Context, id int, param dto.DataScopeParam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.scopes {
		if r.scopes[i].ID == id {
			r.scopes[i].Name = param.Name
			r.scopes[i].Status = param.Status
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) UpdateDataScopeRules(_ context.Context, id int, ruleIDs []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !hasScope(r.scopes, id) {
		return ErrNotFound
	}
	for _, ruleID := range ruleIDs {
		if !hasDataRule(r.dataRules, ruleID) {
			return ErrNotFound
		}
	}
	r.scopeRules[id] = append([]int(nil), ruleIDs...)
	return nil
}

func (r *MemoryRepository) DeleteDataScopes(_ context.Context, ids []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scopes = deleteByIDs(r.scopes, ids, func(item model.DataScope) int { return item.ID })
	for _, id := range ids {
		delete(r.scopeRules, id)
		for roleID, scopeIDs := range r.roleScopes {
			r.roleScopes[roleID] = deleteInt(scopeIDs, id)
		}
	}
	return nil
}

func (r *MemoryRepository) nextRole() int {
	id := r.nextRoleID
	r.nextRoleID++
	return id
}

func (r *MemoryRepository) nextUser() int {
	id := r.nextUserID
	r.nextUserID++
	return id
}

func (r *MemoryRepository) nextMenu() int {
	id := r.nextMenuID
	r.nextMenuID++
	return id
}

func (r *MemoryRepository) nextDept() int {
	id := r.nextDeptID
	r.nextDeptID++
	return id
}

func (r *MemoryRepository) nextDataRule() int {
	id := r.nextDataRuleID
	r.nextDataRuleID++
	return id
}

func (r *MemoryRepository) nextDataScope() int {
	id := r.nextDataScopeID
	r.nextDataScopeID++
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

func rolesByIDs(items []model.Role, ids []int) []model.Role {
	result := make([]model.Role, 0, len(ids))
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

func dataRulesByIDs(items []model.DataRule, ids []int) []model.DataRule {
	result := make([]model.DataRule, 0, len(ids))
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

func hasUser(items []model.User, id int) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasRole(items []model.Role, id int) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasDept(items []model.Dept, id int) bool {
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

func hasDataRule(items []model.DataRule, id int) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func sortUsers(items []model.User) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ID > items[j].ID
	})
}

func sortMenus(items []model.Menu) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort != items[j].Sort {
			return items[i].Sort < items[j].Sort
		}
		return items[i].ID < items[j].ID
	})
}

func sortDepts(items []model.Dept) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort != items[j].Sort {
			return items[i].Sort < items[j].Sort
		}
		return items[i].ID < items[j].ID
	})
}

func sortDataRules(items []model.DataRule) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
}

func sortDataScopes(items []model.DataScope) {
	sort.SliceStable(items, func(i, j int) bool {
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
