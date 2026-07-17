package logger

import (
	"fmt"
	"time"

	"github.com/yuWorm/fba-go/core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a logger that defaults terminal sinks to colored console output
// and every non-terminal sink to JSON. Encoding explicitly forces one format.
func New(opts config.LoggerOptions) (*zap.Logger, error) {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	if opts.Encoding != "" {
		return newUniformLogger(opts, level)
	}
	return newSplitLogger(opts, level)
}

func newUniformLogger(opts config.LoggerOptions, level zapcore.Level) (*zap.Logger, error) {
	if opts.Encoding != "json" && opts.Encoding != "console" {
		return nil, fmt.Errorf("invalid logger encoding %q", opts.Encoding)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.Encoding = opts.Encoding
	if opts.Encoding == "console" {
		cfg.EncoderConfig = consoleEncoderConfig()
	}
	if len(opts.OutputPaths) > 0 {
		cfg.OutputPaths = opts.OutputPaths
	}
	if len(opts.ErrorOutputPaths) > 0 {
		cfg.ErrorOutputPaths = opts.ErrorOutputPaths
	}
	return cfg.Build()
}

func newSplitLogger(opts config.LoggerOptions, level zapcore.Level) (*zap.Logger, error) {
	outputPaths := opts.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stderr"}
	}
	consolePaths, filePaths := splitOutputPaths(outputPaths)

	cores := make([]zapcore.Core, 0, 2)
	closeOutputs := make([]func(), 0, 2)
	addCore := func(paths []string, encoder zapcore.Encoder) error {
		if len(paths) == 0 {
			return nil
		}
		sink, closeSink, err := zap.Open(paths...)
		if err != nil {
			return err
		}
		cores = append(cores, zapcore.NewCore(encoder, sink, level))
		closeOutputs = append(closeOutputs, closeSink)
		return nil
	}
	closeOnError := func() {
		for _, closeOutput := range closeOutputs {
			closeOutput()
		}
	}

	if err := addCore(consolePaths, zapcore.NewConsoleEncoder(consoleEncoderConfig())); err != nil {
		closeOnError()
		return nil, err
	}
	if err := addCore(filePaths, zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())); err != nil {
		closeOnError()
		return nil, err
	}

	errorOutputPaths := opts.ErrorOutputPaths
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}
	errorOutput, _, err := zap.Open(errorOutputPaths...)
	if err != nil {
		closeOnError()
		return nil, err
	}

	core := zapcore.NewTee(cores...)
	core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	return zap.New(
		core,
		zap.ErrorOutput(errorOutput),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	), nil
}

func splitOutputPaths(paths []string) (console []string, files []string) {
	for _, path := range paths {
		if path == "stdout" || path == "stderr" {
			console = append(console, path)
			continue
		}
		files = append(files, path)
	}
	return console, files
}

func consoleEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg
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
