package middleware

import (
	"slices"
	"testing"

	"github.com/yuWorm/fba-go/core/config"
)

func TestFileLogPathsAlsoMirrorToConsole(t *testing.T) {
	accessPaths := accessLogPaths(config.LoggerOptions{AccessLogPath: "logs/access.log"})
	if want := []string{"logs/access.log", "stdout"}; !slices.Equal(accessPaths, want) {
		t.Fatalf("accessLogPaths() = %v, want %v", accessPaths, want)
	}

	errorPaths := errorLogPaths(config.LoggerOptions{ErrorLogPath: "logs/error.log"})
	if want := []string{"logs/error.log", "stderr"}; !slices.Equal(errorPaths, want) {
		t.Fatalf("errorLogPaths() = %v, want %v", errorPaths, want)
	}
}

func TestFileLogPathsDoNotDuplicateConsoleSink(t *testing.T) {
	accessPaths := accessLogPaths(config.LoggerOptions{AccessLogPath: "stdout"})
	if want := []string{"stdout"}; !slices.Equal(accessPaths, want) {
		t.Fatalf("accessLogPaths() = %v, want %v", accessPaths, want)
	}

	errorPaths := errorLogPaths(config.LoggerOptions{ErrorLogPath: "stderr"})
	if want := []string{"stderr"}; !slices.Equal(errorPaths, want) {
		t.Fatalf("errorLogPaths() = %v, want %v", errorPaths, want)
	}
}
