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

type Dept struct {
	ID          int
	Name        string
	ParentID    *int
	Sort        int
	Leader      *string
	Phone       *string
	Email       *string
	Status      int
	Deleted     int
	CreatedTime time.Time
	UpdatedTime *time.Time
	DeletedTime *time.Time
}

type DataRule struct {
	ID          int
	Name        string
	Model       string
	Column      string
	Operator    int
	Expression  int
	Value       string
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
	Users      []User
	Roles      []Role
	Menus      []Menu
	Depts      []Dept
	DataRules  []DataRule
	DataScopes []DataScope
	Plugins    []Plugin
	UserRoles  map[int][]int
	RoleMenus  map[int][]int
	RoleScopes map[int][]int
	ScopeRules map[int][]int
}

func SeedData() Seed {
	dashboardPath := "/dashboard"
	dashboardIcon := "lucide:layout-dashboard"
	dashboardComponent := "Layout"
	created := seedTime()
	headquartersName := "总部"
	return Seed{
		Users: []User{
			{
				ID:           1,
				UUID:         "fixture-user",
				Username:     "admin",
				Nickname:     "Admin",
				Status:       1,
				IsSuperuser:  true,
				IsStaff:      true,
				IsMultiLogin: true,
				JoinTime:     created,
			},
		},
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
		Depts: []Dept{
			{
				ID:          1,
				Name:        headquartersName,
				Sort:        0,
				Status:      1,
				Deleted:     0,
				CreatedTime: created,
			},
		},
		DataRules: []DataRule{
			{
				ID:          1,
				Name:        "本人数据",
				Model:       "user",
				Column:      "id",
				Operator:    0,
				Expression:  0,
				Value:       "{{ user_id }}",
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
		Plugins: []Plugin{
			{
				ID:          "dict",
				Summary:     "数据字典",
				Version:     "0.0.8",
				Description: "Dictionary data plugin",
				Author:      "wu-clan",
				Tags:        []string{"other"},
				Database:    []string{"mysql", "postgresql"},
				DependsOn:   []string{"admin"},
				Enabled:     true,
				BuiltIn:     true,
			},
			{
				ID:          "notice",
				Summary:     "通知公告",
				Version:     "0.0.2",
				Description: "System notice and announcement plugin",
				Author:      "wu-clan",
				Tags:        []string{"other"},
				Database:    []string{"mysql", "postgresql"},
				DependsOn:   []string{"admin"},
				Enabled:     true,
				BuiltIn:     true,
			},
			{
				ID:          "task",
				Summary:     "任务调度",
				Version:     "0.1.0",
				Description: "Task scheduler compatibility plugin",
				Author:      "wu-clan",
				Tags:        []string{"task"},
				Database:    []string{"mysql", "postgresql"},
				DependsOn:   []string{"admin"},
				Enabled:     true,
				BuiltIn:     true,
			},
		},
		UserRoles:  map[int][]int{1: {1}},
		RoleMenus:  map[int][]int{1: {1}},
		RoleScopes: map[int][]int{},
		ScopeRules: map[int][]int{1: {1}},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
