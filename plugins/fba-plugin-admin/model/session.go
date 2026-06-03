package model

import "time"

type Session struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement:false"`
	SessionUUID   string    `gorm:"column:session_uuid;size:128;primaryKey"`
	AccessToken   string    `gorm:"column:access_token;size:2048"`
	Username      string    `gorm:"column:username;size:64;index"`
	Nickname      string    `gorm:"column:nickname;size:64"`
	IP            string    `gorm:"column:ip;size:64"`
	OS            string    `gorm:"column:os;size:64"`
	Browser       string    `gorm:"column:browser;size:64"`
	Device        string    `gorm:"column:device;size:64"`
	Status        int       `gorm:"column:status"`
	LastLoginTime string    `gorm:"column:last_login_time;size:32"`
	ExpireTime    time.Time `gorm:"column:expire_time;index"`
}

func (Session) TableName() string {
	return "sys_session"
}
