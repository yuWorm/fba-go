package config_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/config"
)

func TestWithDefaultsEnablesHTTPDiagnostics(t *testing.T) {
	opts := config.Options{}.WithDefaults()

	if !opts.Middleware.RequestID.Enabled {
		t.Fatal("RequestID.Enabled = false, want true")
	}
	if !opts.Middleware.Recover.Enabled {
		t.Fatal("Recover.Enabled = false, want true")
	}
	if !opts.Middleware.Recover.EnableStackTrace {
		t.Fatal("Recover.EnableStackTrace = false, want true")
	}
	if !opts.Middleware.AccessLog.Enabled {
		t.Fatal("AccessLog.Enabled = false, want true")
	}
	if !opts.Middleware.ErrorLog.Enabled {
		t.Fatal("ErrorLog.Enabled = false, want true")
	}
	if !opts.Middleware.ErrorResponse.IncludeDetail {
		t.Fatal("ErrorResponse.IncludeDetail = false in dev, want true")
	}
}

func TestWithDefaultsHidesErrorDetailOutsideDev(t *testing.T) {
	opts := config.Options{
		App: config.AppOptions{Environment: "prod"},
	}.WithDefaults()

	if opts.Middleware.ErrorResponse.IncludeDetail {
		t.Fatal("ErrorResponse.IncludeDetail = true in prod, want false")
	}
}

func TestWithDefaultsCannotExposeErrorDetailOutsideDev(t *testing.T) {
	opts := config.Options{
		App: config.AppOptions{Environment: "prod"},
		Middleware: config.MiddlewareOptions{
			ErrorResponse: config.ErrorResponseOptions{IncludeDetail: true},
		},
	}.WithDefaults()

	if opts.Middleware.ErrorResponse.IncludeDetail {
		t.Fatal("ErrorResponse.IncludeDetail = true in prod after an explicit enable")
	}
}
