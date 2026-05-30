package app_test

import (
	"context"
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
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
