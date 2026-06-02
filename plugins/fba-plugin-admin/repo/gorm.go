package repo

import (
	"context"
	"strconv"
	"time"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"gorm.io/gorm"
)

const pluginChangedStateKey = "plugins_changed"

type UserRole struct {
	ID     int `gorm:"column:id;primaryKey"`
	UserID int `gorm:"column:user_id;uniqueIndex:idx_sys_user_role"`
	RoleID int `gorm:"column:role_id;uniqueIndex:idx_sys_user_role"`
}

func (UserRole) TableName() string {
	return "sys_user_role"
}

type RoleMenu struct {
	ID     int `gorm:"column:id;primaryKey"`
	RoleID int `gorm:"column:role_id;uniqueIndex:idx_sys_role_menu"`
	MenuID int `gorm:"column:menu_id;uniqueIndex:idx_sys_role_menu"`
}

func (RoleMenu) TableName() string {
	return "sys_role_menu"
}

type RoleDataScope struct {
	ID          int `gorm:"column:id;primaryKey"`
	RoleID      int `gorm:"column:role_id;uniqueIndex:idx_sys_role_data_scope"`
	DataScopeID int `gorm:"column:data_scope_id;uniqueIndex:idx_sys_role_data_scope"`
}

func (RoleDataScope) TableName() string {
	return "sys_role_data_scope"
}

type DataScopeRule struct {
	ID          int `gorm:"column:id;primaryKey"`
	DataScopeID int `gorm:"column:data_scope_id;uniqueIndex:idx_sys_data_scope_rule"`
	DataRuleID  int `gorm:"column:data_rule_id;uniqueIndex:idx_sys_data_scope_rule"`
}

func (DataScopeRule) TableName() string {
	return "sys_data_scope_rule"
}

type PluginState struct {
	Key   string `gorm:"column:key;primaryKey;size:64"`
	Value bool   `gorm:"column:value"`
}

func (PluginState) TableName() string {
	return "sys_plugin_state"
}

type GORMRepository struct {
	provider db.Provider
	seed     model.Seed
	fallback *MemoryRepository
}

func NewGORMRepository(provider db.Provider, seeds ...model.Seed) *GORMRepository {
	seed := SeedData()
	if len(seeds) > 0 {
		seed = seeds[0]
	}
	return &GORMRepository{
		provider: provider,
		seed:     seed,
		fallback: NewMemoryRepository(seed),
	}
}

func (r *GORMRepository) Seed(ctx context.Context) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedIfEmpty(tx, &model.User{}, r.seed.Users); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.Role{}, r.seed.Roles); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.Menu{}, r.seed.Menus); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.Dept{}, r.seed.Depts); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.DataRule{}, r.seed.DataRules); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.DataScope{}, r.seed.DataScopes); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.Plugin{}, r.seed.Plugins); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.LoginLog{}, r.seed.LoginLogs); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.OperaLog{}, r.seed.OperaLogs); err != nil {
			return err
		}
		if err := seedIfEmpty(tx, &model.Session{}, r.seed.Sessions); err != nil {
			return err
		}
		if err := seedUserRoles(tx, r.seed.UserRoles); err != nil {
			return err
		}
		if err := seedRoleMenus(tx, r.seed.RoleMenus); err != nil {
			return err
		}
		if err := seedRoleDataScopes(tx, r.seed.RoleScopes); err != nil {
			return err
		}
		return seedDataScopeRules(tx, r.seed.ScopeRules)
	})
}

func (r *GORMRepository) GetUser(ctx context.Context, id int) (model.User, error) {
	var item model.User
	err := r.provider.Read().WithContext(ctx).Where("id = ? AND deleted = ?", id, 0).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	var item model.User
	err := r.provider.Read().WithContext(ctx).Where("username = ? AND deleted = ?", username, 0).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	err := r.provider.Read().WithContext(ctx).Where("email = ? AND deleted = ?", email, 0).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListUsers(ctx context.Context, filter UserFilter, page int, size int) ([]model.User, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.User{}).Where("deleted = ?", 0)
	if filter.Dept != nil {
		query = query.Where("dept_id = ?", *filter.Dept)
	}
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+filter.Phone+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return paginateGORM[model.User](query.Order("id DESC"), page, size)
}

