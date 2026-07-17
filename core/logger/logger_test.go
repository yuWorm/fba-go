package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewUsesInfoLevelByDefault(t *testing.T) {
	log, err := New(config.LoggerOptions{})
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

func TestNewDefaultsFileOutputToJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	log, err := New(config.LoggerOptions{OutputPaths: []string{path}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("server started", zap.String("component", "api"))
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var entry map[string]any
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("file log is not JSON: %v\n%s", err, content)
	}
	if entry["msg"] != "server started" || entry["component"] != "api" {
		t.Fatalf("file log = %#v, want structured fields", entry)
	}
}

func TestConsoleEncoderUsesColoredHumanReadableOutput(t *testing.T) {
	encoder := zapcore.NewConsoleEncoder(consoleEncoderConfig())
	buffer, err := encoder.EncodeEntry(zapcore.Entry{
		Time:    time.Date(2026, time.July, 18, 0, 49, 0, 0, time.UTC),
		Level:   zapcore.InfoLevel,
		Message: "server started",
	}, []zapcore.Field{zap.String("component", "api")})
	if err != nil {
		t.Fatalf("EncodeEntry() error = %v", err)
	}
	defer buffer.Free()

	output := buffer.String()
	if !strings.Contains(output, "2026-07-18T00:49:00.000Z") {
		t.Fatalf("console log = %q, missing human-readable timestamp", output)
	}
	if !strings.Contains(output, "\x1b[") || !strings.Contains(output, "INFO") {
		t.Fatalf("console log = %q, missing colored level", output)
	}
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Fatalf("console log = %q, want console text instead of JSON", output)
	}
}

func TestSplitOutputPathsSeparatesConsoleSinks(t *testing.T) {
	console, files := splitOutputPaths([]string{"stdout", "logs/app.log", "stderr"})
	if want := []string{"stdout", "stderr"}; !slices.Equal(console, want) {
		t.Fatalf("console paths = %v, want %v", console, want)
	}
	if want := []string{"logs/app.log"}; !slices.Equal(files, want) {
		t.Fatalf("file paths = %v, want %v", files, want)
	}
}

func TestNewAllowsDebugLevel(t *testing.T) {
	log, err := New(config.LoggerOptions{Level: "debug"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = log.Sync() }()

	if !log.Core().Enabled(zapcore.DebugLevel) {
		t.Fatal("debug level disabled, want enabled")
	}
}

func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := New(config.LoggerOptions{Level: "verbose"})
	if err == nil {
		t.Fatal("New() error = nil, want invalid level error")
	}
}
