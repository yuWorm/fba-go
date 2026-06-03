package model

import "time"

type Config struct {
	ID          int        `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:32"`
	Type        *string    `gorm:"column:type;size:32;index"`
	Key         string     `gorm:"column:key;size:64;uniqueIndex:uk_sys_config_key_deleted"`
	Value       string     `gorm:"column:value;type:text"`
	IsFrontend  bool       `gorm:"column:is_frontend"`
	Remark      *string    `gorm:"column:remark;type:text"`
	Deleted     int        `gorm:"column:deleted;uniqueIndex:uk_sys_config_key_deleted"`
	CreatedTime time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime *time.Time `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime *time.Time `gorm:"column:deleted_time;index"`
}

func (Config) TableName() string {
	return "sys_config"
}

func SeedConfigs() []Config {
	created := seedTime()
	ptr := func(value string) *string {
		return &value
	}
	return []Config{
		{
			ID:          1,
			Name:        "邮箱配置状态",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_CONFIG_STATUS",
			Value:       "1",
			IsFrontend:  false,
			Remark:      ptr("邮箱配置状态：1 启用，0 停用"),
			CreatedTime: created,
		},
		{
			ID:          2,
			Name:        "邮件服务器",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_HOST",
			Value:       "smtp.qq.com",
			IsFrontend:  false,
			Remark:      ptr("SMTP 服务器地址"),
			CreatedTime: created,
		},
		{
			ID:          3,
			Name:        "邮件端口",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_PORT",
			Value:       "465",
			IsFrontend:  false,
			Remark:      ptr("SMTP 服务器端口"),
			CreatedTime: created,
		},
		{
			ID:          4,
			Name:        "邮箱账号",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_USERNAME",
			Value:       "fba@qq.com",
			IsFrontend:  false,
			Remark:      ptr("SMTP 登录账号"),
			CreatedTime: created,
		},
		{
			ID:          5,
			Name:        "邮箱密码",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_PASSWORD",
			Value:       "",
			IsFrontend:  false,
			Remark:      ptr("SMTP 登录密码或授权码"),
			CreatedTime: created,
		},
		{
			ID:          6,
			Name:        "SSL 开关",
			Type:        ptr("EMAIL"),
			Key:         "EMAIL_SSL",
			Value:       "true",
			IsFrontend:  false,
			Remark:      ptr("是否启用 SSL"),
			CreatedTime: created,
		},
		{
			ID:          7,
			Name:        "用户安全配置状态",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_SECURITY_CONFIG_STATUS",
			Value:       "1",
			IsFrontend:  false,
			Remark:      ptr("用户安全配置状态：1 启用，0 停用"),
			CreatedTime: created,
		},
		{
			ID:          8,
			Name:        "用户锁定阈值",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_LOCK_THRESHOLD",
			Value:       "5",
			IsFrontend:  false,
			Remark:      ptr("连续登录失败锁定阈值"),
			CreatedTime: created,
		},
		{
			ID:          9,
			Name:        "用户锁定秒数",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_LOCK_SECONDS",
			Value:       "300",
			IsFrontend:  false,
			Remark:      ptr("用户锁定时长，单位秒"),
			CreatedTime: created,
		},
		{
			ID:          10,
			Name:        "密码过期天数",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_EXPIRY_DAYS",
			Value:       "365",
			IsFrontend:  false,
			Remark:      ptr("密码过期天数"),
			CreatedTime: created,
		},
		{
			ID:          11,
			Name:        "密码到期提醒天数",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_REMINDER_DAYS",
			Value:       "7",
			IsFrontend:  false,
			Remark:      ptr("密码到期前提醒天数"),
			CreatedTime: created,
		},
		{
			ID:          12,
			Name:        "历史密码检查次数",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_HISTORY_CHECK_COUNT",
			Value:       "3",
			IsFrontend:  false,
			Remark:      ptr("禁止复用最近 N 次历史密码"),
			CreatedTime: created,
		},
		{
			ID:          13,
			Name:        "密码最小长度",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_MIN_LENGTH",
			Value:       "6",
			IsFrontend:  false,
			Remark:      ptr("密码最小长度"),
			CreatedTime: created,
		},
		{
			ID:          14,
			Name:        "密码最大长度",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_MAX_LENGTH",
			Value:       "32",
			IsFrontend:  false,
			Remark:      ptr("密码最大长度"),
			CreatedTime: created,
		},
		{
			ID:          15,
			Name:        "密码特殊字符要求",
			Type:        ptr("USER_SECURITY"),
			Key:         "USER_PASSWORD_REQUIRE_SPECIAL_CHAR",
			Value:       "false",
			IsFrontend:  false,
			Remark:      ptr("是否要求密码包含特殊字符"),
			CreatedTime: created,
		},
		{
			ID:          16,
			Name:        "登录配置状态",
			Type:        ptr("LOGIN"),
			Key:         "LOGIN_CONFIG_STATUS",
			Value:       "1",
			IsFrontend:  false,
			Remark:      ptr("登录配置状态：1 启用，0 停用"),
			CreatedTime: created,
		},
		{
			ID:          17,
			Name:        "验证码开关",
			Type:        ptr("LOGIN"),
			Key:         "LOGIN_CAPTCHA_ENABLED",
			Value:       "true",
			IsFrontend:  false,
			Remark:      ptr("是否启用登录验证码"),
			CreatedTime: created,
		},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