func (r *GORMRepository) CreateUser(ctx context.Context, param dto.UserCreateParam) (model.User, error) {
	if _, err := r.GetDept(ctx, param.DeptID); err != nil {
		return model.User{}, err
	}
	if err := r.ensureRoles(ctx, param.Roles); err != nil {
		return model.User{}, err
	}
	nickname := param.Username
	if param.Nickname != nil && *param.Nickname != "" {
		nickname = *param.Nickname
	}
	user := model.User{
		UUID:         "gorm-user-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		DeptID:       &param.DeptID,
		Username:     param.Username,
		Nickname:     nickname,
		Password:     param.Password,
		Email:        param.Email,
		Phone:        param.Phone,
		Status:       1,
		IsSuperuser:  false,
		IsStaff:      false,
		IsMultiLogin: false,
		Deleted:      0,
		JoinTime:     time.Now(),
	}
	err := r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return replaceUserRoles(tx, user.ID, param.Roles)
	})
	return user, err
}

func (r *GORMRepository) UpdateUser(ctx context.Context, id int, param dto.UserUpdateParam) error {
	if param.DeptID != nil {
		if _, err := r.GetDept(ctx, *param.DeptID); err != nil {
			return err
		}
	}
	if err := r.ensureRoles(ctx, param.Roles); err != nil {
		return err
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).Where("id = ? AND deleted = ?", id, 0).Updates(map[string]any{
			"dept_id":  param.DeptID,
			"username": param.Username,
			"nickname": param.Nickname,
			"avatar":   param.Avatar,
			"email":    param.Email,
			"phone":    param.Phone,
		})
		if err := rowsError(result); err != nil {
			return err
		}
		return replaceUserRoles(tx, id, param.Roles)
	})
}

func (r *GORMRepository) UpdateUserNickname(ctx context.Context, id int, nickname string) error {
	return updateUserColumns(r.provider.Write().WithContext(ctx), id, map[string]any{"nickname": nickname})
}

func (r *GORMRepository) UpdateUserAvatar(ctx context.Context, id int, avatar *string) error {
	return updateUserColumns(r.provider.Write().WithContext(ctx), id, map[string]any{"avatar": avatar})
}

func (r *GORMRepository) UpdateUserEmail(ctx context.Context, id int, email *string) error {
	return updateUserColumns(r.provider.Write().WithContext(ctx), id, map[string]any{"email": email})
}

func (r *GORMRepository) ResetUserPassword(ctx context.Context, id int, password string) error {
	return updateUserColumns(r.provider.Write().WithContext(ctx), id, map[string]any{"password": password})
}

func (r *GORMRepository) UpdateUserPermission(ctx context.Context, id int, permissionType string) error {
	user, err := r.GetUser(ctx, id)
	if err != nil {
		return err
	}
	updates := map[string]any{}
	switch permissionType {
	case "superuser":
		updates["is_superuser"] = !user.IsSuperuser
	case "staff":
		updates["is_staff"] = !user.IsStaff
	case "status":
		if user.Status == 1 {
			updates["status"] = 0
		} else {
			updates["status"] = 1
		}
	case "multi_login":
		updates["is_multi_login"] = !user.IsMultiLogin
	default:
		return ErrNotFound
	}
	return updateUserColumns(r.provider.Write().WithContext(ctx), id, updates)
}

func (r *GORMRepository) DeleteUser(ctx context.Context, id int) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&UserRole{}).Error; err != nil {
			return err
		}
		result := tx.Model(&model.User{}).Where("id = ? AND deleted = ?", id, 0).Updates(map[string]any{
			"deleted":      id,
			"deleted_time": time.Now(),
		})
		return rowsError(result)
	})
}

func (r *GORMRepository) UserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	if _, err := r.GetUser(ctx, userID); err != nil {
		return nil, err
	}
	var joins []UserRole
	if err := r.provider.Read().WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&joins).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(joins))
	for _, item := range joins {
		ids = append(ids, item.RoleID)
	}
	return r.rolesByIDs(ctx, ids)
}

