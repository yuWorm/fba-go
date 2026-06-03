package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	"github.com/yuWorm/fba-plugin-config/model"
	"github.com/yuWorm/fba-plugin-config/repo"
)

type AdminConfigProvider struct {
	repo repo.Repository
}

func NewAdminConfigProvider(repository repo.Repository) AdminConfigProvider {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return AdminConfigProvider{repo: repository}
}

func (p AdminConfigProvider) LoginConfig(ctx context.Context) (adminservice.LoginConfig, error) {
	cfg, err := (adminservice.DefaultAdminConfigProvider{}).LoginConfig(ctx)
	if err != nil {
		return cfg, err
	}
	values, err := p.valuesByType(ctx, "LOGIN")
	if err != nil {
		return cfg, err
	}
	// Python dynamic_config.py treats *_CONFIG_STATUS=0 as "plugin config disabled"
	// and then falls back to the static admin defaults.
	if values["LOGIN_CONFIG_STATUS"] == "0" {
		return cfg, nil
	}
	if raw, ok := values["LOGIN_CAPTCHA_ENABLED"]; ok {
		parsed, err := parseBool("LOGIN_CAPTCHA_ENABLED", raw)
		if err != nil {
			return cfg, err
		}
		cfg.CaptchaEnabled = parsed
	}
	return cfg, nil
}

func (p AdminConfigProvider) UserSecurityConfig(ctx context.Context) (adminservice.UserSecurityConfig, error) {
	cfg, err := (adminservice.DefaultAdminConfigProvider{}).UserSecurityConfig(ctx)
	if err != nil {
		return cfg, err
	}
	values, err := p.valuesByType(ctx, "USER_SECURITY")
	if err != nil {
		return cfg, err
	}
	if values["USER_SECURITY_CONFIG_STATUS"] == "0" {
		return cfg, nil
	}

	if err := setInt(values, "USER_LOCK_THRESHOLD", &cfg.LockThreshold); err != nil {
		return cfg, err
	}
	if raw, ok := values["USER_LOCK_SECONDS"]; ok {
		seconds, err := parseInt("USER_LOCK_SECONDS", raw)
		if err != nil {
			return cfg, err
		}
		cfg.LockDuration = time.Duration(seconds) * time.Second
	}
	if err := setInt(values, "USER_PASSWORD_EXPIRY_DAYS", &cfg.PasswordExpiry); err != nil {
		return cfg, err
	}
	if err := setInt(values, "USER_PASSWORD_REMINDER_DAYS", &cfg.PasswordReminder); err != nil {
		return cfg, err
	}
	if err := setInt(values, "USER_PASSWORD_HISTORY_CHECK_COUNT", &cfg.HistoryCheckCount); err != nil {
		return cfg, err
	}
	if err := setInt(values, "USER_PASSWORD_MIN_LENGTH", &cfg.MinLength); err != nil {
		return cfg, err
	}
	if err := setInt(values, "USER_PASSWORD_MAX_LENGTH", &cfg.MaxLength); err != nil {
		return cfg, err
	}
	if raw, ok := values["USER_PASSWORD_REQUIRE_SPECIAL_CHAR"]; ok {
		parsed, err := parseBool("USER_PASSWORD_REQUIRE_SPECIAL_CHAR", raw)
		if err != nil {
			return cfg, err
		}
		cfg.RequireSpecialChar = parsed
	}
	return cfg, nil
}

func (p AdminConfigProvider) valuesByType(ctx context.Context, typeName string) (map[string]string, error) {
	items, err := p.repo.All(ctx, typeName)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		if isType(item, typeName) {
			values[item.Key] = item.Value
		}
	}
	return values, nil
}

func isType(item model.Config, typeName string) bool {
	return item.Type != nil && *item.Type == typeName
}

func setInt(values map[string]string, key string, target *int) error {
	raw, ok := values[key]
	if !ok {
		return nil
	}
	parsed, err := parseInt(key, raw)
	if err != nil {
		return err
	}
	*target = parsed
	return nil
}

func parseInt(key string, raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse config %s as int: %w", key, err)
	}
	return value, nil
}

func parseBool(key string, raw string) (bool, error) {
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse config %s as bool: %w", key, err)
	}
	return value, nil
}
