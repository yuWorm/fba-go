package logger

import (
	"fmt"

	"github.com/yuWorm/fba-go/core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(opts config.LoggerOptions) (*zap.Logger, error) {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	encoding := opts.Encoding
	if encoding == "" {
		encoding = "json"
	}
	if encoding != "json" && encoding != "console" {
		return nil, fmt.Errorf("invalid logger encoding %q", encoding)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.Encoding = encoding
	if len(opts.OutputPaths) > 0 {
		cfg.OutputPaths = opts.OutputPaths
	}
	if len(opts.ErrorOutputPaths) > 0 {
		cfg.ErrorOutputPaths = opts.ErrorOutputPaths
	}

	return cfg.Build()
}

func parseLevel(value string) (zapcore.Level, error) {
	if value == "" {
		value = "info"
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("invalid logger level %q: %w", value, err)
	}
	return level, nil
}
