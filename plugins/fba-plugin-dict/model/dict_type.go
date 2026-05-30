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
	sysStatusRemark := "系统通用状态：1/0"
	sysChooseRemark := "系统通用开关：true/false"
	return []DictType{
		{
			ID:          1,
			Name:        "通用状态",
			Code:        "sys_status",
			Remark:      &sysStatusRemark,
			CreatedTime: seedTime(),
		},
		{
			ID:          2,
			Name:        "通用开关",
			Code:        "sys_choose",
			Remark:      &sysChooseRemark,
			CreatedTime: seedTime(),
		},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
