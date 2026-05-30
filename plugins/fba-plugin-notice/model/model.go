package model

import "time"

type Notice struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Title       string     `gorm:"column:title;size:64"`
	Type        int        `gorm:"column:type"`
	Status      int        `gorm:"column:status"`
	Content     string     `gorm:"column:content;type:text"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (Notice) TableName() string {
	return "sys_notice"
}

func SeedNotices() []Notice {
	return []Notice{
		{
			ID:          1,
			Title:       "System Notice",
			Type:        0,
			Status:      1,
			Content:     "Welcome to fba-go.",
			CreatedTime: seedTime(),
		},
	}
}

func seedTime() time.Time {
	return time.Date(2025, 12, 15, 15, 33, 16, 0, time.Local)
}
