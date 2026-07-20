package config_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yuWorm/fba-go/core/config"
)

func TestCORSDefaultsMatchPythonSettings(t *testing.T) {
	opts := config.Options{}.WithDefaults()

	if !opts.CORS.Enabled {
		t.Fatal("CORS.Enabled = false, want true")
	}
	if !opts.CORS.AllowCredentials {
		t.Fatal("CORS.AllowCredentials = false, want true")
	}
	if !reflect.DeepEqual(opts.CORS.AllowedOrigins, []string{"http://127.0.0.1", "http://localhost:5173"}) {
		t.Fatalf("CORS.AllowedOrigins = %#v", opts.CORS.AllowedOrigins)
	}
	if !reflect.DeepEqual(opts.CORS.AllowMethods, []string{"*"}) {
		t.Fatalf("CORS.AllowMethods = %#v", opts.CORS.AllowMethods)
	}
	if !reflect.DeepEqual(opts.CORS.AllowHeaders, []string{"*"}) {
		t.Fatalf("CORS.AllowHeaders = %#v", opts.CORS.AllowHeaders)
	}
	if !reflect.DeepEqual(opts.CORS.ExposeHeaders, []string{"X-Request-ID"}) {
		t.Fatalf("CORS.ExposeHeaders = %#v", opts.CORS.ExposeHeaders)
	}
}

func TestValidateCORSOptionsRejectsWildcardCredentials(t *testing.T) {
	opts := config.Options{
		CORS: config.CORSOptions{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		},
	}.WithDefaults()

	if err := config.ValidateCORSOptions(opts.CORS); !errors.Is(err, config.ErrCORSWildcardCredentials) {
		t.Fatalf("ValidateCORSOptions() error = %v, want ErrCORSWildcardCredentials", err)
	}
}
