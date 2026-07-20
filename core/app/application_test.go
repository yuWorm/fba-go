package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	fba "github.com/yuWorm/fba-go"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/redisx"
	coretask "github.com/yuWorm/fba-go/core/task"
)

func TestNewApplicationBuildsCoreApp(t *testing.T) {
	app, err := fba.NewApplication(fba.Options{})
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if app.HTTP() == nil {
		t.Fatal("HTTP() returned nil")
	}
	if app.Container() == nil {
		t.Fatal("Container() returned nil")
	}
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestShutdownRunsHooksInReverseOrder(t *testing.T) {
	var calls []string
	app, err := fba.NewApplication(fba.Options{
		Hooks: fba.Hooks{
			OnShutdown: []fba.Hook{
				func(context.Context) error {
					calls = append(calls, "first")
					return nil
				},
				func(context.Context) error {
					calls = append(calls, "second")
					return nil
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	got := strings.Join(calls, ",")
	const want = "second,first"
	if got != want {
		t.Fatalf("shutdown hook order = %q, want %q", got, want)
	}
}

func TestRunHTTPUnwindsResourcesWhenStartHookFails(t *testing.T) {
	var calls []string
	startErr := errors.New("start failed")
	app, err := fba.NewApplication(fba.Options{
		Hooks: fba.Hooks{
			OnStart: []fba.Hook{
				func(context.Context) error {
					calls = append(calls, "start-first")
					return nil
				},
				func(context.Context) error {
					calls = append(calls, "start-second")
					return startErr
				},
			},
			OnShutdown: []fba.Hook{
				func(context.Context) error {
					calls = append(calls, "shutdown-first")
					return nil
				},
				func(context.Context) error {
					calls = append(calls, "shutdown-second")
					return nil
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	err = app.RunHTTP(context.Background())

	if !errors.Is(err, startErr) {
		t.Fatalf("RunHTTP() error = %v, want start error", err)
	}
	got := strings.Join(calls, ",")
	const want = "start-first,start-second,shutdown-second,shutdown-first"
	if got != want {
		t.Fatalf("lifecycle calls = %q, want %q", got, want)
	}
}

func TestShutdownRunsEveryHookOnceWhenOneFails(t *testing.T) {
	var calls []string
	shutdownErr := errors.New("shutdown failed")
	app, err := fba.NewApplication(fba.Options{
		Hooks: fba.Hooks{OnShutdown: []fba.Hook{
			func(context.Context) error {
				calls = append(calls, "first")
				return nil
			},
			func(context.Context) error {
				calls = append(calls, "second")
				return shutdownErr
			},
		}},
	})
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if err := app.Shutdown(context.Background()); !errors.Is(err, shutdownErr) {
		t.Fatalf("Shutdown() error = %v, want hook error", err)
	}
	if err := app.Shutdown(context.Background()); !errors.Is(err, shutdownErr) {
		t.Fatalf("second Shutdown() error = %v, want same hook error", err)
	}
	if got := strings.Join(calls, ","); got != "second,first" {
		t.Fatalf("shutdown calls = %q, want second,first exactly once", got)
	}
}

func TestNewApplicationRejectsWildcardCORSWithCredentials(t *testing.T) {
	_, err := fba.NewApplication(fba.Options{
		CORS: config.CORSOptions{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "wildcard origin") {
		t.Fatalf("NewApplication() error = %v, want wildcard origin error", err)
	}
}

func TestNewApplicationProvidesEnabledTaskRuntime(t *testing.T) {
	app, err := fba.NewApplication(fba.Options{
		Task: config.TaskOptions{
			Enabled:     true,
			RedisAddr:   "127.0.0.1:6379",
			Concurrency: 2,
			Queues:      map[string]int{"critical": 2, "default": 1},
		},
	})
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	defer func() {
		if err := app.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown() error = %v", err)
		}
	}()

	var registry coretask.DefinitionRegistry
	if !app.Container().Resolve(&registry) || registry == nil {
		t.Fatal("task definition registry was not provided")
	}
	var runtime coretask.Runtime
	if !app.Container().Resolve(&runtime) {
		t.Fatal("task runtime was not provided")
	}
	if _, ok := runtime.(*coretask.AsynqRuntime); !ok {
		t.Fatalf("task runtime = %T, want *task.AsynqRuntime", runtime)
	}
	var redisClient coretask.BackendRedisClient
	if !app.Container().Resolve(&redisClient) || redisClient == nil {
		t.Fatal("task Redis client was not provided")
	}
}

func TestRealtimeRedisLifecycleFollowsRealtimeState(t *testing.T) {
	disabledApp, err := fba.NewApplication(fba.Options{
		Realtime: config.RealtimeOptions{
			Disabled:      true,
			MultiInstance: config.RealtimeMultiInstanceOptions{Enabled: true},
		},
	})
	if err != nil {
		t.Fatalf("NewApplication(disabled realtime) error = %v", err)
	}
	var disabledRedis redisx.RedisClient
	if disabledApp.Container().Resolve(&disabledRedis) {
		t.Fatal("disabled realtime provisioned a multi-instance Redis client")
	}
	if err := disabledApp.Shutdown(context.Background()); err != nil {
		t.Fatalf("disabledApp.Shutdown() error = %v", err)
	}

	configuredApp, err := fba.NewApplication(fba.Options{
		Redis:    config.RedisOptions{Addr: "127.0.0.1:6379"},
		Realtime: config.RealtimeOptions{Disabled: true},
	})
	if err != nil {
		t.Fatalf("NewApplication(configured Redis) error = %v", err)
	}
	var configuredRedis redisx.RedisClient
	if !configuredApp.Container().Resolve(&configuredRedis) || configuredRedis == nil {
		t.Fatal("explicit Redis configuration was not provided to plugins")
	}
	if err := configuredApp.Shutdown(context.Background()); err != nil {
		t.Fatalf("configuredApp.Shutdown() error = %v", err)
	}
	if err := configuredRedis.Ping(context.Background()).Err(); err == nil {
		t.Fatal("configured Redis Ping() error = nil after application shutdown")
	}

	enabledApp, err := fba.NewApplication(fba.Options{
		Auth: config.AuthOptions{
			JWTSecret: "0123456789abcdef0123456789abcdef",
			JWTIssuer: "application-test",
		},
		Realtime: config.RealtimeOptions{
			Enabled:       true,
			MultiInstance: config.RealtimeMultiInstanceOptions{Enabled: true},
		},
	})
	if err != nil {
		t.Fatalf("NewApplication(enabled realtime) error = %v", err)
	}
	var enabledRedis redisx.RedisClient
	if !enabledApp.Container().Resolve(&enabledRedis) || enabledRedis == nil {
		t.Fatal("enabled multi-instance realtime did not provide Redis")
	}
	if err := enabledApp.Shutdown(context.Background()); err != nil {
		t.Fatalf("enabledApp.Shutdown() error = %v", err)
	}
	if err := enabledRedis.Ping(context.Background()).Err(); err == nil {
		t.Fatal("realtime Redis Ping() error = nil after application shutdown")
	}
}