func (r *GORMRepository) AllRoles(ctx context.Context) ([]model.Role, error) {
	var items []model.Role
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetRole(ctx context.Context, id int) (model.Role, error) {
	var item model.Role
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetRoleByName(ctx context.Context, name string) (model.Role, error) {
	var item model.Role
	err := r.provider.Read().WithContext(ctx).Where("name = ?", name).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListRoles(ctx context.Context, filter RoleFilter, page int, size int) ([]model.Role, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Role{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return paginateGORM[model.Role](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) CreateRole(ctx context.Context, param dto.RoleParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.Role{
		Name:           param.Name,
		Status:         param.Status,
		IsFilterScopes: param.IsFilterScopes,
		Remark:         param.Remark,
		CreatedTime:    time.Now(),
	}).Error
}

func (r *GORMRepository) UpdateRole(ctx context.Context, id int, param dto.RoleParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Role{}).Where("id = ?", id).Updates(map[string]any{
		"name":             param.Name,
		"status":           param.Status,
		"is_filter_scopes": param.IsFilterScopes,
		"remark":           param.Remark,
	})
	return rowsError(result)
}

func (r *GORMRepository) DeleteRoles(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id IN ?", ids).Delete(&RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&RoleDataScope{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, ids).Error
	})
}

func (r *GORMRepository) RoleMenus(ctx context.Context, roleID int) ([]model.Menu, error) {
	if _, err := r.GetRole(ctx, roleID); err != nil {
		return nil, err
	}
	var joins []RoleMenu
	if err := r.provider.Read().WithContext(ctx).Where("role_id = ?", roleID).Order("id ASC").Find(&joins).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(joins))
	for _, item := range joins {
		ids = append(ids, item.MenuID)
	}
	return r.menusByIDs(ctx, ids)
}

func (r *GORMRepository) UpdateRoleMenus(ctx context.Context, roleID int, menuIDs []int) error {
	if _, err := r.GetRole(ctx, roleID); err != nil {
		return err
	}
	knownIDs, err := r.knownMenuIDs(ctx, menuIDs)
	if err != nil {
		return err
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceRoleMenus(tx, roleID, knownIDs)
	})
}

func (r *GORMRepository) RoleScopes(ctx context.Context, roleID int) ([]model.DataScope, error) {
	if _, err := r.GetRole(ctx, roleID); err != nil {
		return nil, err
	}
	var joins []RoleDataScope
	if err := r.provider.Read().WithContext(ctx).Where("role_id = ?", roleID).Order("id ASC").Find(&joins).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(joins))
	for _, item := range joins {
		ids = append(ids, item.DataScopeID)
	}
	return r.dataScopesByIDs(ctx, ids)
}

func (r *GORMRepository) RoleScopeIDs(ctx context.Context, roleID int) ([]int, error) {
	if _, err := r.GetRole(ctx, roleID); err != nil {
		return nil, err
	}
	var joins []RoleDataScope
	if err := r.provider.Read().WithContext(ctx).Where("role_id = ?", roleID).Order("id ASC").Find(&joins).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(joins))
	for _, item := range joins {
		ids = append(ids, item.DataScopeID)
	}
	return ids, nil
}

func (r *GORMRepository) UpdateRoleScopes(ctx context.Context, roleID int, scopeIDs []int) error {
	if _, err := r.GetRole(ctx, roleID); err != nil {
		return err
	}
	knownIDs, err := r.knownDataScopeIDs(ctx, scopeIDs)
	if err != nil {
		return err
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceRoleDataScopes(tx, roleID, knownIDs)
	})
}

func (r *GORMRepository) GetMenu(ctx context.Context, id int) (model.Menu, error) {
	var item model.Menu
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetMenuByTitle(ctx context.Context, title string) (model.Menu, error) {
	var item model.Menu
	err := r.provider.Read().WithContext(ctx).Where("title = ? AND type <> ?", title, 2).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListMenus(ctx context.Context, filter MenuFilter) ([]model.Menu, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Menu{})
	if filter.Title != "" {
		query = query.Where("title LIKE ?", "%"+filter.Title+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var items []model.Menu
	err := query.Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) SidebarMenus(ctx context.Context) ([]model.Menu, error) {
	var items []model.Menu
	err := r.provider.Read().WithContext(ctx).Where("type <> ?", 2).Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) CreateMenu(ctx context.Context, param dto.MenuParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.Menu{
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
		CreatedTime: time.Now(),
	}).Error
}

func (r *GORMRepository) UpdateMenu(ctx context.Context, id int, param dto.MenuParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Menu{}).Where("id = ?", id).Updates(map[string]any{
		"title":     param.Title,
		"name":      param.Name,
		"path":      param.Path,
		"parent_id": param.ParentID,
		"sort":      param.Sort,
		"icon":      param.Icon,
		"type":      param.Type,
		"component": param.Component,
		"perms":     param.Perms,
		"status":    param.Status,
		"display":   param.Display,
		"cache":     param.Cache,
		"link":      param.Link,
		"remark":    param.Remark,
	})
	return rowsError(result)
}

func (r *GORMRepository) DeleteMenu(ctx context.Context, id int) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", id).Delete(&RoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Menu{}, id).Error
	})
}

