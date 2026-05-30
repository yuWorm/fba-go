package observability_test

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/observability"
)

func TestReadinessAggregatesProbeResults(t *testing.T) {
	readiness := observability.NewReadiness()
	readiness.Add("db", func(context.Context) error { return nil })
	readiness.Add("redis", func(context.Context) error { return errors.New("down") })

	result := readiness.Check(context.Background())
	if result.Ready {
		t.Fatal("Ready = true, want false")
	}
	if result.Checks["db"].OK != true {
		t.Fatalf("db check = %+v, want ok", result.Checks["db"])
	}
	if result.Checks["redis"].OK != false || result.Checks["redis"].Error != "down" {
		t.Fatalf("redis check = %+v, want down", result.Checks["redis"])
	}
}

func TestRegisterCoreRoutes(t *testing.T) {
	app := fiber.New()
	observability.RegisterCoreRoutes(app, observability.NewReadiness())

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		resp, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatalf("GET %s error = %v", path, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll(%s) error = %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("GET %s body is empty", path)
		}
	}
}
