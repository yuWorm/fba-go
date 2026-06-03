package model

import "time"

type UserSocial struct {
	ID          int        `gorm:"column:id;primaryKey"`
	SID         string     `gorm:"column:sid;size:256;uniqueIndex:uk_sys_user_social_sid_source_deleted"`
	Source      string     `gorm:"column:source;size:32;uniqueIndex:uk_sys_user_social_user_id_source_deleted;uniqueIndex:uk_sys_user_social_sid_source_deleted"`
	UserID      int        `gorm:"column:user_id;uniqueIndex:uk_sys_user_social_user_id_source_deleted;index"`
	Deleted     int        `gorm:"column:deleted;uniqueIndex:uk_sys_user_social_user_id_source_deleted;uniqueIndex:uk_sys_user_social_sid_source_deleted;index"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime time.Time  `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime *time.Time `gorm:"column:deleted_time;index"`
}

func (UserSocial) TableName() string {
	return "sys_user_social"
}
