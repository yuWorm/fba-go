package model

import (
	"strings"
	"time"
)

type Role struct {
	ID             int        `gorm:"column:id;primaryKey"`
	Name           string     `gorm:"column:name;size:32;index"`
	Status         int        `gorm:"column:status"`
	IsFilterScopes bool       `gorm:"column:is_filter_scopes"`
	Remark         *string    `gorm:"column:remark;type:text"`
	CreatedTime    time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime    *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

type Menu struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Title       string     `gorm:"column:title;size:64;index"`
	Name        string     `gorm:"column:name;size:64"`
	Path        *string    `gorm:"column:path;size:256"`
	ParentID    *int       `gorm:"column:parent_id;index"`
	Sort        int        `gorm:"column:sort"`
	Icon        *string    `gorm:"column:icon;size:128"`
	Type        int        `gorm:"column:type"`
	Component   *string    `gorm:"column:component;size:256"`
	Perms       *string    `gorm:"column:perms;type:text"`
	Status      int        `gorm:"column:status"`
	Display     int        `gorm:"column:display"`
	Cache       int        `gorm:"column:cache"`
	Link        *string    `gorm:"column:link;size:256"`
	Remark      *string    `gorm:"column:remark;type:text"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

type Dept struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:64;index"`
	ParentID    *int       `gorm:"column:parent_id;index"`
	Sort        int        `gorm:"column:sort"`
	Leader      *string    `gorm:"column:leader;size:64"`
	Phone       *string    `gorm:"column:phone;size:32"`
	Email       *string    `gorm:"column:email;size:256"`
	Status      int        `gorm:"column:status"`
	Deleted     int        `gorm:"column:deleted;index"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime *time.Time `gorm:"column:deleted_time;index"`
}

type DataRule struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:64;index"`
	Model       string     `gorm:"column:model;size:64"`
	Column      string     `gorm:"column:column;size:64"`
	Operator    int        `gorm:"column:operator"`
	Expression  int        `gorm:"column:expression"`
	Value       string     `gorm:"column:value;size:256"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

type DataScope struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:64;index"`
	Status      int        `gorm:"column:status"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (Role) TableName() string {
	return "sys_role"
}

func (Menu) TableName() string {
	return "sys_menu"
}

func (Dept) TableName() string {
	return "sys_dept"
}

func (DataRule) TableName() string {
	return "sys_data_rule"
}

func (DataScope) TableName() string {
	return "sys_data_scope"
}

type Seed struct {
	Users                           []User
	Roles                           []Role
	Menus                           []Menu
	Depts                           []Dept
	DataRules                       []DataRule
	DataScopes                      []DataScope
	DataRuleModelTemplateVariables  []DataRuleTemplateVariable
	DataRuleColumnTemplateVariables []DataRuleColumn
	DataRuleValueTemplateVariables  []DataRuleTemplateVariable
	DataRuleModels                  []DataRuleModelMetadata
	Plugins                         []Plugin
	LoginLogs                       []LoginLog
	OperaLogs                       []OperaLog
	Sessions                        []Session
	UserRoles                       map[int][]int
	RoleMenus                       map[int][]int
	RoleScopes                      map[int][]int
	ScopeRules                      map[int][]int
}

func SeedData() Seed {
	dashboardPath := "/dashboard"
	dashboardIcon := "lucide:layout-dashboard"
	dashboardComponent := "Layout"
	created := seedTime()
	headquartersName := "总部"
	country := "中国"
	region := "上海"
	city := "上海"
	userAgent := "fba-go contract"
	browser := "Chrome"
	osName := "macOS"
	device := "Desktop"
	operaUsername := "admin"
	operaMsg := "请求成功"
	adminPerms := strings.Join([]string{
		"sys:user:del",
		"sys:role:add",
		"sys:role:edit",
		"sys:role:menu:edit",
		"sys:role:del",
		"sys:menu:add",
		"sys:menu:edit",
		"sys:menu:del",
		"data:rule:add",
		"data:rule:edit",
		"data:rule:del",
		"data:scope:add",
		"data:scope:edit",
		"data:scope:rule:edit",
		"data:scope:del",
		"sys:file:upload",
		"log:login:del",
		"log:login:clear",
		"log:opera:del",
		"log:opera:clear",
		"dict:type:add",
		"dict:type:edit",
		"dict:type:del",
		"dict:data:add",
		"dict:data:edit",
		"dict:data:del",
		"sys:notice:add",
		"sys:notice:edit",
		"sys:notice:del",
		"sys:task:revoke",
		"sys:task:del",
		"sys:task:add",
		"sys:task:edit",
		"sys:task:exec",
	}, ",")
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
			{
				ID:             2,
				Name:           "viewer",
				Status:         1,
				IsFilterScopes: false,
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
				Perms:       &adminPerms,
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
		DataRuleModelTemplateVariables: []DataRuleTemplateVariable{
			{Key: "__ALL__", Comment: "所有模型"},
		},
		DataRuleColumnTemplateVariables: []DataRuleColumn{
			{Key: "__dept_id__", Comment: "部门 ID"},
			{Key: "__created_by__", Comment: "创建者"},
		},
		DataRuleValueTemplateVariables: []DataRuleTemplateVariable{
			{Key: "${user_id}", Comment: "当前登录用户 ID"},
			{Key: "${dept_id}", Comment: "当前登录用户部门 ID"},
			{Key: "${now}", Comment: "当前时间"},
		},
		DataRuleModels: []DataRuleModelMetadata{
			{
				Name: "user",
				Columns: []DataRuleColumn{
					{Key: "uuid", Comment: "用户 UUID"},
					{Key: "dept_id", Comment: "部门 ID"},
					{Key: "username", Comment: "用户名"},
					{Key: "nickname", Comment: "昵称"},
					{Key: "status", Comment: "状态"},
					{Key: "is_superuser", Comment: "是否超级管理员"},
					{Key: "is_staff", Comment: "是否管理员"},
					{Key: "is_multi_login", Comment: "是否允许多端登录"},
					{Key: "join_time", Comment: "加入时间"},
					{Key: "last_login_time", Comment: "最后登录时间"},
				},
			},
			{
				Name: "role",
				Columns: []DataRuleColumn{
					{Key: "name", Comment: "角色名称"},
					{Key: "status", Comment: "状态"},
					{Key: "is_filter_scopes", Comment: "是否过滤数据权限"},
					{Key: "remark", Comment: "备注"},
				},
			},
			{
				Name: "menu",
				Columns: []DataRuleColumn{
					{Key: "title", Comment: "菜单标题"},
					{Key: "name", Comment: "路由名称"},
					{Key: "path", Comment: "路由路径"},
					{Key: "parent_id", Comment: "父级 ID"},
					{Key: "type", Comment: "类型"},
					{Key: "component", Comment: "组件"},
					{Key: "perms", Comment: "权限标识"},
					{Key: "status", Comment: "状态"},
					{Key: "display", Comment: "是否显示"},
					{Key: "cache", Comment: "是否缓存"},
				},
			},
			{
				Name: "dept",
				Columns: []DataRuleColumn{
					{Key: "name", Comment: "部门名称"},
					{Key: "parent_id", Comment: "父级 ID"},
					{Key: "leader", Comment: "负责人"},
					{Key: "phone", Comment: "联系电话"},
					{Key: "email", Comment: "邮箱"},
					{Key: "status", Comment: "状态"},
				},
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
		LoginLogs: []LoginLog{
			{
				ID:          1,
				UserUUID:    "fixture-user",
				Username:    "admin",
				Status:      1,
				IP:          "127.0.0.1",
				Country:     &country,
				Region:      &region,
				City:        &city,
				UserAgent:   &userAgent,
				Browser:     &browser,
				OS:          &osName,
				Device:      &device,
				Msg:         "登录成功",
				LoginTime:   created,
				CreatedTime: created,
			},
		},
		OperaLogs: []OperaLog{
			{
				ID:          1,
				TraceID:     "fixture-trace",
				Username:    &operaUsername,
				Method:      "GET",
				Title:       "List users",
				Path:        "/api/v1/sys/users",
				IP:          "127.0.0.1",
				Country:     &country,
				Region:      &region,
				City:        &city,
				UserAgent:   &userAgent,
				Browser:     &browser,
				OS:          &osName,
				Device:      &device,
				Args:        map[string]any{"page": "1", "size": "20"},
				Status:      1,
				Code:        "200",
				Msg:         &operaMsg,
				CostTime:    1.2,
				OperaTime:   created,
				CreatedTime: created,
			},
		},
		Sessions: []Session{
			{
				ID:            1,
				SessionUUID:   "fixture-session",
				Username:      "admin",
				Nickname:      "Admin",
				IP:            "127.0.0.1",
				OS:            "macOS",
				Browser:       "Chrome",
				Device:        "Desktop",
				Status:        1,
				LastLoginTime: created.Format("2006-01-02 15:04:05"),
				ExpireTime:    created.Add(2 * time.Hour),
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