func (r *GORMRepository) MenuHasChildren(ctx context.Context, id int) (bool, error) {
	var count int64
	if err := r.provider.Read().WithContext(ctx).Model(&model.Menu{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GORMRepository) GetDept(ctx context.Context, id int) (model.Dept, error) {
	var item model.Dept
	err := r.provider.Read().WithContext(ctx).Where("id = ? AND deleted = ?", id, 0).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetDeptByName(ctx context.Context, name string) (model.Dept, error) {
	var item model.Dept
	err := r.provider.Read().WithContext(ctx).Where("name = ? AND deleted = ?", name, 0).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListDepts(ctx context.Context, filter DeptFilter) ([]model.Dept, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Dept{}).Where("deleted = ?", 0)
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Leader != "" {
		query = query.Where("leader LIKE ?", "%"+filter.Leader+"%")
	}
	if filter.Phone != "" {
		query = query.Where("phone LIKE ?", filter.Phone+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var items []model.Dept
	err := query.Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) CreateDept(ctx context.Context, param dto.DeptParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.Dept{
		Name:        param.Name,
		ParentID:    param.ParentID,
		Sort:        param.Sort,
		Leader:      param.Leader,
		Phone:       param.Phone,
		Email:       param.Email,
		Status:      param.Status,
		Deleted:     0,
		CreatedTime: time.Now(),
	}).Error
}

func (r *GORMRepository) UpdateDept(ctx context.Context, id int, param dto.DeptParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Dept{}).Where("id = ? AND deleted = ?", id, 0).Updates(map[string]any{
		"name":      param.Name,
		"parent_id": param.ParentID,
		"sort":      param.Sort,
		"leader":    param.Leader,
		"phone":     param.Phone,
		"email":     param.Email,
		"status":    param.Status,
	})
	return rowsError(result)
}

func (r *GORMRepository) DeleteDept(ctx context.Context, id int) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.Dept{}).Where("id = ? AND deleted = ?", id, 0).Updates(map[string]any{
		"deleted":      id,
		"deleted_time": time.Now(),
	})
	return rowsError(result)
}

func (r *GORMRepository) DeptHasChildren(ctx context.Context, id int) (bool, error) {
	var count int64
	if err := r.provider.Read().WithContext(ctx).Model(&model.Dept{}).Where("parent_id = ? AND deleted = ?", id, 0).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GORMRepository) DeptHasUsers(ctx context.Context, id int) (bool, error) {
	var count int64
	if err := r.provider.Read().WithContext(ctx).Model(&model.User{}).Where("dept_id = ? AND deleted = ?", id, 0).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GORMRepository) AllDataRules(ctx context.Context) ([]model.DataRule, error) {
	var items []model.DataRule
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) DataRuleModels(ctx context.Context) ([]string, error) {
	return r.fallback.DataRuleModels(ctx)
}

func (r *GORMRepository) DataRuleModelColumns(ctx context.Context, modelName string) ([]model.DataRuleColumn, error) {
	return r.fallback.DataRuleModelColumns(ctx, modelName)
}

func (r *GORMRepository) DataRuleValueTemplateVariables(ctx context.Context) ([]model.DataRuleTemplateVariable, error) {
	return r.fallback.DataRuleValueTemplateVariables(ctx)
}

func (r *GORMRepository) GetDataRule(ctx context.Context, id int) (model.DataRule, error) {
	var item model.DataRule
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetDataRuleByName(ctx context.Context, name string) (model.DataRule, error) {
	var item model.DataRule
	err := r.provider.Read().WithContext(ctx).Where("name = ?", name).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) ListDataRules(ctx context.Context, filter DataRuleFilter, page int, size int) ([]model.DataRule, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.DataRule{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	return paginateGORM[model.DataRule](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) CreateDataRule(ctx context.Context, param dto.DataRuleParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.DataRule{
		Name:        param.Name,
		Model:       param.Model,
		Column:      param.Column,
		Operator:    param.Operator,
		Expression:  param.Expression,
		Value:       param.Value,
		CreatedTime: time.Now(),
	}).Error
}

func (r *GORMRepository) UpdateDataRule(ctx context.Context, id int, param dto.DataRuleParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.DataRule{}).Where("id = ?", id).Updates(map[string]any{
		"name":       param.Name,
		"model":      param.Model,
		"column":     param.Column,
		"operator":   param.Operator,
		"expression": param.Expression,
		"value":      param.Value,
	})
	return rowsError(result)
}

func (r *GORMRepository) DeleteDataRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("data_rule_id IN ?", ids).Delete(&DataScopeRule{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DataRule{}, ids).Error
	})
}

