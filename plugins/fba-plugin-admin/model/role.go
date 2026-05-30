package model

import "time"

type Role struct {
	ID             int
	Name           string
	Status         int
	IsFilterScopes bool
	Remark         *string
	CreatedTime    time.Time
	UpdatedTime    *time.Time
}

type Menu struct {
	ID          int
	Title       string
	Name        string
	Path        *string
	ParentID    *int
	Sort        int
	Icon        *string
	Type        int
	Component   *string
	Perms       *string
	Status      int
	Display     int
	Cache       int
	Link        *string
	Remark      *string
	CreatedTime time.Time
	UpdatedTime *time.Time
}

type DataScope struct {
	ID          int
	Name        string
	Status      int
	CreatedTime time.Time
	UpdatedTime *time.Time
}

type Seed struct {
	Roles      []Role
	Menus      []Menu
	DataScopes []DataScope
	RoleMenus  map[int][]int
	RoleScopes map[int][]int
}

func SeedData() Seed {
	dashboardPath := "/dashboard"
	dashboardIcon := "lucide:layout-dashboard"
	dashboardComponent := "Layout"
	created := seedTime()
	return Seed{
		Roles: []Role{
			{
				ID:             1,
				Name:           "admin",
				Status:         1,
				IsFilterScopes: true,
				CreatedTime:    created,
			},
		},
		Menus: []Menu{
			{
				ID:          1,
				Title:       "仪表盘",
				Name:        "Dashboard",
				Path:        &dashboardPath,
				Sort:        0,
				Icon:        &dashboardIcon,
				Type:        1,
				Component:   &dashboardComponent,
				Status:      1,
				Display:     1,
				Cache:       1,
				CreatedTime: created,
			},
		},
		DataScopes: []DataScope{
			{
				ID:          1,
				Name:        "本人数据范围",
				Status:      1,
				CreatedTime: created,
			},
		},
		RoleMenus:  map[int][]int{1: {1}},
		RoleScopes: map[int][]int{},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
