package model

import "time"

type LoginLog struct {
	ID          int       `gorm:"column:id;primaryKey"`
	UserUUID    string    `gorm:"column:user_uuid;size:64"`
	Username    string    `gorm:"column:username;size:64;index"`
	Status      int       `gorm:"column:status;index"`
	IP          string    `gorm:"column:ip;size:64;index"`
	Country     *string   `gorm:"column:country;size:64"`
	Region      *string   `gorm:"column:region;size:64"`
	City        *string   `gorm:"column:city;size:64"`
	UserAgent   *string   `gorm:"column:user_agent;size:256"`
	Browser     *string   `gorm:"column:browser;size:64"`
	OS          *string   `gorm:"column:os;size:64"`
	Device      *string   `gorm:"column:device;size:64"`
	Msg         string    `gorm:"column:msg;size:256"`
	LoginTime   time.Time `gorm:"column:login_time;index"`
	CreatedTime time.Time `gorm:"column:created_time;autoCreateTime;index"`
}

type OperaLog struct {
	ID          int            `gorm:"column:id;primaryKey"`
	TraceID     string         `gorm:"column:trace_id;size:64;index"`
	Username    *string        `gorm:"column:username;size:64;index"`
	Method      string         `gorm:"column:method;size:16"`
	Title       string         `gorm:"column:title;size:128"`
	Path        string         `gorm:"column:path;size:256"`
	IP          string         `gorm:"column:ip;size:64;index"`
	Country     *string        `gorm:"column:country;size:64"`
	Region      *string        `gorm:"column:region;size:64"`
	City        *string        `gorm:"column:city;size:64"`
	UserAgent   *string        `gorm:"column:user_agent;size:256"`
	Browser     *string        `gorm:"column:browser;size:64"`
	OS          *string        `gorm:"column:os;size:64"`
	Device      *string        `gorm:"column:device;size:64"`
	Args        map[string]any `gorm:"column:args;serializer:json;type:json"`
	Status      int            `gorm:"column:status;index"`
	Code        string         `gorm:"column:code;size:32"`
	Msg         *string        `gorm:"column:msg;size:256"`
	CostTime    float64        `gorm:"column:cost_time"`
	OperaTime   time.Time      `gorm:"column:opera_time;index"`
	CreatedTime time.Time      `gorm:"column:created_time;autoCreateTime;index"`
}

func (LoginLog) TableName() string {
	return "sys_login_log"
}

func (OperaLog) TableName() string {
	return "sys_opera_log"
}
