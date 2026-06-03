package model

import "time"

type UserPasswordHistory struct {
	ID          int       `gorm:"column:id;primaryKey"`
	UserID      int       `gorm:"column:user_id;index"`
	Password    string    `gorm:"column:password;size:256"`
	CreatedTime time.Time `gorm:"column:created_time;autoCreateTime;index"`
}

func (UserPasswordHistory) TableName() string {
	return "sys_user_password_history"
}
