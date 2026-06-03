package service

import (
	"context"
	"time"

	"github.com/yuWorm/fba-plugin-admin/repo"
)

const (
	userPasswordMinLength         = 6
	userPasswordMaxLength         = 32
	userPasswordHistoryCheckCount = 3
	userPasswordExpiryDays        = 365
	userPasswordReminderDays      = 7
)

func validateNewPassword(ctx context.Context, repository repo.Repository, userID int, newPassword string) error {
	if len(newPassword) < userPasswordMinLength {
		return userBadRequest("密码长度不能少于 6 个字符", nil)
	}
	if len(newPassword) > userPasswordMaxLength {
		return userBadRequest("密码长度不能超过 32 个字符", nil)
	}
	if !hasASCIIDigit(newPassword) {
		return userBadRequest("密码必须包含数字", nil)
	}
	if !hasASCIILetter(newPassword) {
		return userBadRequest("密码必须包含字母", nil)
	}
	histories, err := repository.ListUserPasswordHistories(ctx, userID, userPasswordHistoryCheckCount)
	if err != nil {
		return err
	}
	for _, history := range histories {
		if passwordMatchesStored(history.Password, newPassword) {
			return userBadRequest("新密码不能与最近 3 次使用的密码相同", nil)
		}
	}
	return nil
}

func passwordExpiryDaysRemaining(changedAt *time.Time) (*int, error) {
	if userPasswordExpiryDays == 0 {
		return nil, nil
	}
	if changedAt == nil {
		return nil, authError("密码已过期，请修改密码后重新登录")
	}
	expiryTime := changedAt.Add(time.Duration(userPasswordExpiryDays) * 24 * time.Hour)
	remaining := expiryTime.Sub(time.Now())
	if remaining < 0 {
		return nil, authError("密码已过期，请修改密码后重新登录")
	}
	days := int(remaining / (24 * time.Hour))
	if days <= userPasswordReminderDays {
		return &days, nil
	}
	return nil, nil
}

func passwordMatchesStored(stored string, plain string) bool {
	if stored != "" {
		return stored == plain
	}
	// The seeded admin user intentionally keeps an empty password for fixture
	// compatibility while Python treats the initial login password as "admin".
	return plain == "" || plain == "admin"
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
