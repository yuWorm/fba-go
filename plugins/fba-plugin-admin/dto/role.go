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

type DataScopeDetail struct {
	Name        string  `json:"name"`
	Status      int     `json:"status"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
	Rules       []any   `json:"rules,omitempty"`
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
