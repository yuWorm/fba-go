package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"github.com/yuWorm/fba-plugin-admin/service"
)

func TestAuthenticateReturnsTokenInvalidWhenSessionUserIsMissing(t *testing.T) {
	ctx := context.Background()
	sessionUUID := "orphan-session"
	tokenService := coreauth.NewJWTService(config.AuthOptions{AccessTokenTTL: time.Hour})
	token, err := tokenService.CreateAccessToken(ctx, 999, sessionUUID, nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	repository := repo.NewMemoryRepository(model.Seed{
		Sessions: []model.Session{{
			ID:          999,
			SessionUUID: sessionUUID,
			Username:    "missing",
			Status:      1,
			ExpireTime:  time.Now().Add(time.Hour),
		}},
	})
	authService := service.NewAuthService(repository)

	_, err = authService.Authenticate(ctx, "Bearer "+token.Token)

	var appErr *fbaerrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Authenticate() error = %T, want AppError", err)
	}
	if appErr.HTTPStatus() != http.StatusUnauthorized || appErr.PublicMessage() != "Token 无效" {
		t.Fatalf("Authenticate() error = (%d, %q), want (401, Token 无效)", appErr.HTTPStatus(), appErr.PublicMessage())
	}
}

func TestLoginSkipsCaptchaWhenDynamicConfigDisablesIt(t *testing.T) {
	ctx := context.Background()
	repository := repo.NewMemoryRepository(repo.SeedData())
	authService := service.NewAuthServiceWithOptions(repository, service.AuthServiceOptions{
		ConfigProvider: service.StaticAdminConfigProvider{
			Login: service.LoginConfig{CaptchaEnabled: false},
		},
	})

	_, _, err := authService.Login(ctx, dto.AuthLoginParam{
		Username: "admin",
		Password: "admin",
		UUID:     "missing-captcha",
		Captcha:  "wrong",
	}, service.RequestMetadata{})

	if err != nil {
		t.Fatalf("Login() error = %v, want nil when captcha is disabled", err)
	}
}

func TestLoginCaptchaAndLockUseRedisWhenAvailable(t *testing.T) {
	ctx := context.Background()
	redisClient := newFakeAdminRedis()
	repository := repo.NewMemoryRepository(repo.SeedData())
	authService := service.NewAuthServiceWithOptions(repository, service.AuthServiceOptions{
		Redis: redisClient,
		ConfigProvider: service.StaticAdminConfigProvider{
			Login: service.LoginConfig{CaptchaEnabled: true, CaptchaExpire: 5 * time.Minute},
			UserSecurity: service.UserSecurityConfig{
				LockThreshold:     2,
				LockDuration:      5 * time.Minute,
				PasswordExpiry:    365,
				PasswordReminder:  7,
				HistoryCheckCount: 3,
				MinLength:         6,
				MaxLength:         32,
			},
		},
	})
	captcha, err := authService.Captcha(ctx)
	if err != nil {
		t.Fatalf("Captcha() error = %v", err)
	}
	code, ok := redisClient.value("fba:login:captcha:" + captcha.UUID)
	if !ok || code != "1234" {
		t.Fatalf("redis captcha = %q ok %v, want 1234 true", code, ok)
	}

	for i := 0; i < 2; i++ {
		_, _, err = authService.Login(ctx, dto.AuthLoginParam{
			Username: "admin",
			Password: "bad-password",
			UUID:     "fixture-captcha",
			Captcha:  "1234",
		}, service.RequestMetadata{})
	}
	if err == nil || !strings.Contains(err.Error(), "登录失败次数过多") {
		t.Fatalf("Login() error = %v, want redis-backed lock error", err)
	}
	if _, ok := redisClient.value("fba:user:lock:1"); !ok {
		t.Fatal("redis user lock key missing after threshold is reached")
	}
}

func TestPasswordPolicyUsesDynamicConfigProvider(t *testing.T) {
	ctx := context.Background()
	repository := repo.NewMemoryRepository(repo.SeedData())
	userService := service.NewUserServiceWithOptions(repository, service.UserServiceOptions{
		ConfigProvider: service.StaticAdminConfigProvider{
			UserSecurity: service.UserSecurityConfig{
				LockThreshold:      5,
				LockDuration:       5 * time.Minute,
				PasswordExpiry:     365,
				PasswordReminder:   7,
				HistoryCheckCount:  3,
				MinLength:          6,
				MaxLength:          32,
				RequireSpecialChar: true,
			},
		},
	})

	err := userService.ResetPassword(ctx, 1, "Passw0rd")

	if err == nil || !strings.Contains(err.Error(), "密码必须包含特殊字符") {
		t.Fatalf("ResetPassword() error = %v, want special-char policy error", err)
	}
}

