package logger_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/logger"
	"go.uber.org/zap/zapcore"
)

func TestNewUsesInfoJSONDefaults(t *testing.T) {
	log, err := logger.New(config.LoggerOptions{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = log.Sync() }()

	if !log.Core().Enabled(zapcore.InfoLevel) {
		t.Fatal("info level disabled, want enabled")
	}
	if log.Core().Enabled(zapcore.DebugLevel) {
		t.Fatal("debug level enabled, want disabled by default")
	}
}

func TestNewAllowsDebugLevel(t *testing.T) {
	log, err := logger.New(config.LoggerOptions{Level: "debug"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = log.Sync() }()

	if !log.Core().Enabled(zapcore.DebugLevel) {
		t.Fatal("debug level disabled, want enabled")
	}
}

func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := logger.New(config.LoggerOptions{Level: "verbose"})
	if err == nil {
		t.Fatal("New() error = nil, want invalid level error")
	}
}