func (r *GORMRepository) AllDataScopes(ctx context.Context) ([]model.DataScope, error) {
	var items []model.DataScope
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetDataScope(ctx context.Context, id int) (model.DataScope, error) {
	var item model.DataScope
	err := r.provider.Read().WithContext(ctx).First(&item, id).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) GetDataScopeByName(ctx context.Context, name string) (model.DataScope, error) {
	var item model.DataScope
	err := r.provider.Read().WithContext(ctx).Where("name = ?", name).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) DataScopeRules(ctx context.Context, id int) (model.DataScope, []model.DataRule, error) {
	scope, err := r.GetDataScope(ctx, id)
	if err != nil {
		return model.DataScope{}, nil, err
	}
	var joins []DataScopeRule
	if err := r.provider.Read().WithContext(ctx).Where("data_scope_id = ?", id).Order("id ASC").Find(&joins).Error; err != nil {
		return model.DataScope{}, nil, err
	}
	ids := make([]int, 0, len(joins))
	for _, item := range joins {
		ids = append(ids, item.DataRuleID)
	}
	rules, err := r.dataRulesByIDs(ctx, ids)
	return scope, rules, err
}

func (r *GORMRepository) ListDataScopes(ctx context.Context, filter DataScopeFilter, page int, size int) ([]model.DataScope, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.DataScope{})
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return paginateGORM[model.DataScope](query.Order("id ASC"), page, size)
}

func (r *GORMRepository) CreateDataScope(ctx context.Context, param dto.DataScopeParam) error {
	return r.provider.Write().WithContext(ctx).Create(&model.DataScope{
		Name:        param.Name,
		Status:      param.Status,
		CreatedTime: time.Now(),
	}).Error
}

func (r *GORMRepository) UpdateDataScope(ctx context.Context, id int, param dto.DataScopeParam) error {
	result := r.provider.Write().WithContext(ctx).Model(&model.DataScope{}).Where("id = ?", id).Updates(map[string]any{
		"name":   param.Name,
		"status": param.Status,
	})
	return rowsError(result)
}

func (r *GORMRepository) UpdateDataScopeRules(ctx context.Context, id int, ruleIDs []int) error {
	if _, err := r.GetDataScope(ctx, id); err != nil {
		return err
	}
	if err := r.ensureDataRules(ctx, ruleIDs); err != nil {
		return err
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceDataScopeRules(tx, id, ruleIDs)
	})
}

func (r *GORMRepository) DeleteDataScopes(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("data_scope_id IN ?", ids).Delete(&DataScopeRule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("data_scope_id IN ?", ids).Delete(&RoleDataScope{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DataScope{}, ids).Error
	})
}

func (r *GORMRepository) AllPlugins(ctx context.Context) ([]model.Plugin, error) {
	var items []model.Plugin
	err := r.provider.Read().WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetPlugin(ctx context.Context, id string) (model.Plugin, error) {
	var item model.Plugin
	err := r.provider.Read().WithContext(ctx).Where("id = ?", id).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) InstallPlugin(ctx context.Context, param dto.PluginInstallParam) (model.Plugin, error) {
	name := param.Name
	if name == "" {
		name = "plugin"
	}
	var item model.Plugin
	err := r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("id = ?", name).First(&item).Error
		if err == nil {
			if err := tx.Model(&model.Plugin{}).Where("id = ?", name).Update("enabled", true).Error; err != nil {
				return err
			}
			item.Enabled = true
			return setPluginChanged(tx, true)
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		item = model.Plugin{
			ID:          name,
			Summary:     name,
			Version:     "0.0.1",
			Description: "Installed plugin from " + param.Type,
			Author:      "external",
			Tags:        []string{"other"},
			Database:    []string{"mysql", "postgresql"},
			DependsOn:   []string{"admin"},
			Enabled:     true,
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return setPluginChanged(tx, true)
	})
	return item, err
}

func (r *GORMRepository) UninstallPlugin(ctx context.Context, id string) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.Plugin
		if err := tx.Where("id = ?", id).First(&item).Error; err != nil {
			return mapGORMError(err)
		}
		if item.BuiltIn {
			return setPluginChanged(tx, true)
		}
		if err := tx.Delete(&model.Plugin{}, "id = ?", id).Error; err != nil {
			return err
		}
		return setPluginChanged(tx, true)
	})
}

