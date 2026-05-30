package model

import "time"

type DictData struct {
	ID          int        `gorm:"column:id;primaryKey"`
	TypeID      int        `gorm:"column:type_id;index"`
	TypeCode    string     `gorm:"column:type_code;size:32;index"`
	Label       string     `gorm:"column:label;size:32"`
	Value       string     `gorm:"column:value;size:32"`
	Color       *string    `gorm:"column:color;size:32"`
	Sort        int        `gorm:"column:sort"`
	Status      int        `gorm:"column:status"`
	Remark      *string    `gorm:"column:remark"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (DictData) TableName() string {
	return "sys_dict_data"
}

func SeedDictData() []DictData {
	disabledColor := "red"
	disabledRemark := "停用状态"
	enabledColor := "green"
	enabledRemark := "正常状态"
	closedColor := "error"
	closedRemark := "关闭状态"
	openColor := "success"
	openRemark := "开启状态"
	return []DictData{
		{
			ID:          1,
			TypeID:      1,
			TypeCode:    "sys_status",
			Label:       "停用",
			Value:       "0",
			Color:       &disabledColor,
			Sort:        1,
			Status:      1,
			Remark:      &disabledRemark,
			CreatedTime: seedTime(),
		},
		{
			ID:          2,
			TypeID:      1,
			TypeCode:    "sys_status",
			Label:       "正常",
			Value:       "1",
			Color:       &enabledColor,
			Sort:        2,
			Status:      1,
			Remark:      &enabledRemark,
			CreatedTime: seedTime(),
		},
		{
			ID:          3,
			TypeID:      2,
			TypeCode:    "sys_choose",
			Label:       "关闭",
			Value:       "false",
			Color:       &closedColor,
			Sort:        1,
			Status:      1,
			Remark:      &closedRemark,
			CreatedTime: seedTime(),
		},
		{
			ID:          4,
			TypeID:      2,
			TypeCode:    "sys_choose",
			Label:       "开启",
			Value:       "true",
			Color:       &openColor,
			Sort:        2,
			Status:      1,
			Remark:      &openRemark,
			CreatedTime: seedTime(),
		},
	}
}
