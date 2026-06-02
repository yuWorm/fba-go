package model

import "time"

type DictType struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:32"`
	Code        string     `gorm:"column:code;size:32;index"`
	Remark      *string    `gorm:"column:remark"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (DictType) TableName() string {
	return "sys_dict_type"
}

func SeedDictTypes() []DictType {
	created := seedTime()
	ptr := func(value string) *string {
		return &value
	}
	// IDs and codes mirror the Python plugin SQL fixtures so frontend option lookups stay compatible.
	return []DictType{
		{
			ID:          1,
			Name:        "通用状态",
			Code:        "sys_status",
			Remark:      ptr("系统通用状态：1/0"),
			CreatedTime: created,
		},
		{
			ID:          2,
			Name:        "通用开关",
			Code:        "sys_choose",
			Remark:      ptr("系统通用开关：true/false"),
			CreatedTime: created,
		},
		{
			ID:          3,
			Name:        "菜单类型",
			Code:        "sys_menu_type",
			Remark:      ptr("系统菜单类型"),
			CreatedTime: created,
		},
		{
			ID:          4,
			Name:        "登录状态",
			Code:        "sys_login_status",
			Remark:      ptr("用户登录状态"),
			CreatedTime: created,
		},
		{
			ID:          5,
			Name:        "数据规则运算符",
			Code:        "sys_data_rule_operator",
			Remark:      ptr("数据权限规则运算符"),
			CreatedTime: created,
		},
		{
			ID:          6,
			Name:        "数据规则表达式",
			Code:        "sys_data_rule_expression",
			Remark:      ptr("数据权限规则表达式"),
			CreatedTime: created,
		},
		{
			ID:          7,
			Name:        "前端参数配置",
			Code:        "sys_frontend_config",
			Remark:      ptr("前端参数配置类型"),
			CreatedTime: created,
		},
		{
			ID:          8,
			Name:        "任务策略类型",
			Code:        "task_strategy_type",
			Remark:      ptr("定时任务策略类型"),
			CreatedTime: created,
		},
		{
			ID:          9,
			Name:        "任务周期类型",
			Code:        "task_period_type",
			Remark:      ptr("定时任务周期类型"),
			CreatedTime: created,
		},
		{
			ID:          10,
			Name:        "通知公告",
			Code:        "notice",
			Remark:      ptr("通知类型"),
			CreatedTime: created,
		},
		{
			ID:          11,
			Name:        "在线状态",
			Code:        "user_online_status",
			Remark:      ptr("用户在线状态"),
			CreatedTime: created,
		},
		{
			ID:          12,
			Name:        "插件类型",
			Code:        "sys_plugin_type",
			Remark:      ptr("插件类型"),
			CreatedTime: created,
		},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