func (r *GORMRepository) TogglePluginStatus(ctx context.Context, id string) error {
	return r.provider.Write().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.Plugin
		if err := tx.Where("id = ?", id).First(&item).Error; err != nil {
			return mapGORMError(err)
		}
		if err := tx.Model(&model.Plugin{}).Where("id = ?", id).Update("enabled", !item.Enabled).Error; err != nil {
			return err
		}
		return setPluginChanged(tx, true)
	})
}

func (r *GORMRepository) PluginsChanged(ctx context.Context) (bool, error) {
	var state PluginState
	err := r.provider.Read().WithContext(ctx).Where("key = ?", pluginChangedStateKey).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.Value, nil
}

func (r *GORMRepository) ListLoginLogs(ctx context.Context, filter LogFilter, page int, size int) ([]model.LoginLog, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.LoginLog{})
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.IP != "" {
		query = query.Where("ip LIKE ?", "%"+filter.IP+"%")
	}
	return paginateGORM[model.LoginLog](query.Order("created_time DESC, id DESC"), page, size)
}

func (r *GORMRepository) DeleteLoginLogs(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.provider.Write().WithContext(ctx).Delete(&model.LoginLog{}, ids)
	return int(result.RowsAffected), result.Error
}

func (r *GORMRepository) DeleteAllLoginLogs(ctx context.Context) error {
	return r.provider.Write().WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.LoginLog{}).Error
}

func (r *GORMRepository) ListOperaLogs(ctx context.Context, filter LogFilter, page int, size int) ([]model.OperaLog, int64, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.OperaLog{})
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.IP != "" {
		query = query.Where("ip LIKE ?", "%"+filter.IP+"%")
	}
	return paginateGORM[model.OperaLog](query.Order("created_time DESC, id DESC"), page, size)
}

func (r *GORMRepository) DeleteOperaLogs(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.provider.Write().WithContext(ctx).Delete(&model.OperaLog{}, ids)
	return int(result.RowsAffected), result.Error
}

func (r *GORMRepository) DeleteAllOperaLogs(ctx context.Context) error {
	return r.provider.Write().WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.OperaLog{}).Error
}

