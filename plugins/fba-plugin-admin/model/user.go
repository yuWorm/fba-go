package model

import "time"

type User struct {
	ID            int        `gorm:"column:id;primaryKey"`
	UUID          string     `gorm:"column:uuid;size:64;uniqueIndex"`
	DeptID        *int       `gorm:"column:dept_id;index"`
	Username      string     `gorm:"column:username;size:64;index"`
	Nickname      string     `gorm:"column:nickname;size:64"`
	Password      string     `gorm:"column:password;size:256"`
	Avatar        *string    `gorm:"column:avatar;size:256"`
	Email         *string    `gorm:"column:email;size:256;index"`
	Phone         *string    `gorm:"column:phone;size:32"`
	Status        int        `gorm:"column:status;index"`
	IsSuperuser   bool       `gorm:"column:is_superuser"`
	IsStaff       bool       `gorm:"column:is_staff"`
	IsMultiLogin  bool       `gorm:"column:is_multi_login"`
	JoinTime      time.Time  `gorm:"column:join_time;autoCreateTime"`
	LastLoginTime *time.Time `gorm:"column:last_login_time"`
	DeletedTime   *time.Time `gorm:"column:deleted_time;index"`
}

func (User) TableName() string {
	return "sys_user"
}
