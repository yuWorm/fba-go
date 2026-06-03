package service

import (
	"context"
	"time"
)

type LoginConfig struct {
	CaptchaEnabled bool
	CaptchaExpire  time.Duration
}

type UserSecurityConfig struct {
	LockThreshold      int
	LockDuration       time.Duration
	PasswordExpiry     int
	PasswordReminder   int
	HistoryCheckCount  int
	MinLength          int
	MaxLength          int
	RequireSpecialChar bool
}

type AdminConfigProvider interface {
	LoginConfig(context.Context) (LoginConfig, error)
	UserSecurityConfig(context.Context) (UserSecurityConfig, error)
}

type DefaultAdminConfigProvider struct{}

func (DefaultAdminConfigProvider) LoginConfig(context.Context) (LoginConfig, error) {
	return defaultLoginConfig(), nil
}

func (DefaultAdminConfigProvider) UserSecurityConfig(context.Context) (UserSecurityConfig, error) {
	return defaultUserSecurityConfig(), nil
}

type StaticAdminConfigProvider struct {
	Login        LoginConfig
	UserSecurity UserSecurityConfig
}

func (p StaticAdminConfigProvider) LoginConfig(context.Context) (LoginConfig, error) {
	cfg := p.Login
	if cfg.CaptchaExpire <= 0 {
		cfg.CaptchaExpire = defaultLoginConfig().CaptchaExpire
	}
	return cfg, nil
}

func (p StaticAdminConfigProvider) UserSecurityConfig(context.Context) (UserSecurityConfig, error) {
	return normalizeUserSecurityConfig(p.UserSecurity), nil
}

func defaultLoginConfig() LoginConfig {
	return LoginConfig{
		CaptchaEnabled: true,
		CaptchaExpire:  5 * time.Minute,
	}
}

func defaultUserSecurityConfig() UserSecurityConfig {
	return UserSecurityConfig{
		LockThreshold:     5,
		LockDuration:      5 * time.Minute,
		PasswordExpiry:    365,
		PasswordReminder:  7,
		HistoryCheckCount: 3,
		MinLength:         6,
		MaxLength:         32,
	}
}

func normalizeUserSecurityConfig(cfg UserSecurityConfig) UserSecurityConfig {
	defaults := defaultUserSecurityConfig()
	if cfg.LockThreshold < 0 {
		cfg.LockThreshold = defaults.LockThreshold
	}
	if cfg.LockDuration <= 0 && cfg.LockThreshold != 0 {
		cfg.LockDuration = defaults.LockDuration
	}
	if cfg.PasswordExpiry < 0 {
		cfg.PasswordExpiry = defaults.PasswordExpiry
	}
	if cfg.PasswordReminder < 0 {
		cfg.PasswordReminder = defaults.PasswordReminder
	}
	if cfg.HistoryCheckCount <= 0 {
		cfg.HistoryCheckCount = defaults.HistoryCheckCount
	}
	if cfg.MinLength <= 0 {
		cfg.MinLength = defaults.MinLength
	}
	if cfg.MaxLength <= 0 {
		cfg.MaxLength = defaults.MaxLength
	}
	return cfg
}

func adminConfigProvider(provider AdminConfigProvider) AdminConfigProvider {
	if provider == nil {
		return DefaultAdminConfigProvider{}
	}
	return provider
}