func (r *GORMRepository) ListSessions(ctx context.Context, filter SessionFilter) ([]model.Session, error) {
	query := r.provider.Read().WithContext(ctx).Model(&model.Session{})
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	var items []model.Session
	err := query.Order("expire_time DESC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) GetSession(ctx context.Context, userID int, sessionUUID string) (model.Session, error) {
	var item model.Session
	err := r.provider.Read().WithContext(ctx).Where("id = ? AND session_uuid = ?", userID, sessionUUID).First(&item).Error
	return item, mapGORMError(err)
}

func (r *GORMRepository) UpsertSession(ctx context.Context, session model.Session) error {
	return r.provider.Write().WithContext(ctx).Save(&session).Error
}

func (r *GORMRepository) DeleteSession(ctx context.Context, userID int, sessionUUID string) error {
	return r.provider.Write().WithContext(ctx).Where("id = ? AND session_uuid = ?", userID, sessionUUID).Delete(&model.Session{}).Error
}

func seedIfEmpty[T any](tx *gorm.DB, table any, items []T) error {
	if len(items) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(table).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func seedUserRoles(tx *gorm.DB, source map[int][]int) error {
	for userID, roleIDs := range source {
		for _, roleID := range roleIDs {
			if err := tx.Where(UserRole{UserID: userID, RoleID: roleID}).FirstOrCreate(&UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedRoleMenus(tx *gorm.DB, source map[int][]int) error {
	for roleID, menuIDs := range source {
		for _, menuID := range menuIDs {
			if err := tx.Where(RoleMenu{RoleID: roleID, MenuID: menuID}).FirstOrCreate(&RoleMenu{RoleID: roleID, MenuID: menuID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedRoleDataScopes(tx *gorm.DB, source map[int][]int) error {
	for roleID, scopeIDs := range source {
		for _, scopeID := range scopeIDs {
			if err := tx.Where(RoleDataScope{RoleID: roleID, DataScopeID: scopeID}).FirstOrCreate(&RoleDataScope{RoleID: roleID, DataScopeID: scopeID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedDataScopeRules(tx *gorm.DB, source map[int][]int) error {
	for scopeID, ruleIDs := range source {
		for _, ruleID := range ruleIDs {
			if err := tx.Where(DataScopeRule{DataScopeID: scopeID, DataRuleID: ruleID}).FirstOrCreate(&DataScopeRule{DataScopeID: scopeID, DataRuleID: ruleID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func setPluginChanged(tx *gorm.DB, changed bool) error {
	state := PluginState{Key: pluginChangedStateKey, Value: changed}
	return tx.Where("key = ?", pluginChangedStateKey).Assign(state).FirstOrCreate(&state).Error
}

func updateUserColumns(query *gorm.DB, id int, updates map[string]any) error {
	result := query.Model(&model.User{}).Where("id = ? AND deleted = ?", id, 0).Updates(updates)
	return rowsError(result)
}

func replaceUserRoles(tx *gorm.DB, userID int, roleIDs []int) error {
	if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err := tx.Create(&UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceRoleMenus(tx *gorm.DB, roleID int, menuIDs []int) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&RoleMenu{}).Error; err != nil {
		return err
	}
	for _, menuID := range menuIDs {
		if err := tx.Create(&RoleMenu{RoleID: roleID, MenuID: menuID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceRoleDataScopes(tx *gorm.DB, roleID int, scopeIDs []int) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&RoleDataScope{}).Error; err != nil {
		return err
	}
	for _, scopeID := range scopeIDs {
		if err := tx.Create(&RoleDataScope{RoleID: roleID, DataScopeID: scopeID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceDataScopeRules(tx *gorm.DB, scopeID int, ruleIDs []int) error {
	if err := tx.Where("data_scope_id = ?", scopeID).Delete(&DataScopeRule{}).Error; err != nil {
		return err
	}
	for _, ruleID := range ruleIDs {
		if err := tx.Create(&DataScopeRule{DataScopeID: scopeID, DataRuleID: ruleID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GORMRepository) ensureRoles(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := r.provider.Read().WithContext(ctx).Model(&model.Role{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(uniqueInts(ids))) {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) ensureDataRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := r.provider.Read().WithContext(ctx).Model(&model.DataRule{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(uniqueInts(ids))) {
		return ErrNotFound
	}
	return nil
}

func (r *GORMRepository) knownMenuIDs(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.Menu
	if err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	known := make(map[int]struct{}, len(items))
	for _, item := range items {
		known[item.ID] = struct{}{}
	}
	return filterKnownIDs(ids, func(id int) bool {
		_, ok := known[id]
		return ok
	}), nil
}

func (r *GORMRepository) knownDataScopeIDs(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.DataScope
	if err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	known := make(map[int]struct{}, len(items))
	for _, item := range items {
		known[item.ID] = struct{}{}
	}
	return filterKnownIDs(ids, func(id int) bool {
		_, ok := known[id]
		return ok
	}), nil
}

func (r *GORMRepository) rolesByIDs(ctx context.Context, ids []int) ([]model.Role, error) {
	if len(ids) == 0 {
		return []model.Role{}, nil
	}
	var items []model.Role
	err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) menusByIDs(ctx context.Context, ids []int) ([]model.Menu, error) {
	if len(ids) == 0 {
		return []model.Menu{}, nil
	}
	var items []model.Menu
	err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) dataScopesByIDs(ctx context.Context, ids []int) ([]model.DataScope, error) {
	if len(ids) == 0 {
		return []model.DataScope{}, nil
	}
	var items []model.DataScope
	err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *GORMRepository) dataRulesByIDs(ctx context.Context, ids []int) ([]model.DataRule, error) {
	if len(ids) == 0 {
		return []model.DataRule{}, nil
	}
	var items []model.DataRule
	err := r.provider.Read().WithContext(ctx).Where("id IN ?", ids).Order("id ASC").Find(&items).Error
	return items, err
}

func paginateGORM[T any](query *gorm.DB, page int, size int) ([]T, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []T
	err := query.Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func rowsError(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func mapGORMError(err error) error {
	if err == nil {
		return nil
	}
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func uniqueInts(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