func TestDeleteUserRevokesSessionsLikePython(t *testing.T) {
	ctx := context.Background()
	seed := repo.SeedData()
	seed.Users = append(seed.Users, model.User{
		ID:       42,
		UUID:     "delete-session-user",
		Username: "delete_session_user",
		Nickname: "Delete Session User",
		Status:   1,
		IsStaff:  true,
	})
	seed.Sessions = append(seed.Sessions, model.Session{
		ID:          42,
		SessionUUID: "stale-session",
		Username:    "delete_session_user",
		Status:      1,
		ExpireTime:  time.Now().Add(time.Hour),
	})
	repository := repo.NewMemoryRepository(seed)
	userService := service.NewUserService(repository)

	if err := userService.Delete(ctx, 42); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	sessions, err := repository.ListSessions(ctx, repo.SessionFilter{Username: "delete_session_user"})
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions after user delete = %+v, want empty", sessions)
	}
}

func TestUpdateEmailUsesRedisCaptchaWhenAvailable(t *testing.T) {
	ctx := context.Background()
	redisClient := newFakeAdminRedis()
	redisClient.setValue("fba:email:captcha:127.0.0.1", "654321")
	repository := repo.NewMemoryRepository(repo.SeedData())
	userService := service.NewUserServiceWithOptions(repository, service.UserServiceOptions{Redis: redisClient})
	email := "redis-email@example.com"

	if err := userService.UpdateEmail(ctx, 1, "654321", &email); err != nil {
		t.Fatalf("UpdateEmail() error = %v", err)
	}
	if _, ok := redisClient.value("fba:email:captcha:127.0.0.1"); ok {
		t.Fatal("email captcha redis key still exists after successful update")
	}
}

func TestRedisMonitorReadsInfoFromProvider(t *testing.T) {
	ctx := context.Background()
	redisClient := newFakeAdminRedis()
	redisClient.info = strings.Join([]string{
		"# Server",
		"redis_version:7.2.4",
		"redis_mode:standalone",
		"tcp_port:6379",
		"uptime_in_seconds:61",
		"# Clients",
		"connected_clients:3",
		"blocked_clients:1",
		"# Memory",
		"used_memory_human:1.00M",
		"used_memory_rss_human:2.00M",
		"maxmemory_human:0B",
		"mem_fragmentation_ratio:1.23",
		"# Stats",
		"instantaneous_ops_per_sec:4",
		"total_commands_processed:99",
		"rejected_connections:0",
		"cmdstat_get:calls=12,usec=100,usec_per_call=8.33,rejected_calls=0,failed_calls=0",
		"# Keyspace",
		"db0:keys=7,expires=1,avg_ttl=10",
	}, "\r\n")
	monitorService := service.NewMonitorServiceWithRedis(repo.NewMemoryRepository(repo.SeedData()), redisClient)

	monitor, err := monitorService.Redis(ctx)

	if err != nil {
		t.Fatalf("Redis() error = %v", err)
	}
	if monitor.Info.RedisVersion != "7.2.4" || monitor.Info.ConnectedClients != "3" || monitor.Info.KeysNum != "7" {
		t.Fatalf("redis monitor info = %+v, want parsed INFO fields", monitor.Info)
	}
	if len(monitor.Stats) != 1 || monitor.Stats[0].Name != "get" || monitor.Stats[0].Value != "12" {
		t.Fatalf("redis monitor stats = %+v, want get calls=12", monitor.Stats)
	}
}

type fakeAdminRedis struct {
	mu     sync.Mutex
	values map[string]string
	info   string
}

func newFakeAdminRedis() *fakeAdminRedis {
	return &fakeAdminRedis{values: map[string]string{}}
}

func (r *fakeAdminRedis) Get(_ context.Context, key string) *redis.StringCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (r *fakeAdminRedis) Set(_ context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = valueString(value)
	return redis.NewStatusResult("OK", nil)
}

func (r *fakeAdminRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	var deleted int64
	for _, key := range keys {
		if _, ok := r.values[key]; ok {
			deleted++
			delete(r.values, key)
		}
	}
	return redis.NewIntResult(deleted, nil)
}

func (r *fakeAdminRedis) Incr(_ context.Context, key string) *redis.IntCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	var current int64
	if raw := r.values[key]; raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &current)
	}
	current++
	r.values[key] = fmt.Sprintf("%d", current)
	return redis.NewIntResult(current, nil)
}

func (r *fakeAdminRedis) Expire(context.Context, string, time.Duration) *redis.BoolCmd {
	return redis.NewBoolResult(true, nil)
}

func (r *fakeAdminRedis) Info(context.Context, ...string) *redis.StringCmd {
	return redis.NewStringResult(r.info, nil)
}

func (r *fakeAdminRedis) value(key string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.values[key]
	return value, ok
}

func (r *fakeAdminRedis) setValue(key string, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
}

func valueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}
