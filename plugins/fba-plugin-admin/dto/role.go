package dto

import (
	"time"

	"github.com/yuWorm/fba-plugin-admin/model"
)

const TimeLayout = "2006-01-02 15:04:05"

type RoleParam struct {
	Name           string  `json:"name"`
	Status         int     `json:"status"`
	IsFilterScopes bool    `json:"is_filter_scopes"`
	Remark         *string `json:"remark"`
}

type RoleMenuParam struct {
	Menus []int `json:"menus"`
}

type RoleScopeParam struct {
	Scopes []int `json:"scopes"`
}

type DeleteParam struct {
	PKs []int `json:"pks"`
}

type MenuParam struct {
	Title     string  `json:"title"`
	Name      string  `json:"name"`
	Path      *string `json:"path"`
	ParentID  *int    `json:"parent_id"`
	Sort      int     `json:"sort"`
	Icon      *string `json:"icon"`
	Type      int     `json:"type"`
	Component *string `json:"component"`
	Perms     *string `json:"perms"`
	Status    int     `json:"status"`
	Display   int     `json:"display"`
	Cache     int     `json:"cache"`
	Link      *string `json:"link"`
	Remark    *string `json:"remark"`
}

type DeptParam struct {
	Name     string  `json:"name"`
	ParentID *int    `json:"parent_id"`
	Sort     int     `json:"sort"`
	Leader   *string `json:"leader"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	Status   int     `json:"status"`
}

type DataRuleParam struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Column     string `json:"column"`
	Operator   int    `json:"operator"`
	Expression int    `json:"expression"`
	Value      string `json:"value"`
}

type DataRuleColumnDetail struct {
	Key     string `json:"key"`
	Comment string `json:"comment"`
}

type DataRuleTemplateVariableDetail struct {
	Key     string `json:"key"`
	Comment string `json:"comment"`
}

type DataScopeParam struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type DataScopeRuleParam struct {
	Rules []int `json:"rules"`
}

type RoleDetail struct {
	Name           string  `json:"name"`
	Status         int     `json:"status"`
	IsFilterScopes bool    `json:"is_filter_scopes"`
	Remark         *string `json:"remark"`
	ID             int     `json:"id"`
	CreatedTime    string  `json:"created_time"`
	UpdatedTime    *string `json:"updated_time"`
}

type RoleWithRelationDetail struct {
	RoleDetail
	Menus  []MenuDetail      `json:"menus"`
	Scopes []DataScopeDetail `json:"scopes"`
}

type MenuDetail struct {
	Title       string       `json:"title"`
	Name        string       `json:"name"`
	Path        *string      `json:"path"`
	ParentID    *int         `json:"parent_id"`
	Sort        int          `json:"sort"`
	Icon        *string      `json:"icon"`
	Type        int          `json:"type"`
	Component   *string      `json:"component"`
	Perms       *string      `json:"perms"`
	Status      int          `json:"status"`
	Display     int          `json:"display"`
	Cache       int          `json:"cache"`
	Link        *string      `json:"link"`
	Remark      *string      `json:"remark"`
	ID          int          `json:"id"`
	CreatedTime string       `json:"created_time"`
	UpdatedTime *string      `json:"updated_time"`
	Children    []MenuDetail `json:"children,omitempty"`
}

type SidebarMenu struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Path      *string       `json:"path"`
	ParentID  *int          `json:"parent_id"`
	Sort      int           `json:"sort"`
	Type      int           `json:"type"`
	Component *string       `json:"component"`
	Perms     *string       `json:"perms"`
	Remark    *string       `json:"remark"`
	Children  []SidebarMenu `json:"children"`
	Meta      SidebarMeta   `json:"meta"`
}

type SidebarMeta struct {
	Title                    string `json:"title"`
	Icon                     string `json:"icon"`
	IframeSrc                string `json:"iframeSrc"`
	Link                     string `json:"link"`
	KeepAlive                bool   `json:"keepAlive"`
	HideInMenu               bool   `json:"hideInMenu"`
	MenuVisibleWithForbidden bool   `json:"menuVisibleWithForbidden"`
}

type DeptDetail struct {
	Name        string       `json:"name"`
	ParentID    *int         `json:"parent_id"`
	Sort        int          `json:"sort"`
	Leader      *string      `json:"leader"`
	Phone       *string      `json:"phone"`
	Email       *string      `json:"email"`
	Status      int          `json:"status"`
	ID          int          `json:"id"`
	Deleted     int          `json:"deleted"`
	CreatedTime string       `json:"created_time"`
	UpdatedTime *string      `json:"updated_time"`
	DeletedTime *string      `json:"deleted_time"`
	Children    []DeptDetail `json:"children,omitempty"`
}

type DataRuleDetail struct {
	Name        string  `json:"name"`
	Model       string  `json:"model"`
	Column      string  `json:"column"`
	Operator    int     `json:"operator"`
	Expression  int     `json:"expression"`
	Value       string  `json:"value"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

type DataScopeDetail struct {
	Name        string  `json:"name"`
	Status      int     `json:"status"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

type DataScopeWithRelationDetail struct {
	DataScopeDetail
	Rules []DataRuleDetail `json:"rules"`
}

func RoleFromModel(item model.Role) RoleDetail {
	return RoleDetail{
		ID:             item.ID,
		Name:           item.Name,
		Status:         item.Status,
		IsFilterScopes: item.IsFilterScopes,
		Remark:         item.Remark,
		CreatedTime:    formatTime(item.CreatedTime),
		UpdatedTime:    formatTimePtr(item.UpdatedTime),
	}
}

func RolesFromModel(items []model.Role) []RoleDetail {
	result := make([]RoleDetail, 0, len(items))
	for _, item := range items {
		result = append(result, RoleFromModel(item))
	}
	return result
}

func RoleWithRelations(item model.Role, menus []model.Menu, scopes []model.DataScope) RoleWithRelationDetail {
	return RoleWithRelationDetail{
		RoleDetail: RoleFromModel(item),
		Menus:      MenusFromModel(menus),
		Scopes:     DataScopesFromModel(scopes),
	}
}

func MenuFromModel(item model.Menu) MenuDetail {
	return MenuDetail{
		ID:          item.ID,
		Title:       item.Title,
		Name:        item.Name,
		Path:        item.Path,
		ParentID:    item.ParentID,
		Sort:        item.Sort,
		Icon:        item.Icon,
		Type:        item.Type,
		Component:   item.Component,
		Perms:       item.Perms,
		Status:      item.Status,
		Display:     item.Display,
		Cache:       item.Cache,
		Link:        item.Link,
		Remark:      item.Remark,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func MenusFromModel(items []model.Menu) []MenuDetail {
	result := make([]MenuDetail, 0, len(items))
	for _, item := range items {
		result = append(result, MenuFromModel(item))
	}
	return result
}

func SidebarMenuFromModel(item model.Menu) SidebarMenu {
	icon := ""
	if item.Icon != nil {
		icon = *item.Icon
	}
	link := ""
	if item.Link != nil {
		link = *item.Link
	}
	return SidebarMenu{
		ID:        item.ID,
		Name:      item.Name,
		Path:      item.Path,
		ParentID:  item.ParentID,
		Sort:      item.Sort,
		Type:      item.Type,
		Component: item.Component,
		Perms:     item.Perms,
		Remark:    item.Remark,
		Children:  []SidebarMenu{},
		Meta: SidebarMeta{
			Title:                    item.Title,
			Icon:                     icon,
			IframeSrc:                "",
			Link:                     link,
			KeepAlive:                item.Cache == 1,
			HideInMenu:               item.Display == 0,
			MenuVisibleWithForbidden: false,
		},
	}
}

func DeptFromModel(item model.Dept) DeptDetail {
	return DeptDetail{
		ID:          item.ID,
		Name:        item.Name,
		ParentID:    item.ParentID,
		Sort:        item.Sort,
		Leader:      item.Leader,
		Phone:       item.Phone,
		Email:       item.Email,
		Status:      item.Status,
		Deleted:     item.Deleted,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
		DeletedTime: formatTimePtr(item.DeletedTime),
	}
}

func DataRuleFromModel(item model.DataRule) DataRuleDetail {
	return DataRuleDetail{
		ID:          item.ID,
		Name:        item.Name,
		Model:       item.Model,
		Column:      item.Column,
		Operator:    item.Operator,
		Expression:  item.Expression,
		Value:       item.Value,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func DataRuleColumnFromModel(item model.DataRuleColumn) DataRuleColumnDetail {
	return DataRuleColumnDetail{
		Key:     item.Key,
		Comment: item.Comment,
	}
}

func DataRuleColumnsFromModel(items []model.DataRuleColumn) []DataRuleColumnDetail {
	result := make([]DataRuleColumnDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DataRuleColumnFromModel(item))
	}
	return result
}

func DataRuleTemplateVariableFromModel(item model.DataRuleTemplateVariable) DataRuleTemplateVariableDetail {
	return DataRuleTemplateVariableDetail{
		Key:     item.Key,
		Comment: item.Comment,
	}
}

func DataRuleTemplateVariablesFromModel(items []model.DataRuleTemplateVariable) []DataRuleTemplateVariableDetail {
	result := make([]DataRuleTemplateVariableDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DataRuleTemplateVariableFromModel(item))
	}
	return result
}

func DataRulesFromModel(items []model.DataRule) []DataRuleDetail {
	result := make([]DataRuleDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DataRuleFromModel(item))
	}
	return result
}

func DataScopeFromModel(item model.DataScope) DataScopeDetail {
	return DataScopeDetail{
		ID:          item.ID,
		Name:        item.Name,
		Status:      item.Status,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func DataScopesFromModel(items []model.DataScope) []DataScopeDetail {
	result := make([]DataScopeDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DataScopeFromModel(item))
	}
	return result
}

func DataScopeWithRules(item model.DataScope, rules []model.DataRule) DataScopeWithRelationDetail {
	return DataScopeWithRelationDetail{
		DataScopeDetail: DataScopeFromModel(item),
		Rules:           DataRulesFromModel(rules),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(TimeLayout)
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatTime(*value)
	return &formatted
}
