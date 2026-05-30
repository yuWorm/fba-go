package app_test

import (
	"context"
	"strings"
	"testing"

	fba "github.com/yuWorm/fba-go"
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
