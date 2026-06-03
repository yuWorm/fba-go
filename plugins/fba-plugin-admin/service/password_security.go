package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

var passwordHasher = coreauth.NewPasswordService(0)

func hashPassword(password string) (string, error) {
	return passwordHasher.Hash(password)
}

func validateNewPassword(ctx context.Context, repository repo.Repository, userID int, newPassword string, cfg UserSecurityConfig) error {
	cfg = normalizeUserSecurityConfig(cfg)
	if len(newPassword) < cfg.MinLength {
		return userBadRequest("密码长度不能少于 "+itoaConfig(cfg.MinLength)+" 个字符", nil)
	}
	if len(newPassword) > cfg.MaxLength {
		return userBadRequest("密码长度不能超过 "+itoaConfig(cfg.MaxLength)+" 个字符", nil)
	}
	if !hasASCIIDigit(newPassword) {
		return userBadRequest("密码必须包含数字", nil)
	}
	if !hasASCIILetter(newPassword) {
		return userBadRequest("密码必须包含字母", nil)
	}
	if cfg.RequireSpecialChar && !hasSpecialChar(newPassword) {
		return userBadRequest("密码必须包含特殊字符", nil)
	}
	histories, err := repository.ListUserPasswordHistories(ctx, userID, cfg.HistoryCheckCount)
	if err != nil {
		return err
	}
	for _, history := range histories {
		if passwordMatchesStored(history.Password, newPassword) {
			return userBadRequest("新密码不能与最近 "+itoaConfig(cfg.HistoryCheckCount)+" 次使用的密码相同", nil)
		}
	}
	return nil
}

func passwordExpiryDaysRemaining(changedAt *time.Time, cfg UserSecurityConfig) (*int, error) {
	cfg = normalizeUserSecurityConfig(cfg)
	if cfg.PasswordExpiry == 0 {
		return nil, nil
	}
	if changedAt == nil {
		return nil, authError("密码已过期，请修改密码后重新登录")
	}
	expiryTime := changedAt.Add(time.Duration(cfg.PasswordExpiry) * 24 * time.Hour)
	remaining := expiryTime.Sub(time.Now())
	if remaining < 0 {
		return nil, authError("密码已过期，请修改密码后重新登录")
	}
	days := int(remaining / (24 * time.Hour))
	if cfg.PasswordReminder > 0 && days <= cfg.PasswordReminder {
		return &days, nil
	}
	return nil, nil
}

func passwordMatchesStored(stored string, plain string) bool {
	// The seeded admin user intentionally keeps an empty password for fixture
	// compatibility while Python treats the initial login password as "admin".
	if stored == "" {
		return plain == "" || plain == "admin"
	}
	if looksLikeBcryptHash(stored) {
		return passwordHasher.Verify(stored, plain)
	}
	// Keep compatibility with rows created before the Go migration adopted
	// Python-style hashing, so existing deployments can still authenticate.
	return stored == plain
}

func looksLikeBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") ||
		strings.HasPrefix(value, "$2b$") ||
		strings.HasPrefix(value, "$2x$") ||
		strings.HasPrefix(value, "$2y$")
}

func hasASCIIDigit(value string) bool {
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			return true
		}
	}
	return false
}

func hasASCIILetter(value string) bool {
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			return true
		}
	}
	return false
}

func hasSpecialChar(value string) bool {
	for _, ch := range value {
		if !(ch >= '0' && ch <= '9') &&
			!(ch >= 'a' && ch <= 'z') &&
			!(ch >= 'A' && ch <= 'Z') {
			return true
		}
	}
	return false
}

func itoaConfig(value int) string {
	return strconv.Itoa(value)
}
